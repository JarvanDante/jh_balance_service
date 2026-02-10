package deposit

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	logic_mq_outbox "jh_balance_service/internal/logic/mq_outbox"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/util"
)

// DepositDomain 充值领域服务
type DepositDomain struct{}

// CreateDepositOrderRequest 创建充值订单请求
type CreateDepositOrderRequest struct {
	SiteId       int64
	UserId       int64
	Username     string
	PaymentId    int64
	Money        float64
	OrderType    string
	BankCode     string
	Mobile       int64
	UserCardNo   string
	UserCardName string
}

// CreateDepositOrderResult 创建充值订单结果
type CreateDepositOrderResult struct {
	TradeNo       string
	SysTradeNo    string
	OrderSn       string
	ChannelCardNo string
	ChannelName   string
	ImageUrl      string
	Result        string
	SitePayment   *entity.SitePayment
}

// CreateDepositOrder 创建充值订单（核心业务逻辑）
func (d *DepositDomain) CreateDepositOrder(ctx context.Context, req *CreateDepositOrderRequest) (*CreateDepositOrderResult, error) {
	g.Log().Infof(ctx, "========== 充值领域服务：开始处理充值订单 ==========")
	g.Log().Infof(ctx, "创建充值订单: site_id=%d, user_id=%d, payment_id=%d, money=%.2f",
		req.SiteId, req.UserId, req.PaymentId, req.Money)

	// 1. 设置默认订单类型
	if req.OrderType == "" {
		req.OrderType = "deposit"
	}

	// 2. 查询支付渠道配置
	sitePayment, err := dao.SitePayment.GetByPaymentId(ctx, req.SiteId, req.PaymentId)
	if err != nil {
		return nil, gerror.Wrap(err, "查询支付渠道失败")
	}
	if sitePayment == nil {
		return nil, gerror.New("支付渠道不存在或已关闭")
	}

	// 3. 验证充值金额范围
	if err := d.validateDepositRange(req.Money, sitePayment.EachMin, sitePayment.EachMax); err != nil {
		return nil, err
	}

	// 4. 生成交易流水号
	tradeNo := util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_PAYMENT)
	g.Log().Infof(ctx, "生成交易流水号: %s", tradeNo)

	// 5 & 6. 创建充值订单记录 + 资金日志，并写入Outbox（同一事务）
	var (
		paymentLog *entity.BalanceLogPayment
		balanceLog *entity.BalanceLog
	)
	err = dao.BalanceLogPayment.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		paymentResult, err := dao.BalanceLogPayment.Ctx(ctx).TX(tx).Insert(&entity.BalanceLogPayment{
			SiteId:           int(req.SiteId),
			UserId:           int(req.UserId),
			Username:         req.Username,
			TradeNo:          tradeNo,
			Money:            req.Money,
			PaymentAccountId: int(req.PaymentId),
			Status:           balanceConsts.PaymentStatusUnpaid,
			CreatedAt:        gtime.Now(),
			UpdatedAt:        gtime.Now(),
		})
		if err != nil {
			return gerror.Wrap(err, "创建充值订单失败")
		}
		paymentId, err := paymentResult.LastInsertId()
		if err != nil {
			return gerror.Wrap(err, "获取充值订单ID失败")
		}

		paymentLog = &entity.BalanceLogPayment{
			Id:               uint64(paymentId),
			SiteId:           int(req.SiteId),
			UserId:           int(req.UserId),
			Username:         req.Username,
			TradeNo:          tradeNo,
			Money:            req.Money,
			PaymentAccountId: int(req.PaymentId),
			Status:           balanceConsts.PaymentStatusUnpaid,
			CreatedAt:        gtime.Now(),
			UpdatedAt:        gtime.Now(),
		}

		balanceResult, err := dao.BalanceLog.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":    req.SiteId,
			"user_id":    req.UserId,
			"trade_type": balanceConsts.LOG_TYPE_PAYMENT,
			"gateway":    sitePayment.Gateway,
			"trade_no":   tradeNo,
			"money":      req.Money,
			"status":     balanceConsts.StatusPending,
			"created_at": gtime.Now(),
			"updated_at": gtime.Now(),
		})
		if err != nil {
			return gerror.Wrap(err, "创建支付日志失败")
		}
		balanceLogId, err := balanceResult.LastInsertId()
		if err != nil {
			return gerror.Wrap(err, "获取支付日志ID失败")
		}

		balanceLog = &entity.BalanceLog{
			Id:        uint(balanceLogId),
			SiteId:    int(req.SiteId),
			UserId:    int(req.UserId),
			TradeType: balanceConsts.LOG_TYPE_PAYMENT,
			Gateway:   int(sitePayment.Gateway),
			TradeNo:   tradeNo,
			Money:     req.Money,
			Status:    balanceConsts.StatusPending,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}

		if err := logic_mq_outbox.MqOutbox().CreateBalanceLogPaymentOutbox(ctx, tx, "", paymentLog); err != nil {
			return gerror.Wrap(err, "创建充值订单Outbox失败")
		}
		if err := logic_mq_outbox.MqOutbox().CreateBalanceLogOutbox(ctx, tx, "", balanceLog); err != nil {
			return gerror.Wrap(err, "创建资金日志Outbox失败")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	g.Log().Infof(ctx, "充值订单创建成功: trade_no=%s", tradeNo)

	return &CreateDepositOrderResult{
		TradeNo:     tradeNo,
		SitePayment: sitePayment,
	}, nil
}

// validateDepositRange 验证充值金额范围
func (d *DepositDomain) validateDepositRange(amount, min, max float64) error {
	if min > 0 && amount < min {
		return gerror.Newf("最低充值金额为%.2f", min)
	}
	if max > 0 && amount > max {
		return gerror.Newf("最高充值金额为%.2f", max)
	}
	return nil
}

// UpdateDepositOrderInfo 更新充值订单信息
func (d *DepositDomain) UpdateDepositOrderInfo(ctx context.Context, tradeNo, sysTradeNo, channelCardNo, channelName, imageUrl string) error {
	if err := dao.BalanceLogPayment.UpdateDepositOrderInfo(ctx, tradeNo, sysTradeNo, channelCardNo, channelName, imageUrl); err != nil {
		return err
	}
	var updatedPayment entity.BalanceLogPayment
	if err := dao.BalanceLogPayment.Ctx(ctx).
		Where("trade_no", tradeNo).
		Scan(&updatedPayment); err != nil {
		return err
	}
	if err := logic_mq_outbox.MqOutbox().CreateBalanceLogPaymentOutbox(ctx, nil, "", &updatedPayment); err != nil {
		return err
	}
	return nil
}
