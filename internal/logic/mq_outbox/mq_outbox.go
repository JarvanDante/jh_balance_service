package mq_outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/do"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/mq"
	"jh_balance_service/internal/registry"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
)

type sMqOutbox struct{}

var insOutbox = sMqOutbox{}

func MqOutbox() *sMqOutbox {
	return &insOutbox
}

// OutboxStatus 定义 Outbox 状态常量
const (
	OutboxStatusPending = 0 // 待发送
	OutboxStatusSent    = 1 // 已发送
	OutboxStatusFailed  = 2 // 发送失败
)

// CreateBalanceLogOutbox 创建同步 balance_log 到 ClickHouse 的 Outbox 记录
func (s *sMqOutbox) CreateBalanceLogOutbox(ctx context.Context, tx gdb.TX, siteCode string, log *entity.BalanceLog) error {
	if log == nil || log.Id == 0 {
		return fmt.Errorf("balance_log 为空")
	}
	if siteCode == "" {
		siteCode = resolveSiteCode(ctx, int64(log.SiteId))
	}

	outbox := &do.MqOutbox{
		EventType:     "balance_log_sync",
		AggregateType: "balance_log",
		AggregateId:   fmt.Sprintf("%d", log.Id),
		SiteCode:      siteCode,
		Exchange:      "analytics.sync",
		RoutingKey:    "analytics.balance_log",
		Payload:       "{}",
		Status:        OutboxStatusPending,
		RetryCount:    0,
		NextRetryAt:   gtime.Now(),
	}

	return insertOutbox(ctx, tx, outbox, func(eventID uint64) (string, error) {
		msg := &mq.SyncBalanceLogToClickHouseMessage{
			EventID:     eventID,
			SiteID:      uint32(log.SiteId),
			ID:          uint64(log.Id),
			UserID:      uint64(log.UserId),
			TradeType:   uint8(log.TradeType),
			Gateway:     uint8(log.Gateway),
			BankType:    uint8(log.BankType),
			GameID:      uint8(log.GameId),
			TradeNo:     log.TradeNo,
			Money:       log.Money,
			ActualMoney: log.ActualMoney,
			Fee:         log.Fee,
			Status:      uint8(log.Status),
			AdminID:     uint32(log.AdminId),
			Remark:      log.Remark,
			CreatedAt:   formatGTime(log.CreatedAt),
			UpdatedAt:   formatGTime(log.UpdatedAt),
		}
		payload, err := json.Marshal(msg)
		if err != nil {
			return "", fmt.Errorf("序列化消息失败: %w", err)
		}
		return string(payload), nil
	})
}

// CreateBalanceLogPaymentOutbox 创建同步 balance_log_payment 到 ClickHouse 的 Outbox 记录
func (s *sMqOutbox) CreateBalanceLogPaymentOutbox(ctx context.Context, tx gdb.TX, siteCode string, log *entity.BalanceLogPayment) error {
	if log == nil || log.Id == 0 {
		return fmt.Errorf("balance_log_payment 为空")
	}
	if siteCode == "" {
		siteCode = resolveSiteCode(ctx, int64(log.SiteId))
	}

	outbox := &do.MqOutbox{
		EventType:     "balance_log_payment_sync",
		AggregateType: "balance_log_payment",
		AggregateId:   fmt.Sprintf("%d", log.Id),
		SiteCode:      siteCode,
		Exchange:      "analytics.sync",
		RoutingKey:    "analytics.balance_log_payment",
		Payload:       "{}",
		Status:        OutboxStatusPending,
		RetryCount:    0,
		NextRetryAt:   gtime.Now(),
	}

	return insertOutbox(ctx, tx, outbox, func(eventID uint64) (string, error) {
		msg := &mq.SyncBalanceLogPaymentToClickHouseMessage{
			EventID:            eventID,
			SiteID:             uint32(log.SiteId),
			ID:                 uint64(log.Id),
			UserID:             uint64(log.UserId),
			Username:           log.Username,
			ActivityRechargeID: uint32(log.ActivityRechargeId),
			Gateway:            uint8(log.Gateway),
			PaymentID:          uint32(log.PaymentId),
			PaymentAccountID:   uint32(log.PaymentAccountId),
			BankValue:          log.BankValue,
			TradeNo:            log.TradeNo,
			Money:              log.Money,
			ActualMoney:        log.ActualMoney,
			Fee:                log.Fee,
			Status:             uint8(log.Status),
			AdminID:            uint32(log.AdminId),
			Remark:             log.Remark,
			CreatedAt:          formatGTime(log.CreatedAt),
			UpdatedAt:          formatGTime(log.UpdatedAt),
		}
		payload, err := json.Marshal(msg)
		if err != nil {
			return "", fmt.Errorf("序列化消息失败: %w", err)
		}
		return string(payload), nil
	})
}

func insertOutbox(ctx context.Context, tx gdb.TX, outbox *do.MqOutbox, buildPayload func(eventID uint64) (string, error)) error {
	if outbox == nil {
		return fmt.Errorf("outbox不能为空")
	}

	model := dao.MqOutbox.Ctx(ctx)
	if tx != nil {
		model = model.TX(tx)
	}

	result, err := model.Insert(outbox)
	if err != nil {
		return err
	}

	outboxID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取Outbox ID失败: %w", err)
	}

	payload, err := buildPayload(uint64(outboxID))
	if err != nil {
		return err
	}

	updateModel := dao.MqOutbox.Ctx(ctx)
	if tx != nil {
		updateModel = updateModel.TX(tx)
	}

	_, err = updateModel.Where("id", outboxID).Update(do.MqOutbox{
		Payload:   payload,
		UpdatedAt: gtime.Now(),
	})
	return err
}

func resolveSiteCode(ctx context.Context, siteId int64) string {
	if siteCode, ok := middleware.GetSiteCodeFromContext(ctx); ok && siteCode != "" {
		return siteCode
	}
	return registry.GetDBConfigManager().GetSiteCodeBySiteID(siteId)
}

func formatGTime(t *gtime.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("Y-m-d H:i:s")
}
