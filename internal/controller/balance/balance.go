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

func (*Controller) GetChangeList(ctx context.Context, req *v1.GetChangeListReq) (res *v1.GetChangeListRes, err error) {
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

func (*Controller) QueryUserBalance(ctx context.Context, req *v1.QueryUserBalanceReq) (res *v1.QueryUserBalanceRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) QueryGameBalance(ctx context.Context, req *v1.QueryGameBalanceReq) (res *v1.QueryGameBalanceRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) ManualUserBalance(ctx context.Context, req *v1.ManualUserBalanceReq) (res *v1.ManualUserBalanceRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetPaymentAccounts(ctx context.Context, req *v1.GetPaymentAccountsReq) (res *v1.GetPaymentAccountsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) CreatePaymentAccount(ctx context.Context, req *v1.CreatePaymentAccountReq) (res *v1.CreatePaymentAccountRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetPaymentAccountUpdate(ctx context.Context, req *v1.GetPaymentAccountUpdateReq) (res *v1.GetPaymentAccountUpdateRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) UpdatePaymentAccount(ctx context.Context, req *v1.UpdatePaymentAccountReq) (res *v1.UpdatePaymentAccountRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) DeletePaymentAccount(ctx context.Context, req *v1.DeletePaymentAccountReq) (res *v1.DeletePaymentAccountRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetManualList(ctx context.Context, req *v1.GetManualListReq) (res *v1.GetManualListRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
