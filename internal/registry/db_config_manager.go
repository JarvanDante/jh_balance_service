package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/hashicorp/consul/api"
)

// SiteConfig 站点配置结构（匹配新的KV格式）
type SiteConfig struct {
	Meta     MetaConfig       `json:"Meta"`
	Base     BaseConfig       `json:"Base"`
	Domain   DomainConfig     `json:"Domain"`
	Game     []GameConfig     `json:"Game"`
	Payment  []PaymentConfig  `json:"Payment"`
	Withdraw []WithdrawConfig `json:"Withdraw"`
}

// MetaConfig 元数据配置
type MetaConfig struct {
	Version     int64  `json:"version"`
	PublishedAt string `json:"published_at"`
	AdminId     int    `json:"admin_id"`
}

// BaseConfig 基础配置
type BaseConfig struct {
	SiteId        int    `json:"site_id"`
	Code          string `json:"code"`
	TimeZone      string `json:"timezone"`
	Currency      string `json:"currency"`
	DbLinkSite    string `json:"db_link_site"`
	DbLinkBalance string `json:"db_link_balance"`
	DbLinkGame    string `json:"db_link_game"`
	Status        int    `json:"status"`
}

// DomainConfig 域名配置
type DomainConfig struct {
	API      []string `json:"api"`
	H5       []string `json:"h5"`
	Admin    []string `json:"admin"`
	Callback []string `json:"callback"`
}

