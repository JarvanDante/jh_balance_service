package balance

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "jh_balance_service/api/frontend/balance/v1"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/service/frontend"
)

type Controller struct {
	v1.UnimplementedBalanceServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterBalanceServer(s.Server, &Controller{})
}

// InitBalanceUser 初始化用户余额
func (c *Controller) InitBalanceUser(ctx context.Context, req *v1.InitBalanceUserReq) (res *v1.InitBalanceUserRes, err error) {
	g.Log().Infof(ctx, "Frontend Balance Controller收到InitBalanceUser请求")
	return frontend.Balance().InitBalanceUser(ctx, req)
}

// RefreshBalance 刷新用户余额
func (c *Controller) RefreshBalance(ctx context.Context, req *v1.RefreshBalanceReq) (res *v1.RefreshBalanceRes, err error) {
	return frontend.Balance().RefreshBalance(ctx, req)
}

// AdjustBalance 调整用户余额
func (c *Controller) AdjustBalance(ctx context.Context, req *v1.AdjustBalanceReq) (res *v1.AdjustBalanceRes, err error) {
	return frontend.Balance().AdjustBalance(ctx, req)
}

// UpdateTransferStatus 更新转账状态
func (c *Controller) UpdateTransferStatus(ctx context.Context, req *v1.UpdateTransferStatusReq) (res *v1.UpdateTransferStatusRes, err error) {
	return frontend.Balance().UpdateTransferStatus(ctx, req)
}

// CompensateBalance 补偿余额
func (c *Controller) CompensateBalance(ctx context.Context, req *v1.CompensateBalanceReq) (res *v1.CompensateBalanceRes, err error) {
	return frontend.Balance().CompensateBalance(ctx, req)
}

// GetLastBalanceLogGame 获取最后一次进入游戏
func (c *Controller) GetLastBalanceLogGame(ctx context.Context, req *v1.GetLastBalanceLogGameReq) (res *v1.GetLastBalanceLogGameRes, err error) {
	return frontend.Balance().GetLastBalanceLogGame(ctx, req)
}

// GetPaymentList 获取支付渠道列表
func (c *Controller) GetPaymentList(ctx context.Context, req *v1.GetPaymentListReq) (res *v1.GetPaymentListRes, err error) {
	return frontend.Balance().GetPaymentList(ctx, req)
}

// GetWithdrawList 获取提现渠道列表
func (c *Controller) GetWithdrawList(ctx context.Context, req *v1.GetWithdrawListReq) (res *v1.GetWithdrawListRes, err error) {
	return frontend.Balance().GetWithdrawList(ctx, req)
}

// AddBankCard 添加银行卡
func (c *Controller) AddBankCard(ctx context.Context, req *v1.AddBankCardReq) (res *v1.AddBankCardRes, err error) {
	return frontend.Balance().AddBankCard(ctx, req)
}

// GetBankCardList 查询银行卡列表
func (c *Controller) GetBankCardList(ctx context.Context, req *v1.GetBankCardListReq) (res *v1.GetBankCardListRes, err error) {
	return frontend.Balance().GetBankCardList(ctx, req)
}

// DeleteBankCard 删除银行卡
func (c *Controller) DeleteBankCard(ctx context.Context, req *v1.DeleteBankCardReq) (res *v1.DeleteBankCardRes, err error) {
	return frontend.Balance().DeleteBankCard(ctx, req)
}

// SetDefaultBankCard 设置默认银行卡
func (c *Controller) SetDefaultBankCard(ctx context.Context, req *v1.SetDefaultBankCardReq) (res *v1.SetDefaultBankCardRes, err error) {
	return frontend.Balance().SetDefaultBankCard(ctx, req)
}

// GetDepositNonce 获取充值 nonce
func (c *Controller) GetDepositNonce(ctx context.Context, req *v1.GetDepositNonceReq) (res *v1.GetDepositNonceRes, err error) {
	return frontend.Balance().GetDepositNonce(ctx, req)
}

