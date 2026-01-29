package backend

import (
	"context"
	v1 "jh_balance_service/api/backend/balance/v1"
)

type (
	IBalance interface {
		GetBalanceChanges(ctx context.Context, req *v1.GetBalanceChangesReq) (*v1.GetBalanceChangesRes, error)
		GetChangeList(ctx context.Context, req *v1.GetChangeListReq) (*v1.GetChangeListRes, error)
		GetRechargePayments(ctx context.Context, req *v1.GetRechargePaymentsReq) (*v1.GetRechargePaymentsRes, error)
		GetRechargeManuals(ctx context.Context, req *v1.GetRechargeManualsReq) (*v1.GetRechargeManualsRes, error)
		ConfirmPaymentOrder(ctx context.Context, req *v1.ConfirmPaymentOrderReq) (*v1.ConfirmPaymentOrderRes, error)
		GetWithdraws(ctx context.Context, req *v1.GetWithdrawsReq) (*v1.GetWithdrawsRes, error)
		GetWithdrawManuals(ctx context.Context, req *v1.GetWithdrawManualsReq) (*v1.GetWithdrawManualsRes, error)
		GetWithdrawReview(ctx context.Context, req *v1.GetWithdrawReviewReq) (*v1.GetWithdrawReviewRes, error)
		DealWithWithdraw(ctx context.Context, req *v1.DealWithWithdrawReq) (*v1.DealWithWithdrawRes, error)
		QueryUserBalance(ctx context.Context, req *v1.QueryUserBalanceReq) (*v1.QueryUserBalanceRes, error)
		QueryGameBalance(ctx context.Context, req *v1.QueryGameBalanceReq) (*v1.QueryGameBalanceRes, error)
		ManualUserBalance(ctx context.Context, req *v1.ManualUserBalanceReq) (*v1.ManualUserBalanceRes, error)
		GetSitePayments(ctx context.Context, req *v1.GetSitePaymentsReq) (*v1.GetSitePaymentsRes, error)
		UpdateSitePayment(ctx context.Context, req *v1.UpdateSitePaymentReq) (*v1.UpdateSitePaymentRes, error)
		GetSiteWithdraws(ctx context.Context, req *v1.GetSiteWithdrawsReq) (*v1.GetSiteWithdrawsRes, error)
		UpdateSiteWithdraw(ctx context.Context, req *v1.UpdateSiteWithdrawReq) (*v1.UpdateSiteWithdrawRes, error)
		GetManualList(ctx context.Context, req *v1.GetManualListReq) (*v1.GetManualListRes, error)
	}
)

var (
	localBalance IBalance
)

func Balance() IBalance {
	if localBalance == nil {
		panic("implement not found for interface IBalance, forgot register?")
	}
	return localBalance
}

func RegisterBalance(i IBalance) {
	localBalance = i
}