// GameConfig 游戏配置
type GameConfig struct {
	GameId    int    `json:"game_id"`
	Type      int    `json:"type"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	ImageCode string `json:"image_code"`
	DbName    string `json:"db_name"`
	ApiId     string `json:"api_id"`
	ApiKey    string `json:"api_key"`
	ApiUrl    string `json:"api_url"`
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	PaymentId  int     `json:"payment_id"`
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Gateway    int     `json:"gateway"`
	ApiUrl     string  `json:"api_url"`
	MerchantNo string  `json:"merchant_no"`
	Md5Key     string  `json:"md5_key"`
	EachMin    float64 `json:"each_min"`
	EachMax    float64 `json:"each_max"`
	Sort       int     `json:"sort"`
	MoneyList  string  `json:"money_list"`
	IsInput    int     `json:"is_input"`
	Logo       string  `json:"logo"`
	Status     int     `json:"status"`
}

// WithdrawConfig 提现配置
type WithdrawConfig struct {
	WithdrawId int     `json:"withdraw_id"`
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Gateway    int     `json:"gateway"`
	ApiUrl     string  `json:"api_url"`
	MerchantNo string  `json:"merchant_no"`
	Md5Key     string  `json:"md5_key"`
	EachMin    float64 `json:"each_min"`
	EachMax    float64 `json:"each_max"`
	Sort       int     `json:"sort"`
	MoneyList  string  `json:"money_list"`
	IsInput    int     `json:"is_input"`
	Logo       string  `json:"logo"`
	Status     int     `json:"status"`
}

// DBConfigManager 数据库配置管理器
type DBConfigManager struct {
	consul      *api.Client
	siteConfigs map[string]*SiteConfig // key: site_code, value: config
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

var (
	dbManager *DBConfigManager
	once      sync.Once
)

// GetDBConfigManager 获取数据库配置管理器单例
func GetDBConfigManager() *DBConfigManager {
	once.Do(func() {
		dbManager = &DBConfigManager{
			siteConfigs: make(map[string]*SiteConfig),
		}
	})
	return dbManager
}

// Init 初始化数据库配置管理器
func (m *DBConfigManager) Init(consulClient *api.Client) error {
	m.consul = consulClient
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// 初始加载所有站点配置
	if err := m.loadAllSiteConfigs(); err != nil {
		g.Log().Errorf(m.ctx, "初始加载站点配置失败: %v", err)
		return err
	}

	// 启动配置监听
	go m.watchSiteConfigs()

	g.Log().Info(m.ctx, "Balance服务数据库配置管理器初始化完成")
	return nil
}

// Stop 停止配置管理器
func (m *DBConfigManager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	g.Log().Info(m.ctx, "Balance服务数据库配置管理器已停止")
}

// loadAllSiteConfigs 加载所有站点配置
func (m *DBConfigManager) loadAllSiteConfigs() error {
	kv := m.consul.KV()

	// 获取所有 sites/ 前缀的配置
	pairs, _, err := kv.List("sites/", nil)
	if err != nil {
		return fmt.Errorf("获取站点配置列表失败: %v", err)
	}

	updatedCount := 0
	for _, pair := range pairs {
		// 解析 key: sites/{site_code}
		parts := strings.Split(pair.Key, "/")
		if len(parts) != 2 {
			continue
		}

		updated, err := m.updateSiteConfig(pair.Key, pair.Value)
		if err != nil {
			g.Log().Errorf(m.ctx, "更新站点配置失败: key=%s, error=%v", pair.Key, err)
		} else if updated {
			updatedCount++
		}
	}

	// 删除定时任务的日志输出（避免日志过多）
	// if updatedCount > 0 {
	// 	g.Log().Infof(m.ctx, "Balance服务更新了 %d 个站点配置，当前共 %d 个站点", updatedCount, len(m.siteConfigs))
	// }
	return nil
}

// watchSiteConfigs 监听站点配置变化
func (m *DBConfigManager) watchSiteConfigs() {
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			// 每30秒检查一次配置变化
			time.Sleep(30 * time.Second)
			if err := m.loadAllSiteConfigs(); err != nil {
				g.Log().Errorf(m.ctx, "重新加载站点配置失败: %v", err)
			}
		}
	}
}

// updateSiteConfig 更新站点配置，返回是否有更新
func (m *DBConfigManager) updateSiteConfig(key string, value []byte) (bool, error) {
	var config SiteConfig
	if err := json.Unmarshal(value, &config); err != nil {
		return false, fmt.Errorf("解析站点配置失败: %v", err)
	}

	siteCode := config.Base.Code
	if siteCode == "" {
		return false, fmt.Errorf("站点代码为空")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查配置是否有变化
	oldConfig, exists := m.siteConfigs[siteCode]
	if exists && oldConfig.Meta.Version >= config.Meta.Version {
		return false, nil // 配置没有变化或版本更旧
	}

	// 检查数据库连接是否变化
	dbChanged := false
	if !exists {
		// 新站点，需要注册连接池
		dbChanged = true
	} else {
		// 检查Balance数据库连接是否有变化
		if oldConfig.Base.DbLinkBalance != config.Base.DbLinkBalance {
			dbChanged = true
			g.Log().Infof(m.ctx, "站点 %s Balance数据库连接配置发生变化，将重新注册连接池", siteCode)
		}
	}

	// 更新配置到内存
	m.siteConfigs[siteCode] = &config

	// 只在数据库连接变化时才重新注册连接池
	if dbChanged {
		if err := m.createBalanceDBConnection(siteCode, &config); err != nil {
			g.Log().Errorf(m.ctx, "注册Balance数据库配置组失败: site=%s, error=%v", siteCode, err)
			return false, err
		}
		// 删除定时任务的日志输出
		// g.Log().Infof(m.ctx, "Balance服务站点配置已更新（含DB连接池）: site=%s, version=%d", siteCode, config.Meta.Version)
	}
	// 删除定时任务的日志输出
	// else {
	// 	g.Log().Debugf(m.ctx, "Balance服务站点配置已更新（仅业务配置）: site=%s, version=%d", siteCode, config.Meta.Version)
	// }

	// 同步支付配置到数据库
	if err := m.syncPaymentConfig(siteCode, &config); err != nil {
		g.Log().Errorf(m.ctx, "同步支付配置失败: site=%s, error=%v", siteCode, err)
		// 不返回错误，允许继续
	}

	// 同步提现配置到数据库
	if err := m.syncWithdrawConfig(siteCode, &config); err != nil {
		g.Log().Errorf(m.ctx, "同步提现配置失败: site=%s, error=%v", siteCode, err)
		// 不返回错误，允许继续
	}

	return true, nil
}

// createBalanceDBConnection 为站点创建Balance数据库连接
func (m *DBConfigManager) createBalanceDBConnection(siteCode string, config *SiteConfig) error {
	// 如果没有Balance数据库连接配置，跳过
	if config.Base.DbLinkBalance == "" {
		g.Log().Warningf(m.ctx, "站点 %s 没有配置Balance数据库连接", siteCode)
		return nil
	}

	// 创建数据库配置组名称
	groupName := fmt.Sprintf("site_%s", siteCode)

	// 解析连接字符串，移除 "mysql:" 前缀
	link := config.Base.DbLinkBalance
	if len(link) > 6 && link[:6] == "mysql:" {
		link = link[6:] // 移除 "mysql:"
	}

	// 动态设置数据库配置
	configNode := gdb.ConfigNode{
		Type: "mysql",
		Link: link,
	}

	gdb.AddConfigNode(groupName, configNode)

	// 验证配置并测试连接
	registeredConfig := gdb.GetConfig(groupName)
	if registeredConfig != nil && len(registeredConfig) > 0 {
		// 尝试实际连接测试
		testDB := g.DB(groupName)
		if testDB != nil {
			_, err := testDB.GetValue(m.ctx, "SELECT 1")
			if err != nil {
				g.Log().Errorf(m.ctx, "Balance数据库连接测试失败: %s, error: %v", groupName, err)
				return fmt.Errorf("数据库连接测试失败: %v", err)
			}
			g.Log().Debugf(m.ctx, "Balance数据库连接池注册成功: %s", groupName)
		}
	} else {
		g.Log().Errorf(m.ctx, "Balance配置注册失败: %s", groupName)
		return fmt.Errorf("配置注册失败")
	}

	return nil
}

// GetSiteConfig 获取站点配置
func (m *DBConfigManager) GetSiteConfig(siteCode string) (*SiteConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, exists := m.siteConfigs[siteCode]
	if !exists {
		return nil, fmt.Errorf("站点配置不存在: site=%s", siteCode)
	}

	return config, nil
}

// GetAllSiteCodes 获取所有站点代码
func (m *DBConfigManager) GetAllSiteCodes() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	codes := make([]string, 0, len(m.siteConfigs))
	for code := range m.siteConfigs {
		codes = append(codes, code)
	}

	return codes
}

// GetSiteCodeBySiteID 根据站点ID获取站点代码
func (m *DBConfigManager) GetSiteCodeBySiteID(siteID int64) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for code, config := range m.siteConfigs {
		if int64(config.Base.SiteId) == siteID {
			return code
		}
	}

	return ""
}

// syncPaymentConfig 同步支付配置到数据库
func (m *DBConfigManager) syncPaymentConfig(siteCode string, config *SiteConfig) error {
	// 获取站点的数据库连接
	groupName := fmt.Sprintf("site_%s", siteCode)
	db := g.DB(groupName)
	if db == nil {
		g.Log().Warningf(m.ctx, "站点 %s 数据库连接不存在，跳过支付配置同步", siteCode)
		return nil
	}

	// 检查site_payment表是否存在
	tables, err := db.Tables(m.ctx)
	if err != nil {
		return fmt.Errorf("查询数据库表失败: %v", err)
	}

	hasTable := false
	for _, table := range tables {
		if table == "site_payment" {
			hasTable = true
			break
		}
	}

	if !hasTable {
		g.Log().Warningf(m.ctx, "站点 %s 数据库中不存在site_payment表，跳过支付配置同步", siteCode)
		return nil
	}

	siteId := config.Base.SiteId
	g.Log().Infof(m.ctx, "开始同步站点 %s (site_id=%d) 的支付配置，共 %d 个支付", siteCode, siteId, len(config.Payment))

	// 使用事务处理
	err = db.Transaction(m.ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 获取当前数据库中的所有支付ID
		type PaymentIdResult struct {
			PaymentId int `json:"payment_id"`
		}
		var existingPayments []PaymentIdResult
		err := tx.Model("site_payment").
			Where("site_id", siteId).
			Fields("payment_id").
			Scan(&existingPayments)
		if err != nil {
			return fmt.Errorf("查询现有支付失败: %v", err)
		}

		existingMap := make(map[int]bool)
		for _, p := range existingPayments {
			existingMap[p.PaymentId] = true
		}

		// 2. 处理配置中的支付
		configPaymentIds := make(map[int]bool)
		for _, payment := range config.Payment {
			configPaymentIds[payment.PaymentId] = true

			// 检查是否已存在
			count, err := tx.Model("site_payment").
				Where("site_id", siteId).
				Where("payment_id", payment.PaymentId).
				Count()
			if err != nil {
				g.Log().Errorf(ctx, "检查支付是否存在失败: payment_id=%d, error=%v", payment.PaymentId, err)
				continue
			}

			if count > 0 {
				// 更新
				_, err = tx.Model("site_payment").
					Where("site_id", siteId).
					Where("payment_id", payment.PaymentId).
					Data(g.Map{
						"name":        payment.Name,
						"code":        payment.Code,
						"gateway":     payment.Gateway,
						"api_url":     payment.ApiUrl,
						"merchant_no": payment.MerchantNo,
						"md5_key":     payment.Md5Key,
						"each_min":    payment.EachMin,
						"each_max":    payment.EachMax,
						"sort":        payment.Sort,
						"moneyList":   payment.MoneyList,
						"is_input":    payment.IsInput,
						"logo":        payment.Logo,
						"status":      payment.Status,
						"updated_at":  "NOW()",
					}).
					Update()
				if err != nil {
					g.Log().Errorf(ctx, "更新支付配置失败: payment_id=%d, error=%v", payment.PaymentId, err)
				} else {
					g.Log().Debugf(ctx, "更新支付配置成功: site=%s, payment_id=%d, name=%s", siteCode, payment.PaymentId, payment.Name)
				}
			} else {
				// 插入
				_, err = tx.Model("site_payment").
					Data(g.Map{
						"site_id":     siteId,
						"payment_id":  payment.PaymentId,
						"name":        payment.Name,
						"code":        payment.Code,
						"gateway":     payment.Gateway,
						"api_url":     payment.ApiUrl,
						"merchant_no": payment.MerchantNo,
						"md5_key":     payment.Md5Key,
						"each_min":    payment.EachMin,
						"each_max":    payment.EachMax,
						"sort":        payment.Sort,
						"moneyList":   payment.MoneyList,
						"is_input":    payment.IsInput,
						"logo":        payment.Logo,
						"status":      payment.Status,
						"created_at":  "NOW()",
						"updated_at":  "NOW()",
					}).
					Insert()
				if err != nil {
					g.Log().Errorf(ctx, "插入支付配置失败: payment_id=%d, error=%v", payment.PaymentId, err)
				} else {
					g.Log().Infof(ctx, "插入支付配置成功: site=%s, payment_id=%d, name=%s", siteCode, payment.PaymentId, payment.Name)
				}
			}
		}

		// 3. 删除配置中不存在的支付
		for paymentId := range existingMap {
			if !configPaymentIds[paymentId] {
				_, err = tx.Model("site_payment").
					Where("site_id", siteId).
					Where("payment_id", paymentId).
					Delete()
				if err != nil {
					g.Log().Errorf(ctx, "删除支付配置失败: payment_id=%d, error=%v", paymentId, err)
				} else {
					g.Log().Infof(ctx, "删除支付配置成功: site=%s, payment_id=%d", siteCode, paymentId)
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("同步支付配置事务失败: %v", err)
	}

	g.Log().Infof(m.ctx, "站点 %s 支付配置同步完成", siteCode)
	return nil
}

// syncWithdrawConfig 同步提现配置到数据库
func (m *DBConfigManager) syncWithdrawConfig(siteCode string, config *SiteConfig) error {
	// 获取站点的数据库连接
	groupName := fmt.Sprintf("site_%s", siteCode)
	db := g.DB(groupName)
	if db == nil {
		g.Log().Warningf(m.ctx, "站点 %s 数据库连接不存在，跳过提现配置同步", siteCode)
		return nil
	}

	// 检查site_withdraw表是否存在
	tables, err := db.Tables(m.ctx)
	if err != nil {
		return fmt.Errorf("查询数据库表失败: %v", err)
	}

	hasTable := false
	for _, table := range tables {
		if table == "site_withdraw" {
			hasTable = true
			break
		}
	}

	if !hasTable {
		g.Log().Warningf(m.ctx, "站点 %s 数据库中不存在site_withdraw表，跳过提现配置同步", siteCode)
		return nil
	}

	siteId := config.Base.SiteId
	g.Log().Infof(m.ctx, "开始同步站点 %s (site_id=%d) 的提现配置，共 %d 个提现", siteCode, siteId, len(config.Withdraw))

	// 使用事务处理
	err = db.Transaction(m.ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 获取当前数据库中的所有提现ID
		type WithdrawIdResult struct {
			WithdrawId int `json:"withdraw_id"`
		}
		var existingWithdraws []WithdrawIdResult
		err := tx.Model("site_withdraw").
			Where("site_id", siteId).
			Fields("withdraw_id").
			Scan(&existingWithdraws)
		if err != nil {
			return fmt.Errorf("查询现有提现失败: %v", err)
		}

		existingMap := make(map[int]bool)
		for _, w := range existingWithdraws {
			existingMap[w.WithdrawId] = true
		}

		// 2. 处理配置中的提现
		configWithdrawIds := make(map[int]bool)
		for _, withdraw := range config.Withdraw {
			configWithdrawIds[withdraw.WithdrawId] = true

			// 检查是否已存在
			count, err := tx.Model("site_withdraw").
				Where("site_id", siteId).
				Where("withdraw_id", withdraw.WithdrawId).
				Count()
			if err != nil {
				g.Log().Errorf(ctx, "检查提现是否存在失败: withdraw_id=%d, error=%v", withdraw.WithdrawId, err)
				continue
			}

			if count > 0 {
				// 更新
				_, err = tx.Model("site_withdraw").
					Where("site_id", siteId).
					Where("withdraw_id", withdraw.WithdrawId).
					Data(g.Map{
						"name":        withdraw.Name,
						"code":        withdraw.Code,
						"gateway":     withdraw.Gateway,
						"api_url":     withdraw.ApiUrl,
						"merchant_no": withdraw.MerchantNo,
						"md5_key":     withdraw.Md5Key,
						"each_min":    withdraw.EachMin,
						"each_max":    withdraw.EachMax,
						"sort":        withdraw.Sort,
						"moneyList":   withdraw.MoneyList,
						"is_input":    withdraw.IsInput,
						"logo":        withdraw.Logo,
						"status":      withdraw.Status,
						"updated_at":  "NOW()",
					}).
					Update()
				if err != nil {
					g.Log().Errorf(ctx, "更新提现配置失败: withdraw_id=%d, error=%v", withdraw.WithdrawId, err)
				} else {
					g.Log().Debugf(ctx, "更新提现配置成功: site=%s, withdraw_id=%d, name=%s", siteCode, withdraw.WithdrawId, withdraw.Name)
				}
			} else {
				// 插入
				_, err = tx.Model("site_withdraw").
					Data(g.Map{
						"site_id":     siteId,
						"withdraw_id": withdraw.WithdrawId,
						"name":        withdraw.Name,
						"code":        withdraw.Code,
						"gateway":     withdraw.Gateway,
						"api_url":     withdraw.ApiUrl,
						"merchant_no": withdraw.MerchantNo,
						"md5_key":     withdraw.Md5Key,
						"each_min":    withdraw.EachMin,
						"each_max":    withdraw.EachMax,
						"sort":        withdraw.Sort,
						"moneyList":   withdraw.MoneyList,
						"is_input":    withdraw.IsInput,
						"logo":        withdraw.Logo,
						"status":      withdraw.Status,
						"created_at":  "NOW()",
						"updated_at":  "NOW()",
					}).
					Insert()
				if err != nil {
					g.Log().Errorf(ctx, "插入提现配置失败: withdraw_id=%d, error=%v", withdraw.WithdrawId, err)
				} else {
					g.Log().Infof(ctx, "插入提现配置成功: site=%s, withdraw_id=%d, name=%s", siteCode, withdraw.WithdrawId, withdraw.Name)
				}
			}
		}

		// 3. 删除配置中不存在的提现
		for withdrawId := range existingMap {
			if !configWithdrawIds[withdrawId] {
				_, err = tx.Model("site_withdraw").
					Where("site_id", siteId).
					Where("withdraw_id", withdrawId).
					Delete()
				if err != nil {
					g.Log().Errorf(ctx, "删除提现配置失败: withdraw_id=%d, error=%v", withdrawId, err)
				} else {
					g.Log().Infof(ctx, "删除提现配置成功: site=%s, withdraw_id=%d", siteCode, withdrawId)
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("同步提现配置事务失败: %v", err)
	}

	g.Log().Infof(m.ctx, "站点 %s 提现配置同步完成", siteCode)
	return nil
}
