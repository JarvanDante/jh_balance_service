package flow

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
)

// RecordFlowRequirementRequest 记录流水要求请求
// SourceType: 1=充值, 2=红利, 3=返水, 4=活动, 5=人工加款
// FlowMultiple/ExpireDays 可选，若不传则按规则表读取
// TradeNo/SourceId 用于关联订单
// BonusAmount 可为 0
// AdminId 仅人工操作时使用
// Remark 可选
// SiteId 可选，默认从上下文读取
// Username 必填
// Amount 必填
// UserId 必填
// SourceId 必填（如无则传 0）
// TradeNo 可为空
//
// FIFO: 记录按创建时间入库，消费时按 created_at 升序处理
//
// 返回插入ID和完整记录
//
// 注意：此方法只负责记录“需要完成的流水”，不负责更新完成进度
//
// 如果 FlowMultiple 未传，按站点规则表匹配；若匹配不到则返回错误
//
// ExpireDays 未传则使用规则表配置（0 表示永久有效）
//
// Status: 1=进行中
// Progress: 0
// CompletedFlow: 0
// RemainingFlow: required_flow
// RequiredFlow = (amount + bonus_amount) * flow_multiple
//
// StartTime = now
// ExpireTime = now + expire_days (expire_days == 0 -> NULL)
type RecordFlowRequirementRequest struct {
	SiteId       int
	UserId       int64
	Username     string
	SourceType   int
	SourceId     int64
	TradeNo      string
	Amount       float64
	BonusAmount  float64
	FlowMultiple float64
	ExpireDays   int
	Remark       string
	AdminId      int
}

// RecordFlowRequirement 记录流水要求
func RecordFlowRequirement(ctx context.Context, req *RecordFlowRequirementRequest) (int64, *entity.BalanceFlowRequirement, error) {
	if req == nil {
		return 0, nil, gerror.New("请求参数为空")
	}
	if req.UserId <= 0 {
		return 0, nil, gerror.New("用户ID不能为空")
	}
	if req.Username == "" {
		return 0, nil, gerror.New("用户名不能为空")
	}
	if req.SourceType <= 0 {
		return 0, nil, gerror.New("来源类型不能为空")
	}
	if req.Amount < 0 {
		return 0, nil, gerror.New("金额不能小于0")
	}

	siteId := req.SiteId
	if siteId <= 0 {
		if ctxSiteId, ok := middleware.GetSiteIdFromContext(ctx); ok {
			siteId = int(ctxSiteId)
		}
	}
	if siteId <= 0 {
		return 0, nil, gerror.New("站点ID不能为空")
	}

	flowMultiple := req.FlowMultiple
	expireDays := req.ExpireDays
	if flowMultiple <= 0 {
		ruleType := mapSourceTypeToRuleType(req.SourceType)
		if ruleType == 0 {
			return 0, nil, gerror.New("不支持的来源类型")
		}
		rule, err := dao.BalanceFlowRule.GetActiveRule(ctx, siteId, ruleType)
		if err != nil {
			return 0, nil, gerror.Wrap(err, "查询流水规则失败")
		}
		if rule == nil {
			return 0, nil, gerror.New("未配置可用的流水规则")
		}
		flowMultiple = rule.FlowMultiple
		expireDays = rule.ExpireDays
	}

	requiredFlow := (req.Amount + req.BonusAmount) * flowMultiple
	now := gtime.Now()
	var expireTime *gtime.Time
	if expireDays > 0 {
		expireTime = now.AddDate(0, 0, expireDays)
	}

	record := &entity.BalanceFlowRequirement{
		SiteId:        siteId,
		UserId:        req.UserId,
		Username:      req.Username,
		SourceType:    req.SourceType,
		SourceId:      req.SourceId,
		TradeNo:       req.TradeNo,
		Amount:        req.Amount,
		BonusAmount:   req.BonusAmount,
		FlowMultiple:  flowMultiple,
		RequiredFlow:  requiredFlow,
		CompletedFlow: 0,
		RemainingFlow: requiredFlow,
		Status:        1,
		Progress:      0,
		StartTime:     now,
		ExpireTime:    expireTime,
		CompleteTime:  nil,
		Remark:        req.Remark,
		AdminId:       req.AdminId,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	id, err := dao.BalanceFlowRequirement.CreateFlowRequirement(ctx, record)
	if err != nil {
		return 0, nil, gerror.Wrap(err, "创建流水要求记录失败")
	}

	g.Log().Infof(ctx, "记录流水要求成功: id=%d, site_id=%d, user_id=%d, source_type=%d, amount=%.4f, bonus=%.4f, flow_multiple=%.2f, required_flow=%.4f", id, siteId, req.UserId, req.SourceType, req.Amount, req.BonusAmount, flowMultiple, requiredFlow)

	return id, record, nil
}

func mapSourceTypeToRuleType(sourceType int) int {
	switch sourceType {
	case balanceConsts.SourceTypeRecharge:
		return balanceConsts.RuleTypeRecharge
	case balanceConsts.SourceTypeBonus:
		return balanceConsts.RuleTypeBonus
	case balanceConsts.SourceTypeRebate:
		return balanceConsts.RuleTypeRebate
	case balanceConsts.SourceTypeManual:
		return balanceConsts.RuleTypeManual
	case balanceConsts.SourceTypeRegisterActivity:
		return balanceConsts.RuleTypeRegisterActivity
	case balanceConsts.SourceTypeRechargeActivity:
		return balanceConsts.RuleTypeRechargeActivity
	case balanceConsts.SourceTypeUpgradeActivity:
		return balanceConsts.RuleTypeUpgradeActivity
	case balanceConsts.SourceTypeDailySign:
		return balanceConsts.RuleTypeDailySign
	case balanceConsts.SourceTypeLuckySpin:
		return balanceConsts.RuleTypeLuckySpin
	default:
		return 0
	}
}
