package balance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "jh_balance_service/api/frontend/balance/v1"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
)

// GetRegisterBonusRecentWinners 获取注册送活动最近中奖榜
func (s *sBalance) GetRegisterBonusRecentWinners(ctx context.Context, req *v1.GetRegisterBonusRecentWinnersReq) (*v1.GetRegisterBonusRecentWinnersRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetRegisterBonusRecentWinnersRes{Success: false, Message: "未获取站点ID"}, nil
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	siteCode := s.ensureSiteRuntimeContext(ctx, int(siteId))
	if siteCode != "" && ctx.Value("dynamic_db_group") == nil {
		ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
	}

	var records []*entity.BalanceRegisterReward
	if err := dao.BalanceRegisterReward.Ctx(ctx).
		Where("site_id", siteId).
		OrderDesc("id").
		Limit(limit).
		Scan(&records); err != nil {
		g.Log().Errorf(ctx, "查询注册送活动中奖记录失败: site_id=%d, err=%v", siteId, err)
		return &v1.GetRegisterBonusRecentWinnersRes{Success: false, Message: "查询注册送活动中奖记录失败"}, nil
	}

	userIdSet := make(map[int]struct{}, len(records))
	for _, record := range records {
		if record != nil && record.UserId > 0 {
			userIdSet[record.UserId] = struct{}{}
		}
	}

	usernames := map[int]string{}
	if len(userIdSet) > 0 {
		ids := make([]int, 0, len(userIdSet))
		for id := range userIdSet {
			ids = append(ids, id)
		}
		var users []struct {
			UserId   int    `json:"user_id"`
			Username string `json:"username"`
		}
		err := dao.BalanceUser.Ctx(ctx).
			Where("site_id", siteId).
			WhereIn("user_id", ids).
			Fields("user_id, username").
			Scan(&users)
		if err == nil {
			for _, user := range users {
				usernames[user.UserId] = user.Username
			}
		}
	}

	list := make([]*v1.RegisterBonusRecentWinnerItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		maskName := maskLuckyUsername(usernames[record.UserId])
		if maskName == "" {
			maskName = fmt.Sprintf("U***%d", record.UserId%1000)
		}
		list = append(list, &v1.RegisterBonusRecentWinnerItem{
			UserId:      int32(record.UserId),
			MaskName:    maskName,
			BonusAmount: round2(record.BonusAmount),
			CreatedAt:   formatTime(record.CreatedAt),
		})
	}

	list = s.appendRegisterBonusRandomWinners(ctx, int(siteId), limit, list)

	return &v1.GetRegisterBonusRecentWinnersRes{
		Success: true,
		Message: "获取成功",
		List:    list,
	}, nil
}

func (s *sBalance) appendRegisterBonusRandomWinners(
	ctx context.Context,
	siteId int,
	limit int,
	list []*v1.RegisterBonusRecentWinnerItem,
) []*v1.RegisterBonusRecentWinnerItem {
	if limit <= 0 || len(list) >= limit {
		return list
	}

	const maxExtra = 5
	need := limit - len(list)
	if need > maxExtra {
		need = maxExtra
	}
	if need <= 0 {
		return list
	}

	cfg, _ := getRegisterActivityParams(ctx, siteId)
	bonusCandidates := []float64{1, 2, 5, 8, 10, 20, 50}
	if cfg != nil && cfg.RewardAmount > 0 {
		bonusCandidates = []float64{cfg.RewardAmount}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < need; i++ {
		bonusAmount := bonusCandidates[r.Intn(len(bonusCandidates))]
		maskName := fmt.Sprintf("U***%03d", r.Intn(1000))
		createdAt := gtime.Now().Add(-time.Duration(r.Intn(72*3600)) * time.Second)

		list = append(list, &v1.RegisterBonusRecentWinnerItem{
			UserId:      0,
			MaskName:    maskName,
			BonusAmount: round2(bonusAmount),
			CreatedAt:   formatTime(createdAt),
		})
	}
	return list
}
