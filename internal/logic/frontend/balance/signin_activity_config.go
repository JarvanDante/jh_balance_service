package balance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sitev1 "jh_balance_service/api/grpc/site/backend/v1"
	"jh_balance_service/internal/grpc_client"

	"github.com/gogf/gf/v2/util/gconv"
)

// getSigninActivity 获取当前站点有效的签到活动配置（全局编码 SIGNIN）。
func getSigninActivity(ctx context.Context, siteId int) (*sitev1.SiteActivityInfo, error) {
	if siteId <= 0 {
		return nil, fmt.Errorf("siteId invalid")
	}

	ctx = grpc_client.WithSiteContext(ctx, siteId)
	client, conn, err := grpc_client.GetSiteClient(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	listRes, err := client.GetSiteActivityList(ctx, &sitev1.GetSiteActivityListReq{
		SiteId: int32(siteId),
		Code:   activityCodeSignin,
	})
	if err != nil {
		return nil, err
	}
	if listRes == nil || len(listRes.List) == 0 {
		return nil, nil
	}

	now := time.Now()
	for _, activity := range listRes.List {
		if activity == nil || activity.ActivityId == 0 || activity.Status != 1 {
			continue
		}

		startTime, err := parseActivityTime(activity.StartTime)
		if err == nil && startTime != nil && now.Before(*startTime) {
			continue
		}
		endTime, err := parseActivityTime(activity.EndTime)
		if err == nil && endTime != nil && now.After(*endTime) {
			continue
		}

		return activity, nil
	}

	return nil, nil
}

func parseSigninClaimThreshold(activity *sitev1.SiteActivityInfo) (minRechargeAmount float64, minValidBetAmount float64) {
	if activity == nil || activity.GetFixedParams() == "" {
		return 0, 0
	}

	params := map[string]interface{}{}
	if err := json.Unmarshal([]byte(activity.GetFixedParams()), &params); err != nil {
		return 0, 0
	}

	minRechargeAmount = gconv.Float64(params["min_recharge_amount"])
	if minRechargeAmount < 0 {
		minRechargeAmount = 0
	}

	minValidBetAmount = gconv.Float64(params["min_valid_bet_amount"])
	if minValidBetAmount < 0 {
		minValidBetAmount = 0
	}

	return minRechargeAmount, minValidBetAmount
}
