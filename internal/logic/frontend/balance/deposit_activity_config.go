package balance

import (
	"context"
	"fmt"
	"sync"
	"time"

	sitev1 "jh_balance_service/api/grpc/site/backend/v1"
	"jh_balance_service/internal/grpc_client"

	"github.com/gogf/gf/v2/util/gconv"
)

type depositActivityParams struct {
	ActivityId        int
	TriggerEvent      int
	RewardType        int
	MinRechargeAmount float64
	RewardAmount      float64
}

const depositConfigCacheTTL = 30 * time.Second
const depositConfigErrorTTL = 5 * time.Second

type depositConfigCacheEntry struct {
	value     *depositActivityParams
	err       error
	expiresAt time.Time
}

var depositConfigCache = struct {
	sync.RWMutex
	items map[string]depositConfigCacheEntry
}{
	items: map[string]depositConfigCacheEntry{},
}

func getDepositActivityParams(ctx context.Context, siteId int) (*depositActivityParams, error) {
	if siteId <= 0 {
		return nil, fmt.Errorf("siteId invalid")
	}

	cacheKey := fmt.Sprintf("site:%d:DEPOSIT", siteId)
	now := time.Now()
	depositConfigCache.RLock()
	if entry, ok := depositConfigCache.items[cacheKey]; ok && now.Before(entry.expiresAt) {
		depositConfigCache.RUnlock()
		return entry.value, entry.err
	}
	depositConfigCache.RUnlock()

	client, conn, err := grpc_client.GetSiteClient(ctx)
	if err != nil {
		depositConfigCache.Lock()
		depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
			value:     nil,
			err:       err,
			expiresAt: now.Add(depositConfigErrorTTL),
		}
		depositConfigCache.Unlock()
		return nil, err
	}
	defer conn.Close()

	listRes, err := client.GetSiteActivityList(ctx, &sitev1.GetSiteActivityListReq{
		SiteId: int32(siteId),
		Code:   "DEPOSIT",
	})
	if err != nil {
		depositConfigCache.Lock()
		depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
			value:     nil,
			err:       err,
			expiresAt: now.Add(depositConfigErrorTTL),
		}
		depositConfigCache.Unlock()
		return nil, err
	}

	if listRes == nil || len(listRes.List) == 0 {
		depositConfigCache.Lock()
		depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(depositConfigCacheTTL),
		}
		depositConfigCache.Unlock()
		return nil, nil
	}

	activity := listRes.List[0]
	if activity == nil || activity.ActivityId == 0 || activity.Status != 1 {
		depositConfigCache.Lock()
		depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(depositConfigCacheTTL),
		}
		depositConfigCache.Unlock()
		return nil, nil
	}

	startTime, err := parseActivityTime(activity.StartTime)
	if err == nil && startTime != nil && now.Before(*startTime) {
		depositConfigCache.Lock()
		depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(depositConfigCacheTTL),
		}
		depositConfigCache.Unlock()
		return nil, nil
	}
	endTime, err := parseActivityTime(activity.EndTime)
	if err == nil && endTime != nil && now.After(*endTime) {
		depositConfigCache.Lock()
		depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(depositConfigCacheTTL),
		}
		depositConfigCache.Unlock()
		return nil, nil
	}

	paramRes, err := client.GetSiteActivityParams(ctx, &sitev1.GetSiteActivityParamsReq{
		SiteId:     int32(siteId),
		ActivityId: activity.ActivityId,
	})
	if err != nil {
		depositConfigCache.Lock()
		depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
			value:     nil,
			err:       err,
			expiresAt: now.Add(depositConfigErrorTTL),
		}
		depositConfigCache.Unlock()
		return nil, err
	}

	values := map[string]string{}
	if paramRes != nil {
		values = paramRes.Values
	}

	value := &depositActivityParams{
		ActivityId:        int(activity.ActivityId),
		TriggerEvent:      gconv.Int(values["trigger_event"]),
		RewardType:        gconv.Int(values["reward_type"]),
		MinRechargeAmount: gconv.Float64(values["min_recharge_amount"]),
		RewardAmount:      gconv.Float64(values["reward_amount"]),
	}

	depositConfigCache.Lock()
	depositConfigCache.items[cacheKey] = depositConfigCacheEntry{
		value:     value,
		err:       nil,
		expiresAt: now.Add(depositConfigCacheTTL),
	}
	depositConfigCache.Unlock()

	return value, nil
}
