package balance

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"google.golang.org/grpc/metadata"

	v1 "jh_balance_service/api/frontend/balance/v1"
	sitev1 "jh_balance_service/api/grpc/site/backend/v1"
	userv1 "jh_balance_service/api/grpc/user/backend/v1"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/grpc_client"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/registry"
	"jh_balance_service/internal/util"
)

type luckyPrizeConfig struct {
	SegmentNo         int32   `json:"segment_no"`
	PrizeName         string  `json:"prize_name"`
	RewardAmount      float64 `json:"reward_amount"`
	Weight            int32   `json:"weight"`
	UserDailyWinLimit int32   `json:"user_daily_win_limit"`
	MinVipLevel       int32   `json:"min_vip_level"`
	Icon              string  `json:"icon"`
}

type luckySpinUserInfo struct {
	Username string
	GradeId  int32
	Mobile   string
}

const luckyWheelSegmentCount = 8

// GetLuckySpinInfo 获取大转盘活动信息
func (s *sBalance) GetLuckySpinInfo(ctx context.Context, req *v1.GetLuckySpinInfoReq) (*v1.GetLuckySpinInfoRes, error) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		return &v1.GetLuckySpinInfoRes{Success: false, Message: "未登录"}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetLuckySpinInfoRes{Success: false, Message: "未获取站点ID"}, nil
	}

	siteCode := s.ensureSiteRuntimeContext(ctx, int(siteId))
	if siteCode != "" && ctx.Value("dynamic_db_group") == nil {
		ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
	}

	cfg, err := getLuckyActivityConfig(ctx, int(siteId))
	if err != nil {
		g.Log().Errorf(ctx, "获取大转盘活动配置失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetLuckySpinInfoRes{Success: false, Message: "获取大转盘活动配置失败"}, nil
	}

	res := &v1.GetLuckySpinInfoRes{
		Success:         true,
		Message:         "获取成功",
		Active:          false,
		FreeSpinsPerDay: 0,
		FreeSpinsLeft:   0,
		TodaySpinCount:  0,
		JackpotAmount:   0,
		TotalWon:        0,
		PrizeList:       make([]*v1.LuckySpinPrizeItem, 0),
	}
	if cfg == nil {
		res.Message = "大转盘活动未开启"
		return res, nil
	}

	activitySiteCode := cfg.SiteCode
	if activitySiteCode == "" {
		activitySiteCode = siteCode
	}
	prizeConfigs, err := s.getLuckyPrizeConfigs(ctx, int(siteId), activitySiteCode, cfg.ActivityId)
	if err != nil {
		msg := "查询大转盘奖项配置失败"
		if strings.Contains(err.Error(), "未配置") {
			msg = err.Error()
		}
		g.Log().Errorf(ctx, "查询大转盘奖项配置失败: site_id=%d, user_id=%d, activity_id=%d, err=%v",
			siteId, userId, cfg.ActivityId, err)
		return &v1.GetLuckySpinInfoRes{Success: false, Message: msg}, nil
	}

	todaySpinCount, totalWon, err := s.getUserLuckySpinStats(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Errorf(ctx, "查询大转盘统计失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetLuckySpinInfoRes{Success: false, Message: "查询大转盘统计失败"}, nil
	}

	totalGranted, _, freeSpinsLeft, err := s.getUserLuckySpinChanceStats(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Errorf(ctx, "查询大转盘次数失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetLuckySpinInfoRes{Success: false, Message: "查询大转盘次数失败"}, nil
	}

	jackpotAmount := cfg.JackpotShowValue
	if jackpotAmount <= 0 {
		jackpotAmount = calcLuckyJackpotAmount(prizeConfigs)
	}

	prizeList := make([]*v1.LuckySpinPrizeItem, 0, len(prizeConfigs))
	for _, prize := range prizeConfigs {
		prizeList = append(prizeList, &v1.LuckySpinPrizeItem{
			SegmentNo:         prize.SegmentNo,
			PrizeName:         prize.PrizeName,
			RewardAmount:      round2(prize.RewardAmount),
			Weight:            prize.Weight,
			UserDailyWinLimit: prize.UserDailyWinLimit,
			MinVipLevel:       prize.MinVipLevel,
			Icon:              prize.Icon,
		})
	}

	res.Active = true
	res.FreeSpinsPerDay = totalGranted
	res.FreeSpinsLeft = freeSpinsLeft
	res.TodaySpinCount = todaySpinCount
	res.JackpotAmount = round2(jackpotAmount)
	res.TotalWon = round2(totalWon)
	res.PrizeList = prizeList
	return res, nil
}

// LuckySpin 执行大转盘抽奖
func (s *sBalance) LuckySpin(ctx context.Context, req *v1.LuckySpinReq) (*v1.LuckySpinRes, error) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		return &v1.LuckySpinRes{Success: false, Message: "未登录"}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.LuckySpinRes{Success: false, Message: "未获取站点ID"}, nil
	}

	siteCode := s.ensureSiteRuntimeContext(ctx, int(siteId))
	if siteCode != "" && ctx.Value("dynamic_db_group") == nil {
		ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
	}

	cfg, err := getLuckyActivityConfig(ctx, int(siteId))
	if err != nil {
		g.Log().Errorf(ctx, "获取大转盘活动配置失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.LuckySpinRes{Success: false, Message: "获取大转盘活动配置失败"}, nil
	}
	if cfg == nil {
		return &v1.LuckySpinRes{Success: false, Message: "大转盘活动未开启"}, nil
	}

	activitySiteCode := cfg.SiteCode
	if activitySiteCode == "" {
		activitySiteCode = siteCode
	}
	prizeConfigs, err := s.getLuckyPrizeConfigs(ctx, int(siteId), activitySiteCode, cfg.ActivityId)
	if err != nil {
		msg := "查询大转盘奖项配置失败"
		if strings.Contains(err.Error(), "未配置") {
			msg = err.Error()
		}
		g.Log().Errorf(ctx, "查询大转盘奖项配置失败: site_id=%d, user_id=%d, activity_id=%d, err=%v",
			siteId, userId, cfg.ActivityId, err)
		return &v1.LuckySpinRes{Success: false, Message: msg}, nil
	}
	if len(prizeConfigs) == 0 {
		return &v1.LuckySpinRes{Success: false, Message: "大转盘奖项未配置"}, nil
	}

	userInfo, err := s.getLuckySpinUserInfo(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Errorf(ctx, "查询用户信息失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.LuckySpinRes{Success: false, Message: "查询用户信息失败"}, nil
	}
	if cfg.NeedBindMobile && strings.TrimSpace(userInfo.Mobile) == "" {
		return &v1.LuckySpinRes{Success: false, Message: "请先绑定手机号再参与大转盘"}, nil
	}

	vipQualifiedPrizes := make([]*luckyPrizeConfig, 0, len(prizeConfigs))
	for _, prize := range prizeConfigs {
		if prize.MinVipLevel > 0 && userInfo.GradeId < prize.MinVipLevel {
			continue
		}
		vipQualifiedPrizes = append(vipQualifiedPrizes, prize)
	}
	if len(vipQualifiedPrizes) == 0 {
		return &v1.LuckySpinRes{Success: false, Message: "当前VIP等级无可用奖项"}, nil
	}

	var (
		selectedPrize *luckyPrizeConfig
		tradeNo       string
	)
	todayDate := gtime.Now().Format("Y-m-d")

	err = dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var balanceUser *entity.BalanceUser
		if err := dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("site_id", siteId).
			Where("user_id", userId).
			LockUpdate().
			Scan(&balanceUser); err != nil {
			return err
		}
		if balanceUser == nil || balanceUser.Id == 0 {
			return fmt.Errorf("用户余额记录不存在")
		}

		_, _, leftCount, err := s.getUserLuckySpinChanceStatsTx(ctx, tx, int(siteId), int(userId))
		if err != nil {
			return err
		}
		if leftCount <= 0 {
			return fmt.Errorf("抽奖次数不足")
		}

		eligiblePrizes := make([]*luckyPrizeConfig, 0, len(vipQualifiedPrizes))
		for _, prize := range vipQualifiedPrizes {
			if prize.Weight <= 0 {
				continue
			}
			if prize.UserDailyWinLimit > 0 {
				winCount, err := dao.BalanceWheelReward.Ctx(ctx).TX(tx).
					Where("site_id", siteId).
					Where("user_id", userId).
					Where("date", todayDate).
					Where("segment_no", prize.SegmentNo).
					Where("reward_amount > 0").
					Count()
				if err != nil {
					return err
				}
				if winCount >= int(prize.UserDailyWinLimit) {
					continue
				}
			}
			eligiblePrizes = append(eligiblePrizes, prize)
		}

		selectedPrize = drawLuckyPrizeByWeight(eligiblePrizes)
		if selectedPrize == nil {
			return fmt.Errorf("当前无可中奖奖项，请稍后再试")
		}

		amount := round2(selectedPrize.RewardAmount)
		tradeNo = util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_BONUS)
		now := gtime.Now()

		if amount > 0 {
			balanceOld := balanceUser.Balance
			balanceNew := balanceOld + amount
			remark := fmt.Sprintf("大转盘中奖[%s]", selectedPrize.PrizeName)

			if _, err := dao.BalanceUser.Ctx(ctx).TX(tx).
				Where("id", balanceUser.Id).
				Update(g.Map{
					"balance":    balanceNew,
					"updated_at": now,
				}); err != nil {
				return err
			}

			if _, err := dao.BalanceLog.Ctx(ctx).TX(tx).Insert(g.Map{
				"site_id":      siteId,
				"user_id":      userId,
				"trade_type":   balanceConsts.LOG_TYPE_BONUS,
				"trade_no":     tradeNo,
				"money":        amount,
				"actual_money": amount,
				"fee":          0,
				"status":       balanceConsts.StatusSuccess,
				"remark":       remark,
				"created_at":   now,
				"updated_at":   now,
			}); err != nil {
				return err
			}

			if _, err := dao.BalanceChange.Ctx(ctx).TX(tx).Insert(g.Map{
				"site_id":        siteId,
				"trade_type":     balanceConsts.TradeTypeBonus,
				"user_id":        userId,
				"username":       balanceUser.Username,
				"trade_no":       tradeNo,
				"balance_old":    balanceOld,
				"money":          amount,
				"balance_new":    balanceNew,
				"balance_frozen": balanceUser.BalanceFrozen,
				"status":         balanceConsts.StatusSuccess,
				"remark":         remark,
				"created_at":     now,
				"updated_at":     now,
			}); err != nil {
				return err
			}

			bonusResult, err := dao.BalanceLogBonus.Ctx(ctx).TX(tx).Insert(g.Map{
				"site_id":              siteId,
				"trade_type":           balanceConsts.BonusTradeTypeLuckySpin,
				"user_id":              userId,
				"username":             balanceUser.Username,
				"trade_no":             tradeNo,
				"activity_recharge_id": 0,
				"money":                amount,
				"fee":                  0,
				"status":               balanceConsts.StatusSuccess,
				"admin_id":             0,
				"remark":               remark,
				"created_at":           now,
				"updated_at":           now,
			})
			if err != nil {
				return err
			}
			bonusId, _ := bonusResult.LastInsertId()

			rule, err := s.getLuckySpinFlowRule(ctx, int(siteId))
			if err != nil {
				return err
			}
			if rule == nil {
				return fmt.Errorf("未配置可用的流水规则")
			}

			flowMultiple := rule.FlowMultiple
			requiredFlow := amount * flowMultiple
			var expireTime *gtime.Time
			if rule.ExpireDays > 0 {
				expireTime = now.AddDate(0, 0, rule.ExpireDays)
			}

			if _, err := dao.BalanceFlowRequirement.Ctx(ctx).TX(tx).Insert(g.Map{
				"site_id":        siteId,
				"user_id":        userId,
				"username":       balanceUser.Username,
				"source_type":    balanceConsts.SourceTypeLuckySpin,
				"source_id":      bonusId,
				"trade_no":       tradeNo,
				"amount":         0,
				"bonus_amount":   amount,
				"flow_multiple":  flowMultiple,
				"required_flow":  requiredFlow,
				"completed_flow": 0,
				"remaining_flow": requiredFlow,
				"status":         1,
				"progress":       0,
				"start_time":     now,
				"expire_time":    expireTime,
				"complete_time":  nil,
				"remark":         remark,
				"admin_id":       0,
				"created_at":     now,
				"updated_at":     now,
			}); err != nil {
				return err
			}
		}

		if _, err := dao.BalanceWheelReward.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":       siteId,
			"user_id":       userId,
			"date":          todayDate,
			"segment_no":    selectedPrize.SegmentNo,
			"prize_name":    selectedPrize.PrizeName,
			"reward_amount": amount,
			"trade_no":      tradeNo,
			"created_at":    now,
			"updated_at":    now,
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		g.Log().Errorf(ctx, "执行大转盘抽奖失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.LuckySpinRes{Success: false, Message: err.Error()}, nil
	}

	todaySpinCountRes, totalWon, err := s.getUserLuckySpinStats(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Warningf(ctx, "抽奖后查询大转盘统计失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
	}

	_, _, freeSpinsLeft, err := s.getUserLuckySpinChanceStats(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Warningf(ctx, "抽奖后查询大转盘次数失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		freeSpinsLeft = 0
	}

	return &v1.LuckySpinRes{
		Success:        true,
		Message:        "抽奖成功",
		TradeNo:        tradeNo,
		SegmentNo:      selectedPrize.SegmentNo,
		PrizeName:      selectedPrize.PrizeName,
		RewardAmount:   round2(selectedPrize.RewardAmount),
		FreeSpinsLeft:  freeSpinsLeft,
		TodaySpinCount: todaySpinCountRes,
		TotalWon:       round2(totalWon),
	}, nil
}

// GetLuckySpinRecords 获取我的大转盘中奖记录
func (s *sBalance) GetLuckySpinRecords(ctx context.Context, req *v1.GetLuckySpinRecordsReq) (*v1.GetLuckySpinRecordsRes, error) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		return &v1.GetLuckySpinRecordsRes{Success: false, Message: "未登录"}, nil
	}
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetLuckySpinRecordsRes{Success: false, Message: "未获取站点ID"}, nil
	}

	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	siteCode := s.ensureSiteRuntimeContext(ctx, int(siteId))
	if siteCode != "" && ctx.Value("dynamic_db_group") == nil {
		ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
	}

	model := dao.BalanceWheelReward.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId)

	count, err := model.Count()
	if err != nil {
		g.Log().Errorf(ctx, "查询大转盘记录总数失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetLuckySpinRecordsRes{Success: false, Message: "查询大转盘记录失败"}, nil
	}

	var records []*entity.BalanceWheelReward
	if err := model.
		OrderDesc("id").
		Limit(int(size)).
		Offset(int((page - 1) * size)).
		Scan(&records); err != nil {
		g.Log().Errorf(ctx, "查询大转盘记录失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetLuckySpinRecordsRes{Success: false, Message: "查询大转盘记录失败"}, nil
	}

	list := make([]*v1.LuckySpinRecordItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		list = append(list, &v1.LuckySpinRecordItem{
			TradeNo:      record.TradeNo,
			SegmentNo:    int32(record.SegmentNo),
			PrizeName:    record.PrizeName,
			RewardAmount: round2(record.RewardAmount),
			CreatedAt:    formatTime(record.CreatedAt),
		})
	}

	return &v1.GetLuckySpinRecordsRes{
		Success: true,
		Message: "获取成功",
		List:    list,
		Count:   int32(count),
		Page:    page,
		Size:    size,
	}, nil
}

// GetLuckySpinRecentWinners 获取大转盘最近中奖榜
func (s *sBalance) GetLuckySpinRecentWinners(ctx context.Context, req *v1.GetLuckySpinRecentWinnersReq) (*v1.GetLuckySpinRecentWinnersRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetLuckySpinRecentWinnersRes{Success: false, Message: "未获取站点ID"}, nil
	}

	limit := req.Limit
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

	var winners []*entity.BalanceWheelReward
	if err := dao.BalanceWheelReward.Ctx(ctx).
		Where("site_id", siteId).
		Where("reward_amount > 0").
		OrderDesc("id").
		Limit(int(limit)).
		Scan(&winners); err != nil {
		g.Log().Errorf(ctx, "查询大转盘最近中奖榜失败: site_id=%d, err=%v", siteId, err)
		return &v1.GetLuckySpinRecentWinnersRes{Success: false, Message: "查询大转盘最近中奖榜失败"}, nil
	}

	userIdSet := make(map[int]struct{}, len(winners))
	for _, winner := range winners {
		if winner != nil && winner.UserId > 0 {
			userIdSet[winner.UserId] = struct{}{}
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

	list := make([]*v1.LuckySpinRecentWinnerItem, 0, len(winners))
	for _, winner := range winners {
		if winner == nil {
			continue
		}
		maskName := maskLuckyUsername(usernames[winner.UserId])
		if maskName == "" {
			maskName = fmt.Sprintf("U***%d", winner.UserId%1000)
		}
		list = append(list, &v1.LuckySpinRecentWinnerItem{
			UserId:       int32(winner.UserId),
			MaskName:     maskName,
			PrizeName:    winner.PrizeName,
			RewardAmount: round2(winner.RewardAmount),
			CreatedAt:    formatTime(winner.CreatedAt),
		})
	}

	return &v1.GetLuckySpinRecentWinnersRes{
		Success: true,
		Message: "获取成功",
		List:    list,
	}, nil
}

func (s *sBalance) getLuckyPrizeConfigs(ctx context.Context, siteId int, siteCode string, activityId int) ([]*luckyPrizeConfig, error) {
	if siteId <= 0 || activityId <= 0 {
		return nil, fmt.Errorf("大转盘参数无效")
	}

	client, conn, err := grpc_client.GetSiteClient(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	paramRes, err := client.GetSiteActivityParams(ctx, &sitev1.GetSiteActivityParamsReq{
		SiteId:     int32(siteId),
		ActivityId: int32(activityId),
		SiteCode:   strings.TrimSpace(siteCode),
	})
	if err != nil {
		return nil, err
	}

	values := map[string]string{}
	if paramRes != nil {
		values = paramRes.Values
	}

	rows := make([]*luckyPrizeConfig, 0, luckyWheelSegmentCount)
	hasConfigured := false
	for segment := 1; segment <= luckyWheelSegmentCount; segment++ {
		keyPrefix := fmt.Sprintf("segment_%d_", segment)
		prizeName := strings.TrimSpace(values[keyPrefix+"prize_name"])
		rewardAmount := gconv.Float64(values[keyPrefix+"reward_amount"])
		if rewardAmount < 0 {
			rewardAmount = 0
		}
		weight := gconv.Int(values[keyPrefix+"weight"])
		if weight < 0 {
			weight = 0
		}
		userDailyWinLimit := gconv.Int(values[keyPrefix+"user_daily_win_limit"])
		if userDailyWinLimit < 0 {
			userDailyWinLimit = 0
		}
		minVipLevel := gconv.Int(values[keyPrefix+"min_vip_level"])
		if minVipLevel < 0 {
			minVipLevel = 0
		}
		icon := strings.TrimSpace(values[keyPrefix+"icon"])

		if prizeName != "" || rewardAmount > 0 || weight > 0 || userDailyWinLimit > 0 || minVipLevel > 0 || icon != "" {
			hasConfigured = true
		}

		rows = append(rows, &luckyPrizeConfig{
			SegmentNo:         int32(segment),
			PrizeName:         prizeName,
			RewardAmount:      round2(rewardAmount),
			Weight:            int32(weight),
			UserDailyWinLimit: int32(userDailyWinLimit),
			MinVipLevel:       int32(minVipLevel),
			Icon:              icon,
		})
	}

	if !hasConfigured {
		return nil, fmt.Errorf("大转盘奖项未配置")
	}
	return rows, nil
}

func (s *sBalance) getLuckySpinUserInfo(ctx context.Context, siteId int, userId int) (*luckySpinUserInfo, error) {
	userClient, userConn, err := grpc_client.GetUserClient(ctx)
	if err != nil {
		return nil, err
	}
	defer userConn.Close()

	ctxWithMeta := metadata.AppendToOutgoingContext(ctx, "site_id", gconv.String(siteId))
	res, err := userClient.GetUserBasicInfo(ctxWithMeta, &userv1.GetUserBasicInfoReq{
		Id: int32(userId),
	})
	if err != nil {
		return nil, err
	}
	if res == nil || res.User == nil {
		return nil, fmt.Errorf("用户信息不存在")
	}

	return &luckySpinUserInfo{
		Username: strings.TrimSpace(res.User.Username),
		GradeId:  res.User.GradeId,
		Mobile:   strings.TrimSpace(res.User.Mobile),
	}, nil
}

func (s *sBalance) getUserLuckySpinStats(ctx context.Context, siteId int, userId int) (int32, float64, error) {
	todayDate := gtime.Now().Format("Y-m-d")

	model := dao.BalanceWheelReward.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId)

	todayCount, err := model.Where("date", todayDate).Count()
	if err != nil {
		return 0, 0, err
	}
	totalWon, err := model.Sum("reward_amount")
	if err != nil {
		return 0, 0, err
	}
	return int32(todayCount), round2(totalWon), nil
}

// getUserLuckySpinChanceStats 获取用户累计发放次数、已抽次数、剩余次数。
// 规则：注册1次 + 每次成功充值1次 + 每次VIP升级红利1次。
func (s *sBalance) getUserLuckySpinChanceStats(ctx context.Context, siteId int, userId int) (int32, int32, int32, error) {
	depositCount, err := dao.BalanceLogPayment.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", balanceConsts.StatusSuccess).
		Where("actual_money > 0").
		Count()
	if err != nil {
		return 0, 0, 0, err
	}

	upgradeCount, err := dao.BalanceLogBonus.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", balanceConsts.StatusSuccess).
		Where("trade_type", balanceConsts.BonusTradeTypeUpgrade).
		Count()
	if err != nil {
		return 0, 0, 0, err
	}

	usedCount, err := dao.BalanceWheelReward.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Count()
	if err != nil {
		return 0, 0, 0, err
	}

	grantedCount := 1 + depositCount + upgradeCount
	leftCount := grantedCount - usedCount
	if leftCount < 0 {
		leftCount = 0
	}
	return int32(grantedCount), int32(usedCount), int32(leftCount), nil
}

func (s *sBalance) getUserLuckySpinChanceStatsTx(ctx context.Context, tx gdb.TX, siteId int, userId int) (int, int, int, error) {
	depositCount, err := dao.BalanceLogPayment.Ctx(ctx).TX(tx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", balanceConsts.StatusSuccess).
		Where("actual_money > 0").
		Count()
	if err != nil {
		return 0, 0, 0, err
	}

	upgradeCount, err := dao.BalanceLogBonus.Ctx(ctx).TX(tx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", balanceConsts.StatusSuccess).
		Where("trade_type", balanceConsts.BonusTradeTypeUpgrade).
		Count()
	if err != nil {
		return 0, 0, 0, err
	}

	usedCount, err := dao.BalanceWheelReward.Ctx(ctx).TX(tx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Count()
	if err != nil {
		return 0, 0, 0, err
	}

	grantedCount := 1 + depositCount + upgradeCount
	leftCount := grantedCount - usedCount
	if leftCount < 0 {
		leftCount = 0
	}
	return grantedCount, usedCount, leftCount, nil
}

func (s *sBalance) getLuckySpinFlowRule(ctx context.Context, siteId int) (*entity.BalanceFlowRule, error) {
	rule, err := dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeLuckySpin)
	if err != nil {
		return nil, fmt.Errorf("查询流水规则失败")
	}
	if rule == nil {
		rule, err = dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeBonus)
		if err != nil {
			return nil, fmt.Errorf("查询流水规则失败")
		}
	}
	return rule, nil
}

func (s *sBalance) ensureSiteRuntimeContext(ctx context.Context, siteId int) string {
	siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
	if siteCode == "" {
		siteCode = registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
	}
	return siteCode
}

func drawLuckyPrizeByWeight(prizes []*luckyPrizeConfig) *luckyPrizeConfig {
	if len(prizes) == 0 {
		return nil
	}

	totalWeight := 0
	for _, prize := range prizes {
		if prize != nil && prize.Weight > 0 {
			totalWeight += int(prize.Weight)
		}
	}
	if totalWeight <= 0 {
		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomValue := r.Intn(totalWeight) + 1

	current := 0
	for _, prize := range prizes {
		if prize == nil || prize.Weight <= 0 {
			continue
		}
		current += int(prize.Weight)
		if randomValue <= current {
			return prize
		}
	}
	return nil
}

func calcLuckyJackpotAmount(prizes []*luckyPrizeConfig) float64 {
	maxAmount := 0.0
	for _, prize := range prizes {
		if prize == nil {
			continue
		}
		if prize.RewardAmount > maxAmount {
			maxAmount = prize.RewardAmount
		}
	}
	return round2(maxAmount)
}

func maskLuckyUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}
	runes := []rune(username)
	if len(runes) <= 2 {
		return string(runes[0]) + "*"
	}
	if len(runes) <= 4 {
		return string(runes[0]) + "**" + string(runes[len(runes)-1])
	}
	return string(runes[0:2]) + "***" + string(runes[len(runes)-2:])
}

func formatTime(t *gtime.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("Y-m-d H:i:s")
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
