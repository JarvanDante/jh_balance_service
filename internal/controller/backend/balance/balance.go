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

// DealWithWithdraw 处理提现
func (*Controller) DealWithWithdraw(ctx context.Context, req *v1.DealWithWithdrawReq) (res *v1.DealWithWithdrawRes, err error) {
	return backend.Balance().DealWithWithdraw(ctx, req)
}
