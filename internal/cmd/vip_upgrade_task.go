package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	gamebackendv1 "jh_balance_service/api/grpc/game/backend/v1"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/grpc_client"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/registry"
	"jh_balance_service/internal/util"
)

const (
	vipUpgradeBatchSize = 500
)

var (
	VipUpgrade = &gcmd.Command{
		Name:  "vip-upgrade",
		Usage: "vip-upgrade",
		Brief: "执行一次VIP升级批处理任务",
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			g.Log().Info(ctx, "VIP Upgrade Batch 启动中...")

			// 初始化Consul客户端
			consulClient, err := registry.InitConsul()
			if err != nil {
				g.Log().Errorf(ctx, "Failed to initialize Consul: %v", err)
				return err
			}

			// 初始化数据库配置管理器
			dbManager := registry.GetDBConfigManager()
			if err := dbManager.Init(consulClient); err != nil {
				g.Log().Errorf(ctx, "Failed to init balance db config manager: %v", err)
				return err
			}
			defer dbManager.Stop()

			if err := runVipUpgradeBatch(ctx, dbManager); err != nil {
				g.Log().Errorf(ctx, "VIP Upgrade Batch 执行失败: %v", err)
				return err
			}

			g.Log().Info(ctx, "VIP Upgrade Batch 执行完成")
			return nil
		},
	}
)

