package util

import (
	"context"
	"fmt"
	"strings"

	messagev1 "jh_balance_service/api/grpc/message/backend/v1"
	"jh_balance_service/internal/grpc_client"

	"github.com/gogf/gf/v2/util/gconv"
	"google.golang.org/grpc/metadata"
)

const (
	MessageEventRegisterSystem = "REGISTER_SYSTEM"
	MessageEventDepositSystem  = "DEPOSIT_SYSTEM"
	MessageEventWithdrawSystem = "WITHDRAW_SYSTEM"
	MessageEventVipSystem      = "VIP_SYSTEM"

	MessageEventRegisterActivity = "REGISTER_ACTIVITY"
	MessageEventDepositActivity  = "DEPOSIT_ACTIVITY"
	MessageEventInviteActivity   = "INVITE_ACTIVITY"
	MessageEventSigninActivity   = "SIGNIN_ACTIVITY"
	MessageEventLuckyActivity    = "LUCKY_ACTIVITY"
)

// SendUserMessage 发送用户消息(使用app_service模板)
func SendUserMessage(ctx context.Context, siteId int, userId int, eventCode string, bizId string, amount float64, gradeName string, vars map[string]string) error {
	if siteId <= 0 || userId <= 0 {
		return fmt.Errorf("invalid site_id or user_id")
	}
	if strings.TrimSpace(eventCode) == "" {
		return fmt.Errorf("event_code is empty")
	}
	if strings.TrimSpace(bizId) == "" {
		return fmt.Errorf("biz_id is empty")
	}

	req := &messagev1.SendUserMessageReq{
		SiteId:    int32(siteId),
		UserId:    int32(userId),
		EventCode: strings.TrimSpace(eventCode),
		BizId:     strings.TrimSpace(bizId),
	}
	if amount > 0 {
		req.Amount = amount
	}
	if strings.TrimSpace(gradeName) != "" {
		req.GradeName = strings.TrimSpace(gradeName)
	}
	if len(vars) > 0 {
		req.Vars = vars
	}

	ctxWithMeta := metadata.AppendToOutgoingContext(ctx, "site_id", gconv.String(siteId))
	res, err := grpc_client.SendUserMessage(ctxWithMeta, req)
	if err != nil {
		return err
	}
	if res != nil && !res.Success {
		return fmt.Errorf(res.Message)
	}

	return nil
}
