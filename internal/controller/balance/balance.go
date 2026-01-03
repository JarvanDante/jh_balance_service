package balance

import (
	"context"
	v1 "jh_balance_service/api/backend/balance/v1"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct {
	v1.UnimplementedBalanceServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterBalanceServer(s.Server, &Controller{})
}

func (*Controller) GetBalanceChanges(ctx context.Context, req *v1.GetBalanceChangesReq) (res *v1.GetBalanceChangesRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetRechargePayments(ctx context.Context, req *v1.GetRechargePaymentsReq) (res *v1.GetRechargePaymentsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetRechargeManuals(ctx context.Context, req *v1.GetRechargeManualsReq) (res *v1.GetRechargeManualsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) ConfirmPaymentOrder(ctx context.Context, req *v1.ConfirmPaymentOrderReq) (res *v1.ConfirmPaymentOrderRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetWithdraws(ctx context.Context, req *v1.GetWithdrawsReq) (res *v1.GetWithdrawsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetWithdrawManuals(ctx context.Context, req *v1.GetWithdrawManualsReq) (res *v1.GetWithdrawManualsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetWithdrawReview(ctx context.Context, req *v1.GetWithdrawReviewReq) (res *v1.GetWithdrawReviewRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) DealWithWithdraw(ctx context.Context, req *v1.DealWithWithdrawReq) (res *v1.DealWithWithdrawRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
