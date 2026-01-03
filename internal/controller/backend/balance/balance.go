package balance

import (
	"context"
	v1 "jh_balance_service/api/backend/balance/v1"
	"jh_balance_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedBalanceServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterBalanceServer(s.Server, &Controller{})
}

// GetBalanceChanges 获取账变记录
func (*Controller) GetBalanceChanges(ctx context.Context, req *v1.GetBalanceChangesReq) (res *v1.GetBalanceChangesRes, err error) {
	return backend.Balance().GetBalanceChanges(ctx, req)
}

// GetRechargePayments 获取充值记录
func (*Controller) GetRechargePayments(ctx context.Context, req *v1.GetRechargePaymentsReq) (res *v1.GetRechargePaymentsRes, err error) {
	return backend.Balance().GetRechargePayments(ctx, req)
}

// GetRechargeManuals 获取后台加款记录
func (*Controller) GetRechargeManuals(ctx context.Context, req *v1.GetRechargeManualsReq) (res *v1.GetRechargeManualsRes, err error) {
	return backend.Balance().GetRechargeManuals(ctx, req)
}

// ConfirmPaymentOrder 确认支付订单
func (*Controller) ConfirmPaymentOrder(ctx context.Context, req *v1.ConfirmPaymentOrderReq) (res *v1.ConfirmPaymentOrderRes, err error) {
	return backend.Balance().ConfirmPaymentOrder(ctx, req)
}

// GetWithdraws 获取提现记录
func (*Controller) GetWithdraws(ctx context.Context, req *v1.GetWithdrawsReq) (res *v1.GetWithdrawsRes, err error) {
	return backend.Balance().GetWithdraws(ctx, req)
}

// GetWithdrawManuals 获取后台提现记录
func (*Controller) GetWithdrawManuals(ctx context.Context, req *v1.GetWithdrawManualsReq) (res *v1.GetWithdrawManualsRes, err error) {
	return backend.Balance().GetWithdrawManuals(ctx, req)
}

// GetWithdrawReview 获取提现审核信息
func (*Controller) GetWithdrawReview(ctx context.Context, req *v1.GetWithdrawReviewReq) (res *v1.GetWithdrawReviewRes, err error) {
	return backend.Balance().GetWithdrawReview(ctx, req)
}

// QueryUserBalance 查询用户余额
func (*Controller) QueryUserBalance(ctx context.Context, req *v1.QueryUserBalanceReq) (res *v1.QueryUserBalanceRes, err error) {
	return backend.Balance().QueryUserBalance(ctx, req)
}

// QueryGameBalance 查询游戏余额
func (*Controller) QueryGameBalance(ctx context.Context, req *v1.QueryGameBalanceReq) (res *v1.QueryGameBalanceRes, err error) {
	return backend.Balance().QueryGameBalance(ctx, req)
}

// ManualUserBalance 手动操作用户余额
func (*Controller) ManualUserBalance(ctx context.Context, req *v1.ManualUserBalanceReq) (res *v1.ManualUserBalanceRes, err error) {
	return backend.Balance().ManualUserBalance(ctx, req)
}

// GetPaymentAccounts 获取支付接口列表
func (*Controller) GetPaymentAccounts(ctx context.Context, req *v1.GetPaymentAccountsReq) (res *v1.GetPaymentAccountsRes, err error) {
	return backend.Balance().GetPaymentAccounts(ctx, req)
}

// CreatePaymentAccount 创建支付接口
func (*Controller) CreatePaymentAccount(ctx context.Context, req *v1.CreatePaymentAccountReq) (res *v1.CreatePaymentAccountRes, err error) {
	return backend.Balance().CreatePaymentAccount(ctx, req)
}

// GetPaymentAccountUpdate 获取支付接口编辑信息
func (*Controller) GetPaymentAccountUpdate(ctx context.Context, req *v1.GetPaymentAccountUpdateReq) (res *v1.GetPaymentAccountUpdateRes, err error) {
	return backend.Balance().GetPaymentAccountUpdate(ctx, req)
}

// UpdatePaymentAccount 更新支付接口
func (*Controller) UpdatePaymentAccount(ctx context.Context, req *v1.UpdatePaymentAccountReq) (res *v1.UpdatePaymentAccountRes, err error) {
	return backend.Balance().UpdatePaymentAccount(ctx, req)
}

// DeletePaymentAccount 删除支付接口
func (*Controller) DeletePaymentAccount(ctx context.Context, req *v1.DeletePaymentAccountReq) (res *v1.DeletePaymentAccountRes, err error) {
	return backend.Balance().DeletePaymentAccount(ctx, req)
}
