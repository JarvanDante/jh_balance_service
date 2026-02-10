package withdraw

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	logic_mq_outbox "jh_balance_service/internal/logic/mq_outbox"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/util"
)

// WithdrawDomain 提现领域服务
type WithdrawDomain struct{}

// ApplyWithdrawRequest 申请提现请求
type ApplyWithdrawRequest struct {
	SiteId     int64
	UserId     int64
	Username   string
	Money      float64
	WithdrawId int64
	BankCardId int64
}

// ApplyWithdrawResult 申请提现结果
type ApplyWithdrawResult struct {
	TradeNo string
	Message string
}

// ApplyWithdraw 申请提现（核心业务逻辑）
func (d *WithdrawDomain) ApplyWithdraw(ctx context.Context, req *ApplyWithdrawRequest) (*ApplyWithdrawResult, error) {
	g.Log().Infof(ctx, "========== 提现领域服务：开始处理提现申请 ==========")
	g.Log().Infof(ctx, "申请提现: site_id=%d, user_id=%d, money=%.2f, withdraw_id=%d",
		req.SiteId, req.UserId, req.Money, req.WithdrawId)

	// 1. 业务规则验证
	if err := d.validateWithdrawAmount(req.Money); err != nil {
		return nil, err
	}

	// 2. 查询用户银行卡
	userBank, err := dao.UserBank.GetDefaultBankCard(ctx, req.SiteId, req.UserId)
	if err != nil {
		return nil, gerror.Wrap(err, "查询银行卡失败")
	}
	if userBank == nil {
		return nil, gerror.New("请先绑定银行卡")
	}

	// 3. 查询提现渠道
	siteWithdraw, err := dao.SiteWithdraw.GetByWithdrawId(ctx, req.SiteId, req.WithdrawId)
	if err != nil {
		return nil, gerror.Wrap(err, "查询提现渠道失败")
	}
	if siteWithdraw == nil {
		return nil, gerror.New("提现渠道不存在或已关闭")
	}

	// 4. 验证提现金额范围
	if err := d.validateWithdrawRange(req.Money, siteWithdraw.EachMin, siteWithdraw.EachMax); err != nil {
		return nil, err
	}

	// 5. 检查是否有待审核订单
	hasPending, err := dao.BalanceLogWithdraw.HasPendingOrder(ctx, req.SiteId, req.UserId)
	if err != nil {
		return nil, gerror.Wrap(err, "查询待审核订单失败")
	}
	if hasPending {
		return nil, gerror.New("提现申请还在处理中，处理完成即可再次提现")
	}

	// 6. 生成交易流水号
	tradeNo := util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_WITHDRAW)

	// 7. 执行提现事务
	err = d.executeWithdrawTransaction(ctx, req, tradeNo, userBank, siteWithdraw)
	if err != nil {
		return nil, err
	}

	g.Log().Infof(ctx, "申请提现成功: trade_no=%s, money=%.2f", tradeNo, req.Money)

	return &ApplyWithdrawResult{
		TradeNo: tradeNo,
		Message: "申请提现成功",
	}, nil
}

// validateWithdrawAmount 验证提现金额
func (d *WithdrawDomain) validateWithdrawAmount(amount float64) error {
	if amount <= 0 {
		return gerror.New("提现金额必须大于0")
	}
	return nil
}

// validateWithdrawRange 验证提现金额范围
func (d *WithdrawDomain) validateWithdrawRange(amount, min, max float64) error {
	if amount > max {
		return gerror.Newf("提现最高金额：%.2f", max)
	}
	if amount < min {
		return gerror.Newf("提现最低金额：%.2f", min)
	}
	return nil
}

// executeWithdrawTransaction 执行提现事务
func (d *WithdrawDomain) executeWithdrawTransaction(
	ctx context.Context,
	req *ApplyWithdrawRequest,
	tradeNo string,
	userBank *entity.UserBank,
	siteWithdraw *entity.SiteWithdraw,
) error {
	return dao.BalanceUser.Transaction(ctx, func(ctx context.Context) error {
		// 1. 查询并锁定用户余额
		balanceUser, err := dao.BalanceUser.GetByUserIdWithLock(ctx, req.SiteId, req.UserId)
		if err != nil {
			return gerror.Wrap(err, "查询用户余额失败")
		}
		if balanceUser == nil {
			return gerror.New("用户余额记录不存在")
		}

		// 2. 检查余额是否足够
		if err := d.checkBalance(balanceUser.Balance, req.Money); err != nil {
			return err
		}

		// 3. 计算新余额
		newBalance := balanceUser.Balance - req.Money
		if newBalance < 0 {
			newBalance = 0
		}
		newBalanceFrozen := balanceUser.BalanceFrozen + req.Money

		// 4. 冻结余额
		if err := dao.BalanceUser.FreezeBalance(ctx, req.SiteId, req.UserId, req.Money); err != nil {
			return gerror.Wrap(err, "冻结余额失败")
		}

		// 5. 创建提现订单
		if err := dao.BalanceLogWithdraw.CreateWithdrawOrder(ctx, &dao.CreateWithdrawOrderParams{
			SiteId:            req.SiteId,
			UserId:            req.UserId,
			Username:          req.Username,
			TradeNo:           tradeNo,
			Money:             req.Money,
			CardAccount:       userBank.CardAccount,
			CardNo:            userBank.CardNo,
			Channel:           siteWithdraw.Name,
			WithdrawAccountId: req.WithdrawId,
		}); err != nil {
			return gerror.Wrap(err, "创建提现订单失败")
		}

		// 6. 记录余额日志
		if _, err := dao.BalanceLog.CreateWithdrawLog(ctx, &dao.CreateWithdrawLogParams{
			SiteId:  req.SiteId,
			UserId:  req.UserId,
			TradeNo: tradeNo,
			Money:   req.Money,
		}); err != nil {
			return gerror.Wrap(err, "记录余额日志失败")
		}
		var balanceLog entity.BalanceLog
		if err := dao.BalanceLog.Ctx(ctx).
			Where("trade_no", tradeNo).
			Scan(&balanceLog); err != nil {
			return gerror.Wrap(err, "查询余额日志失败")
		}
		if err := logic_mq_outbox.MqOutbox().CreateBalanceLogOutbox(ctx, nil, "", &balanceLog); err != nil {
			return gerror.Wrap(err, "创建余额日志Outbox失败")
		}

		// 7. 记录账变
		if err := dao.BalanceChange.CreateWithdrawChange(ctx, &dao.CreateWithdrawChangeParams{
			SiteId:        req.SiteId,
			UserId:        req.UserId,
			Username:      req.Username,
			TradeNo:       tradeNo,
			BalanceOld:    balanceUser.Balance,
			Money:         req.Money,
			BalanceNew:    newBalance,
			BalanceFrozen: newBalanceFrozen,
		}); err != nil {
			return gerror.Wrap(err, "记录账变失败")
		}

		return nil
	})
}

// checkBalance 检查余额是否足够
func (d *WithdrawDomain) checkBalance(balance, needAmount float64) error {
	// 四舍五入到2位小数后比较，避免浮点数精度问题
	balanceRounded := float64(int(balance*100+0.5)) / 100
	moneyRounded := float64(int(needAmount*100+0.5)) / 100

	g.Log().Infof(context.Background(), "余额检查: 当前余额=%.2f, 需要=%.2f", balance, needAmount)

	if balanceRounded < moneyRounded {
		return gerror.Newf("余额不足，当前余额: %.2f, 需要: %.2f", balance, needAmount)
	}

	return nil
}
