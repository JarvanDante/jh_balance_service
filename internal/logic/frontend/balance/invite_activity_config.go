package balance

import (
	"context"
	"fmt"
	"sync"
	"time"

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

const inviteConfigCacheTTL = 30 * time.Second

type inviteConfigCacheEntry struct {
	value     *inviteActivityParams
	expiresAt time.Time
}

var inviteConfigCache = struct {
	sync.RWMutex
	items map[string]inviteConfigCacheEntry
}{
	items: map[string]inviteConfigCacheEntry{},
}

func getInviteActivityParams(ctx context.Context, siteId int) (*inviteActivityParams, error) {
	if siteId <= 0 {
		return nil, fmt.Errorf("siteId invalid")
	}

	cacheKey := fmt.Sprintf("site:%d:INVITE", siteId)
	now := time.Now()
	inviteConfigCache.RLock()
	if entry, ok := inviteConfigCache.items[cacheKey]; ok && now.Before(entry.expiresAt) {
		inviteConfigCache.RUnlock()
		return entry.value, nil
	}
	inviteConfigCache.RUnlock()

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
		inviteConfigCache.Lock()
		inviteConfigCache.items[cacheKey] = inviteConfigCacheEntry{
			value:     nil,
			expiresAt: now.Add(inviteConfigCacheTTL),
		}
		inviteConfigCache.Unlock()
		return nil, nil
	}

	value := &inviteActivityParams{
		ActivityId:          int(res.ActivityId),
		Table:               res.Table,
		TriggerEvent:        int(res.TriggerEvent),
		RewardType:          int(res.RewardType),
		MinRechargeAmount:   res.MinRechargeAmount,
		InviterRewardAmount: res.InviterRewardAmount,
	}

	inviteConfigCache.Lock()
	inviteConfigCache.items[cacheKey] = inviteConfigCacheEntry{
		value:     value,
		expiresAt: now.Add(inviteConfigCacheTTL),
	}
	inviteConfigCache.Unlock()

	return value, nil
}
