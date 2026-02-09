package balance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "jh_balance_service/api/frontend/balance/v1"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
)

// GetDepositBonusRecentWinners 获取充值送活动最近中奖榜
func (s *sBalance) GetDepositBonusRecentWinners(ctx context.Context, req *v1.GetDepositBonusRecentWinnersReq) (*v1.GetDepositBonusRecentWinnersRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetDepositBonusRecentWinnersRes{Success: false, Message: "未获取站点ID"}, nil
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

	var records []*entity.BalanceDepositReward
	if err := dao.BalanceDepositReward.Ctx(ctx).
		Where("site_id", siteId).
		OrderDesc("id").
		Limit(limit).
		Scan(&records); err != nil {
		g.Log().Errorf(ctx, "查询充值送活动中奖记录失败: site_id=%d, err=%v", siteId, err)
		return &v1.GetDepositBonusRecentWinnersRes{Success: false, Message: "查询充值送活动中奖记录失败"}, nil
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

	list := make([]*v1.DepositBonusRecentWinnerItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		maskName := maskLuckyUsername(usernames[record.UserId])
		if maskName == "" {
			maskName = fmt.Sprintf("U***%d", record.UserId%1000)
		}
		depositAmount := s.getDepositAmountForReward(ctx, int(siteId), record.UserId, record.CreatedAt)
		list = append(list, &v1.DepositBonusRecentWinnerItem{
			UserId:        int32(record.UserId),
			MaskName:      maskName,
			DepositAmount: round2(depositAmount),
			BonusAmount:   round2(record.BonusAmount),
			CreatedAt:     formatTime(record.CreatedAt),
		})
	}

	list = s.appendDepositBonusRandomWinners(ctx, int(siteId), limit, list)

	return &v1.GetDepositBonusRecentWinnersRes{
		Success: true,
		Message: "获取成功",
		List:    list,
	}, nil
}

func (s *sBalance) getDepositAmountForReward(ctx context.Context, siteId int, userId int, rewardTime *gtime.Time) float64 {
	if siteId <= 0 || userId <= 0 {
		return 0
	}
	model := dao.BalanceLogPayment.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", balanceConsts.StatusSuccess).
		Where("actual_money > 0")
	if rewardTime != nil {
		model = model.Where("created_at <= ?", rewardTime)
	}
	value, err := model.OrderDesc("id").Limit(1).Value("actual_money")
	if err != nil {
		return 0
	}
	return gconv.Float64(value)
}

func (s *sBalance) appendDepositBonusRandomWinners(
	ctx context.Context,
	siteId int,
	limit int,
	list []*v1.DepositBonusRecentWinnerItem,
) []*v1.DepositBonusRecentWinnerItem {
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

	cfg, _ := getDepositActivityParams(ctx, siteId)
	randomPairs := []struct {
		Deposit float64
		Bonus   float64
	}{
		{Deposit: 100, Bonus: 50},
		{Deposit: 500, Bonus: 300},
		{Deposit: 1000, Bonus: 800},
		{Deposit: 3000, Bonus: 3000},
		{Deposit: 5000, Bonus: 5000},
		{Deposit: 10000, Bonus: 8888},
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < need; i++ {
		pair := randomPairs[r.Intn(len(randomPairs))]
		depositAmount := pair.Deposit
		bonusAmount := pair.Bonus
		if cfg != nil {
			if cfg.RewardType == 2 {
				bonusAmount = depositAmount * cfg.RewardAmount
			} else if cfg.RewardType == 1 && cfg.RewardAmount > 0 {
				bonusAmount = cfg.RewardAmount
			}
		}
		maskName := fmt.Sprintf("U***%03d", r.Intn(1000))
		createdAt := gtime.Now().Add(-time.Duration(r.Intn(72*3600)) * time.Second)

		list = append(list, &v1.DepositBonusRecentWinnerItem{
			UserId:        0,
			MaskName:      maskName,
			DepositAmount: round2(depositAmount),
			BonusAmount:   round2(bonusAmount),
			CreatedAt:     formatTime(createdAt),
		})
	}
	return list
}