func runVipUpgradeBatch(ctx context.Context, dbManager *registry.DBConfigManager) error {
	siteCodes := dbManager.GetAllSiteCodes()
	var firstErr error
	for _, siteCode := range siteCodes {
		if err := runVipUpgradeForSite(ctx, dbManager, siteCode); err != nil {
			g.Log().Errorf(ctx, "VIP升级任务失败: site=%s, err=%v", siteCode, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func runVipUpgradeForSite(ctx context.Context, dbManager *registry.DBConfigManager, siteCode string) error {
	cfg, err := dbManager.GetSiteConfig(siteCode)
	if err != nil {
		return err
	}
	if cfg.Base.Status == 0 {
		return nil
	}
	if cfg.Base.SiteId <= 0 {
		return fmt.Errorf("站点ID为空: site=%s", siteCode)
	}

	balanceGroup := "site_" + siteCode
	if gdb.GetConfig(balanceGroup) == nil {
		return fmt.Errorf("Balance数据库组不存在: %s", balanceGroup)
	}

	siteGroup := "sitecfg_" + siteCode
	if cfg.Base.DbLinkSite == "" {
		return fmt.Errorf("站点数据库连接为空: site=%s", siteCode)
	}
	if err := ensureMySQLGroup(ctx, siteGroup, cfg.Base.DbLinkSite); err != nil {
		return fmt.Errorf("站点数据库连接失败: %s, err=%w", siteGroup, err)
	}

	balanceCtx := context.WithValue(ctx, "dynamic_db_group", balanceGroup)

	depositUsers, err := syncDepositTotals(balanceCtx, balanceGroup, cfg.Base.SiteId)
	if err != nil {
		return fmt.Errorf("同步充值统计失败: %w", err)
	}
	betUsers, err := syncBetTotals(balanceCtx, balanceGroup, cfg.Base.SiteId, siteCode)
	if err != nil {
		return fmt.Errorf("同步投注统计失败: %w", err)
	}
	reportUsers, err := fetchReportUserIDs(balanceCtx, balanceGroup, cfg.Base.SiteId)
	if err != nil {
		return fmt.Errorf("读取累计统计用户失败: %w", err)
	}

	userIds := mergeUserIds(mergeUserIds(depositUsers, betUsers), reportUsers)
	if len(userIds) == 0 {
		return nil
	}

	totals, err := fetchUserTotals(balanceCtx, balanceGroup, cfg.Base.SiteId, userIds)
	if err != nil {
		return fmt.Errorf("读取统计数据失败: %w", err)
	}

	users, err := fetchUserInfos(ctx, siteGroup, cfg.Base.SiteId, userIds)
	if err != nil {
		return fmt.Errorf("读取用户信息失败: %w", err)
	}

	grades, err := fetchUserGrades(ctx, siteGroup, cfg.Base.SiteId)
	if err != nil {
		return fmt.Errorf("读取等级配置失败: %w", err)
	}
	if len(grades) == 0 {
		return nil
	}

	for _, userId := range userIds {
		total, ok := totals[userId]
		if !ok {
			continue
		}
		user, ok := users[userId]
		if !ok {
			continue
		}

		nextGrade := findNextGrade(user.GradeId, grades)
		if nextGrade == nil {
			continue
		}

		if total.DepositAmount < nextGrade.PaymentLimit || total.BetAmount < nextGrade.BetLimit {
			continue
		}

		updated, err := updateUserGrade(ctx, siteGroup, cfg.Base.SiteId, userId, user.GradeId, nextGrade.Id)
		if err != nil {
			g.Log().Errorf(ctx, "更新会员等级失败: site=%s, user_id=%d, err=%v", siteCode, userId, err)
			continue
		}
		if !updated {
			continue
		}

		if nextGrade.MoneyLimit > 0 {
			if _, err := grantVipUpgradeBonus(balanceCtx, cfg.Base.SiteId, userId, nextGrade.MoneyLimit, nextGrade.Name, nextGrade.Id); err != nil {
				g.Log().Errorf(ctx, "发放升级红利失败: site=%s, user_id=%d, err=%v", siteCode, userId, err)
			}
		}
	}

	return nil
}

type aggRow struct {
	UserId  int64       `json:"user_id"`
	Amount  float64     `json:"amount"`
	MaxId   int64       `json:"max_id"`
	MaxTime *gtime.Time `json:"max_time"`
}

func syncDepositTotals(ctx context.Context, balanceGroup string, siteId int) ([]int64, error) {
	var rows []aggRow
	err := g.DB(balanceGroup).Model("balance_log_payment").Ctx(ctx).As("p").
		LeftJoin("user_report_total r", "r.site_id=p.site_id AND r.user_id=p.user_id").
		Fields("p.user_id, SUM(p.actual_money) AS amount, MAX(p.id) AS max_id, MAX(p.created_at) AS max_time").
		Where("p.site_id", siteId).
		Where("p.status", balanceConsts.StatusSuccess).
		Where("p.actual_money > 0").
		Where("p.id > IFNULL(r.last_deposit_id,0)").
		Group("p.user_id").
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	userIds := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.UserId <= 0 || row.Amount <= 0 || row.MaxId <= 0 {
			continue
		}
		if err := upsertDepositTotal(ctx, balanceGroup, siteId, row); err != nil {
			g.Log().Errorf(ctx, "更新充值统计失败: user_id=%d, err=%v", row.UserId, err)
			continue
		}
		userIds = append(userIds, row.UserId)
	}

	return userIds, nil
}

type betCursorRow struct {
	Id        int64  `json:"id"`
	UserId    int64  `json:"user_id"`
	LastBetId uint64 `json:"last_bet_id"`
}

func syncBetTotals(ctx context.Context, balanceGroup string, siteId int, siteCode string) ([]int64, error) {
	gameClient, gameConn, err := grpc_client.GetGameBackendClient(ctx)
	if err != nil {
		return nil, err
	}
	defer gameConn.Close()

	userSet := make(map[int64]struct{})
	var cursorId int64

	for {
		var rows []betCursorRow
		err := g.DB(balanceGroup).Model("user_report_total").Ctx(ctx).
			Fields("id, user_id, last_bet_id").
			Where("site_id", siteId).
			Where("id > ?", cursorId).
			Order("id ASC").
			Limit(vipUpgradeBatchSize).
			Scan(&rows)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			break
		}

		cursors := make([]*gamebackendv1.UserBetCursorItem, 0, len(rows))
		for _, row := range rows {
			if row.Id > cursorId {
				cursorId = row.Id
			}
			if row.UserId <= 0 {
				continue
			}
			cursors = append(cursors, &gamebackendv1.UserBetCursorItem{
				UserId:    row.UserId,
				LastBetId: row.LastBetId,
			})
		}

		if len(cursors) > 0 {
			resp, err := gameClient.GetUserBetIncrementBatch(ctx, &gamebackendv1.GetUserBetIncrementBatchReq{
				SiteId:   int32(siteId),
				SiteCode: siteCode,
				Cursors:  cursors,
			})
			if err != nil {
				return nil, err
			}
			if resp == nil || !resp.Success {
				if resp != nil {
					return nil, fmt.Errorf("GetUserBetIncrementBatch失败: %s", resp.Message)
				}
				return nil, fmt.Errorf("GetUserBetIncrementBatch失败: 空响应")
			}

			for _, item := range resp.List {
				if item == nil {
					continue
				}
				if item.UserId <= 0 || item.MaxBetId == 0 {
					continue
				}
				amount := gconv.Float64(item.ValidAmount)
				if amount <= 0 {
					continue
				}
				row := aggRow{
					UserId:  item.UserId,
					Amount:  amount,
					MaxId:   int64(item.MaxBetId),
					MaxTime: parseNullableTime(item.MaxSettleTime),
				}
				if err := upsertBetTotal(ctx, balanceGroup, siteId, row); err != nil {
					g.Log().Errorf(ctx, "更新投注统计失败: user_id=%d, err=%v", row.UserId, err)
					continue
				}
				userSet[row.UserId] = struct{}{}
			}
		}

		if len(rows) < vipUpgradeBatchSize {
			break
		}
	}

	userIds := make([]int64, 0, len(userSet))
	for userId := range userSet {
		userIds = append(userIds, userId)
	}
	return userIds, nil
}

func upsertDepositTotal(ctx context.Context, balanceGroup string, siteId int, row aggRow) error {
	_, err := g.DB(balanceGroup).Exec(ctx, `
INSERT INTO user_report_total
(site_id, user_id, deposit_amount, last_deposit_id, last_deposit_time, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, NOW(), NOW())
ON DUPLICATE KEY UPDATE
deposit_amount = deposit_amount + VALUES(deposit_amount),
last_deposit_id = GREATEST(last_deposit_id, VALUES(last_deposit_id)),
last_deposit_time = IF(last_deposit_time IS NULL OR last_deposit_time < VALUES(last_deposit_time), VALUES(last_deposit_time), last_deposit_time),
updated_at = NOW()
`, siteId, row.UserId, row.Amount, row.MaxId, row.MaxTime)
	return err
}

func upsertBetTotal(ctx context.Context, balanceGroup string, siteId int, row aggRow) error {
	_, err := g.DB(balanceGroup).Exec(ctx, `
INSERT INTO user_report_total
(site_id, user_id, bet_amount, last_bet_id, last_bet_settle_time, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, NOW(), NOW())
ON DUPLICATE KEY UPDATE
bet_amount = bet_amount + VALUES(bet_amount),
last_bet_id = GREATEST(last_bet_id, VALUES(last_bet_id)),
last_bet_settle_time = IF(last_bet_settle_time IS NULL OR last_bet_settle_time < VALUES(last_bet_settle_time), VALUES(last_bet_settle_time), last_bet_settle_time),
updated_at = NOW()
`, siteId, row.UserId, row.Amount, row.MaxId, row.MaxTime)
	return err
}

type userTotal struct {
	UserId        int64   `json:"user_id"`
	DepositAmount float64 `json:"deposit_amount"`
	BetAmount     float64 `json:"bet_amount"`
}

type reportUserRow struct {
	Id     int64 `json:"id"`
	UserId int64 `json:"user_id"`
}

func fetchReportUserIDs(ctx context.Context, balanceGroup string, siteId int) ([]int64, error) {
	set := make(map[int64]struct{})
	var cursorId int64

	for {
		var rows []reportUserRow
		err := g.DB(balanceGroup).Model("user_report_total").Ctx(ctx).
			Fields("id, user_id").
			Where("site_id", siteId).
			Where("id > ?", cursorId).
			Order("id ASC").
			Limit(vipUpgradeBatchSize).
			Scan(&rows)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			if row.Id > cursorId {
				cursorId = row.Id
			}
			if row.UserId > 0 {
				set[row.UserId] = struct{}{}
			}
		}
		if len(rows) < vipUpgradeBatchSize {
			break
		}
	}

	userIds := make([]int64, 0, len(set))
	for userId := range set {
		userIds = append(userIds, userId)
	}
	return userIds, nil
}

func fetchUserTotals(ctx context.Context, balanceGroup string, siteId int, userIds []int64) (map[int64]userTotal, error) {
	result := make(map[int64]userTotal, len(userIds))
	for _, chunk := range chunkIDs(userIds, vipUpgradeBatchSize) {
		var rows []userTotal
		if err := g.DB(balanceGroup).Model("user_report_total").Ctx(ctx).
			Fields("user_id, deposit_amount, bet_amount").
			Where("site_id", siteId).
			Where("user_id IN (?)", chunk).
			Scan(&rows); err != nil {
			return nil, err
		}
		for _, row := range rows {
			result[row.UserId] = row
		}
	}
	return result, nil
}

type userInfo struct {
	UserId   int64  `json:"user_id"`
	GradeId  int    `json:"grade_id"`
	Username string `json:"username"`
}

func fetchUserInfos(ctx context.Context, siteGroup string, siteId int, userIds []int64) (map[int64]userInfo, error) {
	result := make(map[int64]userInfo, len(userIds))
	for _, chunk := range chunkIDs(userIds, vipUpgradeBatchSize) {
		var rows []userInfo
		if err := g.DB(siteGroup).Model("user").Ctx(ctx).
			Fields("id AS user_id, grade_id, username").
			Where("site_id", siteId).
			Where("id IN (?)", chunk).
			Scan(&rows); err != nil {
			return nil, err
		}
		for _, row := range rows {
			result[row.UserId] = row
		}
	}
	return result, nil
}

type gradeInfo struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	PaymentLimit float64 `json:"payment_limit"`
	BetLimit     float64 `json:"bet_limit"`
	MoneyLimit   float64 `json:"money_limit"`
}

