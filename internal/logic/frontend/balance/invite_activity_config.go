package balance

import (
	"context"
	"fmt"

	sitev1 "jh_balance_service/api/grpc/site/backend/v1"
	"jh_balance_service/internal/grpc_client"
)

type inviteActivityParams struct {
	ActivityId          int
	Table               string
	TriggerEvent        int
	RewardType          int
	MinRechargeAmount   float64
	InviterRewardAmount float64
}

func getInviteActivityParams(ctx context.Context, siteId int) (*inviteActivityParams, error) {
	if siteId <= 0 {
		return nil, fmt.Errorf("siteId invalid")
	}

	client, conn, err := grpc_client.GetSiteClient(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := client.GetInviteBonusConfig(ctx, &sitev1.GetInviteBonusConfigReq{
		SiteId: int32(siteId),
		Code:   "INVITE",
	})
	if err != nil {
		return nil, err
	}
	if res == nil || res.ActivityId == 0 || res.Status != 1 {
		return nil, nil
	}

	return &inviteActivityParams{
		ActivityId:          int(res.ActivityId),
		Table:               res.Table,
		TriggerEvent:        int(res.TriggerEvent),
		RewardType:          int(res.RewardType),
		MinRechargeAmount:   res.MinRechargeAmount,
		InviterRewardAmount: res.InviterRewardAmount,
	}, nil
}
