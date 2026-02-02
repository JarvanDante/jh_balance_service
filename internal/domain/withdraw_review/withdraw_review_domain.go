package withdraw_review

import (
	"context"
	"fmt"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/model/entity"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WithdrawReviewDomain 提现审核领域服务
type WithdrawReviewDomain struct {
	ctx context.Context
}

// New 创建提现审核领域服务实例
func New(ctx context.Context) *WithdrawReviewDomain {
	return &WithdrawReviewDomain{
		ctx: ctx,
	}
}

// ReviewType 审核类型
type ReviewType int32

const (
	ReviewTypeReject     ReviewType = 0 // 拒绝
	ReviewTypeApprove    ReviewType = 1 // 通过
	ReviewTypeCompensate ReviewType = 2 // 补单
)

// ReviewRequest 审核请求参数
type ReviewRequest struct {
	WithdrawID int64      // 提现记录ID
	Type       ReviewType // 审核类型
	Fee        float64    // 手续费
	Remark     string     // 备注
}

// ReviewResult 审核结果
type ReviewResult struct {
	Success bool
	Message string
}

// ProcessWithdrawReview 处理提现审核
func (d *WithdrawReviewDomain) ProcessWithdrawReview(tx gdb.TX, req *ReviewRequest, withdraw *entity.BalanceLogWithdraw) (*ReviewResult, error) {
	// 1. 先将订单状态改为 Processing，防止并发处理
	rowsAffected, err := d.lockWithdrawOrder(tx, req.WithdrawID)
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return &ReviewResult{
			Success: false,
			Message: "订单已被其他请求处理，请勿重复操作",
		}, nil
	}

	g.Log().Infof(d.ctx, "提现订单开始处理: id=%d, trade_no=%s, status=Processing", req.WithdrawID, withdraw.TradeNo)

	// 2. 确定最终状态和消息
	finalStatus, message := d.determineFinalStatus(req.Type)

	// 3. 查询用户余额
	balanceUser, err := dao.BalanceUser.GetByUserIdWithTx(d.ctx, tx, int64(withdraw.SiteId), int64(withdraw.UserId))
	if err != nil {
		return nil, fmt.Errorf("查找用户余额记录失败: %v", err)
	}

	// 4. 更新提现记录
	err = dao.BalanceLogWithdraw.UpdateStatusWithTx(d.ctx, tx, req.WithdrawID, finalStatus, req.Fee, message)
	if err != nil {
		return nil, fmt.Errorf("更新提现记录失败: %v", err)
	}

	// 5. 更新日志记录
	err = dao.BalanceLog.UpdateByTradeNoWithTx(d.ctx, tx, withdraw.TradeNo, finalStatus, withdraw.Fee, message)
	if err != nil {
		return nil, fmt.Errorf("更新日志记录失败: %v", err)
	}

	// 6. 根据最终状态处理余额
	if finalStatus == balanceConsts.StatusSuccess {
		err = d.handleWithdrawSuccess(tx, withdraw, balanceUser, message)
	} else if finalStatus == balanceConsts.StatusFail {
		err = d.handleWithdrawFail(tx, withdraw, balanceUser, message)
	}

	if err != nil {
		return nil, err
	}

	return &ReviewResult{
		Success: true,
		Message: message,
	}, nil
}

// lockWithdrawOrder 锁定提现订单，将状态改为 Processing
func (d *WithdrawReviewDomain) lockWithdrawOrder(tx gdb.TX, withdrawID int64) (int64, error) {
	result, err := dao.BalanceLogWithdraw.Ctx(d.ctx).TX(tx).
		Where("id = ?", withdrawID).
		Where("status = ?", balanceConsts.StatusPending).
		Update(g.Map{
			"status":     balanceConsts.StatusProcessing,
			"updated_at": time.Now(),
		})
	if err != nil {
		return 0, fmt.Errorf("更新订单为处理中状态失败: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}

	return rowsAffected, nil
}

// determineFinalStatus 根据审核类型确定最终状态和消息
func (d *WithdrawReviewDomain) determineFinalStatus(reviewType ReviewType) (int, string) {
	switch reviewType {
	case ReviewTypeApprove:
		return balanceConsts.StatusSuccess, "提现确认成功"
	case ReviewTypeReject:
		return balanceConsts.StatusFail, "提现已拒绝"
	case ReviewTypeCompensate:
		return balanceConsts.StatusSuccess, "补单提现成功"
	default:
		return balanceConsts.StatusFail, "无效的处理类型"
	}
}

// handleWithdrawSuccess 处理提现成功的情况
func (d *WithdrawReviewDomain) handleWithdrawSuccess(tx gdb.TX, withdraw *entity.BalanceLogWithdraw, balanceUser *entity.BalanceUser, message string) error {
	// 计算新余额
	newBalance := balanceUser.Balance
	newBalanceFrozen := balanceUser.BalanceFrozen - withdraw.Money

	// 更新账变记录
	err := dao.BalanceChange.UpdateByTradeNoWithTx(d.ctx, tx, withdraw.TradeNo, balanceConsts.TradeTypeWithdrawSuccess, balanceConsts.StatusSuccess, message, newBalance, newBalanceFrozen)
	if err != nil {
		return fmt.Errorf("更新账变记录失败: %v", err)
	}

	// 减少冻结金额
	err = dao.BalanceUser.UpdateFrozenBalanceWithTx(d.ctx, tx, int64(withdraw.SiteId), int64(withdraw.UserId), newBalanceFrozen)
	if err != nil {
		return fmt.Errorf("减少冻结金额失败: %v", err)
	}

	g.Log().Infof(d.ctx, "提现成功，减少用户冻结金额: user_id=%d, amount=%.2f", withdraw.UserId, withdraw.Money)
	return nil
}

// handleWithdrawFail 处理提现失败的情况
func (d *WithdrawReviewDomain) handleWithdrawFail(tx gdb.TX, withdraw *entity.BalanceLogWithdraw, balanceUser *entity.BalanceUser, message string) error {
	// 计算新余额：从冻结金额转回可用余额
	newBalance := balanceUser.Balance + withdraw.Money
	newBalanceFrozen := balanceUser.BalanceFrozen - withdraw.Money

	// 更新账变记录
	err := dao.BalanceChange.UpdateByTradeNoWithTx(d.ctx, tx, withdraw.TradeNo, balanceConsts.TradeTypeWithdrawRefuse, balanceConsts.StatusFail, message, newBalance, newBalanceFrozen)
	if err != nil {
		return fmt.Errorf("更新账变记录失败: %v", err)
	}

	// 解冻余额
	err = dao.BalanceUser.UpdateBalanceWithTx(d.ctx, tx, int64(withdraw.SiteId), int64(withdraw.UserId), newBalance, newBalanceFrozen)
	if err != nil {
		return fmt.Errorf("解冻余额失败: %v", err)
	}

	g.Log().Infof(d.ctx, "提现失败，解冻用户余额: user_id=%d, amount=%.2f", withdraw.UserId, withdraw.Money)
	return nil
}