func fetchUserGrades(ctx context.Context, siteGroup string, siteId int) ([]gradeInfo, error) {
	var grades []gradeInfo
	if err := g.DB(siteGroup).Model("user_grade").Ctx(ctx).
		Fields("id, name, payment_limit, bet_limit, money_limit").
		Where("site_id", siteId).
		Where("status", 1).
		Order("id ASC").
		Scan(&grades); err != nil {
		return nil, err
	}
	return grades, nil
}

func findNextGrade(currentGradeId int, grades []gradeInfo) *gradeInfo {
	for i := range grades {
		if grades[i].Id > currentGradeId {
			return &grades[i]
		}
	}
	return nil
}

func updateUserGrade(ctx context.Context, siteGroup string, siteId int, userId int64, oldGradeId, newGradeId int) (bool, error) {
	now := gtime.Now()
	result, err := g.DB(siteGroup).Model("user").Ctx(ctx).
		Where("site_id", siteId).
		Where("id", userId).
		Where("grade_id", oldGradeId).
		Data(g.Map{
			"grade_id":   newGradeId,
			"updated_at": now,
		}).Update()
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func grantVipUpgradeBonus(ctx context.Context, siteId int, userId int64, amount float64, gradeName string, gradeId int) (string, error) {
	if siteId <= 0 {
		return "", fmt.Errorf("站点ID不能为空")
	}
	if userId <= 0 {
		return "", fmt.Errorf("用户ID不能为空")
	}
	if amount <= 0 {
		return "", fmt.Errorf("红利金额必须大于0")
	}

	// 兜底设置动态数据库组（避免无metadata时路由失败）
	if ctx.Value("dynamic_db_group") == nil {
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	// 优先使用升级送流水规则，未配置则回退红利规则
	rule, err := dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeUpgradeActivity)
	if err != nil {
		return "", fmt.Errorf("查询流水规则失败")
	}
	if rule == nil {
		rule, err = dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeBonus)
		if err != nil {
			return "", fmt.Errorf("查询流水规则失败")
		}
	}
	if rule == nil {
		return "", fmt.Errorf("未配置可用的流水规则")
	}

	var tradeNo string
	now := gtime.Now()

	err = dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var balanceUser *entity.BalanceUser
		if err := dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("site_id", siteId).
			Where("user_id", userId).
			LockUpdate().
			Scan(&balanceUser); err != nil {
			return err
		}
		if balanceUser == nil || balanceUser.Id == 0 {
			return fmt.Errorf("用户余额记录不存在")
		}

		balanceOld := balanceUser.Balance
		balanceNew := balanceOld + amount

		if _, err := dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("id", balanceUser.Id).
			Update(g.Map{
				"balance":    balanceNew,
				"updated_at": now,
			}); err != nil {
			return err
		}

		tradeNo = util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_BONUS)
		remark := fmt.Sprintf("升级VIP%d礼金", gradeId)
		if strings.TrimSpace(gradeName) != "" {
			remark = fmt.Sprintf("升级%s礼金", strings.TrimSpace(gradeName))
		}

		if _, err := dao.BalanceLog.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":      siteId,
			"user_id":      userId,
			"trade_type":   balanceConsts.LOG_TYPE_BONUS,
			"trade_no":     tradeNo,
			"money":        amount,
			"actual_money": amount,
			"fee":          0,
			"status":       balanceConsts.StatusSuccess,
			"remark":       remark,
			"created_at":   now,
			"updated_at":   now,
		}); err != nil {
			return err
		}

		if _, err := dao.BalanceChange.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"trade_type":     balanceConsts.TradeTypeBonus,
			"user_id":        userId,
			"username":       balanceUser.Username,
			"trade_no":       tradeNo,
			"balance_old":    balanceOld,
			"money":          amount,
			"balance_new":    balanceNew,
			"balance_frozen": balanceUser.BalanceFrozen,
			"status":         balanceConsts.StatusSuccess,
			"remark":         remark,
			"created_at":     now,
			"updated_at":     now,
		}); err != nil {
			return err
		}

		bonusResult, err := dao.BalanceLogBonus.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":              siteId,
			"trade_type":           balanceConsts.BonusTradeTypeUpgrade,
			"user_id":              userId,
			"username":             balanceUser.Username,
			"trade_no":             tradeNo,
			"activity_recharge_id": 0,
			"money":                amount,
			"fee":                  0,
			"status":               balanceConsts.StatusSuccess,
			"admin_id":             0,
			"remark":               remark,
			"created_at":           now,
			"updated_at":           now,
		})
		if err != nil {
			return err
		}
		bonusId, _ := bonusResult.LastInsertId()

		flowMultiple := rule.FlowMultiple
		expireDays := rule.ExpireDays
		requiredFlow := amount * flowMultiple
		var expireTime *gtime.Time
		if expireDays > 0 {
			expireTime = now.AddDate(0, 0, expireDays)
		}

		if _, err := dao.BalanceFlowRequirement.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"user_id":        userId,
			"username":       balanceUser.Username,
			"source_type":    balanceConsts.SourceTypeUpgradeActivity,
			"source_id":      bonusId,
			"trade_no":       tradeNo,
			"amount":         0,
			"bonus_amount":   amount,
			"flow_multiple":  flowMultiple,
			"required_flow":  requiredFlow,
			"completed_flow": 0,
			"remaining_flow": requiredFlow,
			"status":         1,
			"progress":       0,
			"start_time":     now,
			"expire_time":    expireTime,
			"complete_time":  nil,
			"remark":         remark,
			"admin_id":       0,
			"created_at":     now,
			"updated_at":     now,
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return tradeNo, nil
}

