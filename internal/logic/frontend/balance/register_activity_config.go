package balance

import (
	"context"
	"fmt"
	"sync"
	"time"

	sitev1 "jh_balance_service/api/grpc/site/backend/v1"
	"jh_balance_service/internal/grpc_client"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

type registerActivityParams struct {
	ActivityId   int
	RewardAmount float64
	LimitIp      int
	LimitDevice  int
}

const registerConfigCacheTTL = 30 * time.Second
const registerConfigErrorTTL = 5 * time.Second

type registerConfigCacheEntry struct {
	value     *registerActivityParams
	err       error
	expiresAt time.Time
}

var registerConfigCache = struct {
	sync.RWMutex
	items map[string]registerConfigCacheEntry
}{
	items: map[string]registerConfigCacheEntry{},
}

func parseActivityTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := gtime.StrToTime(value)
	if err != nil {
		return nil, err
	}
	parsed := t.Time
	return &parsed, nil
}

func getRegisterActivityParams(ctx context.Context, siteId int) (*registerActivityParams, error) {
	if siteId <= 0 {
		return nil, fmt.Errorf("siteId invalid")
	}

	cacheKey := fmt.Sprintf("site:%d:%s", siteId, activityCodeRegister)
	now := time.Now()
	registerConfigCache.RLock()
	if entry, ok := registerConfigCache.items[cacheKey]; ok && now.Before(entry.expiresAt) {
		registerConfigCache.RUnlock()
		return entry.value, entry.err
	}
	registerConfigCache.RUnlock()

	client, conn, err := grpc_client.GetSiteClient(ctx)
	if err != nil {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       err,
			expiresAt: now.Add(registerConfigErrorTTL),
		}
		registerConfigCache.Unlock()
		return nil, err
	}
	defer conn.Close()

	listRes, err := client.GetSiteActivityList(ctx, &sitev1.GetSiteActivityListReq{
		SiteId: int32(siteId),
		Code:   activityCodeRegister,
	})
	if err != nil {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       err,
			expiresAt: now.Add(registerConfigErrorTTL),
		}
		registerConfigCache.Unlock()
		return nil, err
	}

	if listRes == nil || len(listRes.List) == 0 {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(registerConfigCacheTTL),
		}
		registerConfigCache.Unlock()
		return nil, nil
	}

	activity := listRes.List[0]
	if activity == nil || activity.ActivityId == 0 || activity.Status != 1 {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(registerConfigCacheTTL),
		}
		registerConfigCache.Unlock()
		return nil, nil
	}

	startTime, err := parseActivityTime(activity.StartTime)
	if err == nil && startTime != nil && now.Before(*startTime) {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(registerConfigCacheTTL),
		}
		registerConfigCache.Unlock()
		return nil, nil
	}
	endTime, err := parseActivityTime(activity.EndTime)
	if err == nil && endTime != nil && now.After(*endTime) {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(registerConfigCacheTTL),
		}
		registerConfigCache.Unlock()
		return nil, nil
	}

	paramRes, err := client.GetSiteActivityParams(ctx, &sitev1.GetSiteActivityParamsReq{
		SiteId:     int32(siteId),
		ActivityId: activity.ActivityId,
	})
	if err != nil {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       err,
			expiresAt: now.Add(registerConfigErrorTTL),
		}
		registerConfigCache.Unlock()
		return nil, err
	}

	values := map[string]string{}
	if paramRes != nil {
		values = paramRes.Values
	}

	rewardAmount := gconv.Float64(values["reward_amount"])
	limitIp := gconv.Int(values["limit_ip"])
	limitDevice := gconv.Int(values["limit_device"])
	if rewardAmount <= 0 {
		registerConfigCache.Lock()
		registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
			value:     nil,
			err:       nil,
			expiresAt: now.Add(registerConfigCacheTTL),
		}
		registerConfigCache.Unlock()
		return nil, nil
	}

	value := &registerActivityParams{
		ActivityId:   int(activity.ActivityId),
		RewardAmount: rewardAmount,
		LimitIp:      limitIp,
		LimitDevice:  limitDevice,
	}

	registerConfigCache.Lock()
	registerConfigCache.items[cacheKey] = registerConfigCacheEntry{
		value:     value,
		err:       nil,
		expiresAt: now.Add(registerConfigCacheTTL),
	}
	registerConfigCache.Unlock()

	return value, nil
}
