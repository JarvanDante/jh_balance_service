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

type luckyActivityConfig struct {
	ActivityId       int
	SiteCode         string
	NeedBindMobile   bool
	JackpotShowValue float64
}

// getLuckyActivityConfig 获取当前站点有效的大转盘活动配置（全局编码 LUCKY）。
func getLuckyActivityConfig(ctx context.Context, siteId int) (*luckyActivityConfig, error) {
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
		Code:   activityCodeLucky,
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

		cfg := &luckyActivityConfig{
			ActivityId: int(activity.ActivityId),
			SiteCode:   activity.SiteCode,
		}

		if activity.FixedParams != "" {
			var fixedParams map[string]interface{}
			if err := json.Unmarshal([]byte(activity.FixedParams), &fixedParams); err == nil {
				if raw, ok := fixedParams["need_bind_mobile"]; ok {
					cfg.NeedBindMobile = gconv.Bool(raw) || gconv.Int(raw) == 1
				}
				if raw, ok := fixedParams["jackpot_amount"]; ok {
					cfg.JackpotShowValue = gconv.Float64(raw)
				}
			}
		}

		return cfg, nil
	}

	return nil, nil
}