func mergeUserIds(a, b []int64) []int64 {
	set := make(map[int64]struct{}, len(a)+len(b))
	for _, id := range a {
		if id > 0 {
			set[id] = struct{}{}
		}
	}
	for _, id := range b {
		if id > 0 {
			set[id] = struct{}{}
		}
	}
	if len(set) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	return ids
}

func chunkIDs(ids []int64, size int) [][]int64 {
	if size <= 0 {
		size = vipUpgradeBatchSize
	}
	if len(ids) <= size {
		return [][]int64{ids}
	}
	var chunks [][]int64
	for i := 0; i < len(ids); i += size {
		end := i + size
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[i:end])
	}
	return chunks
}

func ensureMySQLGroup(ctx context.Context, groupName, link string) error {
	if groupName == "" || link == "" {
		return fmt.Errorf("参数为空")
	}
	if gdb.GetConfig(groupName) != nil {
		return nil
	}

	if strings.HasPrefix(link, "mysql:") {
		link = link[6:]
	}

	configNode := gdb.ConfigNode{
		Type: "mysql",
		Link: link,
	}
	gdb.AddConfigNode(groupName, configNode)

	registered := gdb.GetConfig(groupName)
	if registered == nil || len(registered) == 0 {
		return fmt.Errorf("数据库配置注册失败")
	}

	testDB := g.DB(groupName)
	if testDB == nil {
		return fmt.Errorf("数据库实例为空")
	}
	if _, err := testDB.GetValue(ctx, "SELECT 1"); err != nil {
		return err
	}
	return nil
}

func parseNullableTime(value string) *gtime.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	t, err := gtime.StrToTime(value)
	if err != nil {
		return nil
	}
	return t
}
