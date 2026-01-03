package backend

import (
	"context"
	v1 "jh_balance_service/api/backend/balance/v1"
)

type (
	IBalance interface {
		GetBalanceChanges(ctx context.Context, req *v1.GetBalanceChangesReq) (*v1.GetBalanceChangesRes, error)
		GetRechargePayments(ctx context.Context, req *v1.GetRechargePaymentsReq) (*v1.GetRechargePaymentsRes, error)
		GetRechargeManuals(ctx context.Context, req *v1.GetRechargeManualsReq) (*v1.GetRechargeManualsRes, error)
		ConfirmPaymentOrder(ctx context.Context, req *v1.ConfirmPaymentOrderReq) (*v1.ConfirmPaymentOrderRes, error)
		GetWithdraws(ctx context.Context, req *v1.GetWithdrawsReq) (*v1.GetWithdrawsRes, error)
		GetWithdrawManuals(ctx context.Context, req *v1.GetWithdrawManualsReq) (*v1.GetWithdrawManualsRes, error)
		GetWithdrawReview(ctx context.Context, req *v1.GetWithdrawReviewReq) (*v1.GetWithdrawReviewRes, error)
		DealWithWithdraw(ctx context.Context, req *v1.DealWithWithdrawReq) (*v1.DealWithWithdrawRes, error)
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
