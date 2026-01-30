package frontend

import (
	"context"

	v1 "jh_balance_service/api/frontend/balance/v1"
)

type IBalance interface {
	InitBalanceUser(ctx context.Context, req *v1.InitBalanceUserReq) (res *v1.InitBalanceUserRes, err error)
	RefreshBalance(ctx context.Context, req *v1.RefreshBalanceReq) (res *v1.RefreshBalanceRes, err error)
	AdjustBalance(ctx context.Context, req *v1.AdjustBalanceReq) (res *v1.AdjustBalanceRes, err error)
	UpdateTransferStatus(ctx context.Context, req *v1.UpdateTransferStatusReq) (res *v1.UpdateTransferStatusRes, err error)
	CompensateBalance(ctx context.Context, req *v1.CompensateBalanceReq) (res *v1.CompensateBalanceRes, err error)
	GetLastBalanceLogGame(ctx context.Context, req *v1.GetLastBalanceLogGameReq) (res *v1.GetLastBalanceLogGameRes, err error)
	GetPaymentList(ctx context.Context, req *v1.GetPaymentListReq) (res *v1.GetPaymentListRes, err error)
	GetWithdrawList(ctx context.Context, req *v1.GetWithdrawListReq) (res *v1.GetWithdrawListRes, err error)
}

var localBalance IBalance

func Balance() IBalance {
	if localBalance == nil {
		panic("implement not found for interface IBalance, forgot register?")
	}
	return localBalance
}

func RegisterBalance(i IBalance) {
	localBalance = i
}
