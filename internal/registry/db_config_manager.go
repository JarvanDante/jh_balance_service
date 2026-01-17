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
	Meta   MetaConfig   `json:"Meta"`
	Base   BaseConfig   `json:"Base"`
	Domain DomainConfig `json:"Domain"`
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

	// 只在有配置更新时输出日志
	if updatedCount > 0 {
		g.Log().Infof(m.ctx, "Balance服务更新了 %d 个站点配置，当前共 %d 个站点", updatedCount, len(m.siteConfigs))
	}
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
		g.Log().Infof(m.ctx, "Balance服务站点配置已更新（含DB连接池）: site=%s, version=%d", siteCode, config.Meta.Version)
	} else {
		g.Log().Debugf(m.ctx, "Balance服务站点配置已更新（仅业务配置）: site=%s, version=%d", siteCode, config.Meta.Version)
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