// GetWithdrawNonce 获取提现 nonce
func (c *Controller) GetWithdrawNonce(ctx context.Context, req *v1.GetWithdrawNonceReq) (res *v1.GetWithdrawNonceRes, err error) {
	return frontend.Balance().GetWithdrawNonce(ctx, req)
}

// CreateDepositOrder 创建充值订单
func (c *Controller) CreateDepositOrder(ctx context.Context, req *v1.CreateDepositOrderReq) (res *v1.CreateDepositOrderRes, err error) {
	// 参数验证
	if req.PaymentId <= 0 {
		g.Log().Warning(ctx, "支付渠道ID无效")
		return nil, gerror.New("请选择支付渠道")
	}

	if req.Money < 0.01 {
		g.Log().Warning(ctx, "充值金额无效")
		return nil, gerror.New("充值金额最小为0.01")
	}

	if req.Nonce == "" {
		g.Log().Warning(ctx, "nonce不能为空")
		return nil, gerror.New("请求参数无效，请刷新页面重试")
	}

	return frontend.Balance().CreateDepositOrder(ctx, req)
}

// GetBalanceLog 获取账户明细
func (c *Controller) GetBalanceLog(ctx context.Context, req *v1.GetBalanceLogReq) (res *v1.GetBalanceLogRes, err error) {
	return frontend.Balance().GetBalanceLog(ctx, req)
}

// GetUserFlowRequirements 获取用户流水要求列表
func (c *Controller) GetUserFlowRequirements(ctx context.Context, req *v1.GetUserFlowRequirementsReq) (res *v1.GetUserFlowRequirementsRes, err error) {
	return frontend.Balance().GetUserFlowRequirements(ctx, req)
}

// DepositCallback 充值异步回调
func (c *Controller) DepositCallback(ctx context.Context, req *v1.DepositCallbackReq) (res *v1.DepositCallbackRes, err error) {
	g.Log().Infof(ctx, "Frontend Balance Controller收到DepositCallback请求")
	return frontend.Balance().DepositCallback(ctx, req)
}

// Withdraw 申请提现
func (c *Controller) Withdraw(ctx context.Context, req *v1.WithdrawReq) (res *v1.WithdrawRes, err error) {
	// 参数验证
	if req.WithdrawId <= 0 {
		g.Log().Warning(ctx, "提现渠道ID无效")
		return &v1.WithdrawRes{
			Success: false,
			Message: "请选择提现渠道",
		}, nil
	}

	if req.Money <= 0 {
		g.Log().Warning(ctx, "提现金额无效")
		return &v1.WithdrawRes{
			Success: false,
			Message: "提现金额必须大于0",
		}, nil
	}

	if req.Nonce == "" {
		g.Log().Warning(ctx, "nonce不能为空")
		return &v1.WithdrawRes{
			Success: false,
			Message: "请求参数无效，请刷新页面重试",
		}, nil
	}

	// 检查是否存在未完成的流水要求
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		g.Log().Warning(ctx, "未获取用户ID")
		return &v1.WithdrawRes{
			Success: false,
			Message: "未登录",
		}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		g.Log().Warning(ctx, "未获取站点ID")
		return &v1.WithdrawRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}

	flowAgg := struct {
		RemainingFlow float64 `json:"remaining_flow"`
	}{}
	err = dao.BalanceFlowRequirement.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", 1).
		Where("remaining_flow > ?", 0).
		Fields("IFNULL(SUM(remaining_flow),0) as remaining_flow").
		Scan(&flowAgg)
	if err != nil {
		g.Log().Errorf(ctx, "查询流水要求失败: %v", err)
		return &v1.WithdrawRes{
			Success: false,
			Message: "系统繁忙，请稍后再试",
		}, nil
	}

	if flowAgg.RemainingFlow > 0 {
		return &v1.WithdrawRes{
			Success: false,
			Message: "您还有未完成的流水要求，暂不可提现",
		}, nil
	}

	return frontend.Balance().Withdraw(ctx, req)
}
