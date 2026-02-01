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
	AddBankCard(ctx context.Context, req *v1.AddBankCardReq) (res *v1.AddBankCardRes, err error)
	GetBankCardList(ctx context.Context, req *v1.GetBankCardListReq) (res *v1.GetBankCardListRes, err error)
	DeleteBankCard(ctx context.Context, req *v1.DeleteBankCardReq) (res *v1.DeleteBankCardRes, err error)
	SetDefaultBankCard(ctx context.Context, req *v1.SetDefaultBankCardReq) (res *v1.SetDefaultBankCardRes, err error)
	GetDepositNonce(ctx context.Context, req *v1.GetDepositNonceReq) (res *v1.GetDepositNonceRes, err error)
	GetWithdrawNonce(ctx context.Context, req *v1.GetWithdrawNonceReq) (res *v1.GetWithdrawNonceRes, err error)
	CreateDepositOrder(ctx context.Context, req *v1.CreateDepositOrderReq) (res *v1.CreateDepositOrderRes, err error)
	GetBalanceLog(ctx context.Context, req *v1.GetBalanceLogReq) (res *v1.GetBalanceLogRes, err error)
	DepositCallback(ctx context.Context, req *v1.DepositCallbackReq) (res *v1.DepositCallbackRes, err error)
	Withdraw(ctx context.Context, req *v1.WithdrawReq) (res *v1.WithdrawRes, err error)
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
