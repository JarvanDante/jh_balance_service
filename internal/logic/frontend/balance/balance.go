package balance

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	v1 "jh_balance_service/api/frontend/balance/v1"
	gamev1 "jh_balance_service/api/grpc/game/frontend/v1"
	userv1 "jh_balance_service/api/grpc/user/backend/v1"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/domain/common"
	"jh_balance_service/internal/domain/deposit"
	flow "jh_balance_service/internal/domain/flow"
	"jh_balance_service/internal/domain/withdraw"
	"jh_balance_service/internal/grpc_client"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/registry"
	"jh_balance_service/internal/service/frontend"
	"jh_balance_service/internal/util"
)

type sBalance struct{}

func New() *sBalance {
	return &sBalance{}
}

func init() {
	frontend.RegisterBalance(New())
}

// InitBalanceUser 初始化用户余额
func (s *sBalance) InitBalanceUser(ctx context.Context, req *v1.InitBalanceUserReq) (res *v1.InitBalanceUserRes, err error) {
	g.Log().Infof(ctx, "========== Balance服务收到InitBalanceUser请求 ==========")
	g.Log().Infof(ctx, "请求参数: site_id=%d, user_id=%d, username=%s, currency=%s",
		req.SiteId, req.UserId, req.Username, req.Currency)

	// 参数验证
	if req.SiteId <= 0 {
		g.Log().Errorf(ctx, "站点ID无效: %d", req.SiteId)
		return &v1.InitBalanceUserRes{
			Success: false,
			Message: "站点ID无效",
		}, nil
	}

	// 兜底设置动态数据库组（MQ消费等场景没有metadata时）
	if ctx.Value("dynamic_db_group") == nil {
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(req.SiteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	if req.UserId <= 0 {
		g.Log().Errorf(ctx, "用户ID无效: %d", req.UserId)
		return &v1.InitBalanceUserRes{
			Success: false,
			Message: "用户ID无效",
		}, nil
	}

	if req.Username == "" {
		g.Log().Error(ctx, "用户名不能为空")
		return &v1.InitBalanceUserRes{
			Success: false,
			Message: "用户名不能为空",
		}, nil
	}

	if req.Currency == "" {
		req.Currency = "USD" // 默认货币
	}

	g.Log().Infof(ctx, "参数验证通过，开始查询用户余额是否存在")

	// 检查用户余额是否已存在
	var existingBalance *entity.BalanceUser
	err = dao.BalanceUser.Ctx(ctx).
		Where("site_id", req.SiteId).
		Where("user_id", req.UserId).
		Scan(&existingBalance)

	if err != nil {
		errMsg := fmt.Sprintf("查询用户余额失败: %v", err)
		g.Log().Errorf(ctx, errMsg)
		return &v1.InitBalanceUserRes{
			Success: false,
			Message: errMsg,
		}, nil
	}

	// 如果已存在，返回成功
	if existingBalance != nil {
		g.Log().Infof(ctx, "用户余额已存在: site_id=%d, user_id=%d", req.SiteId, req.UserId)
		return &v1.InitBalanceUserRes{
			Success: true,
			Message: "用户余额已存在",
		}, nil
	}

	g.Log().Infof(ctx, "用户余额不存在，开始创建新记录")

	// 创建新的用户余额记录
	balanceUser := &entity.BalanceUser{
		SiteId:        int(req.SiteId),
		UserId:        int(req.UserId),
		Username:      req.Username,
		Balance:       0.00,
		BalanceFrozen: 0.00,
		BalanceDebt:   0.00,
		Points:        0,
		BalanceStatus: balanceConsts.BalanceStatusNormal, // 正常
		Currency:      req.Currency,
		CreatedAt:     gtime.Now(),
		UpdatedAt:     gtime.Now(),
	}

	result, err := dao.BalanceUser.Ctx(ctx).Insert(balanceUser)
	if err != nil {
		errMsg := fmt.Sprintf("插入用户余额失败: %v", err)
		g.Log().Errorf(ctx, errMsg)
		return &v1.InitBalanceUserRes{
			Success: false,
			Message: errMsg,
		}, nil
	}

	rowsAffected, _ := result.RowsAffected()
	g.Log().Infof(ctx, "用户余额初始化成功: site_id=%d, user_id=%d, username=%s, rows_affected=%d",
		req.SiteId, req.UserId, req.Username, rowsAffected)

	// 尝试发放注册送彩金（不影响主流程）
	s.tryGrantRegisterBonus(ctx, req.SiteId, req.UserId)

	return &v1.InitBalanceUserRes{
		Success: true,
		Message: "用户余额初始化成功",
	}, nil
}

// RefreshBalance 刷新用户余额
func (s *sBalance) RefreshBalance(ctx context.Context, req *v1.RefreshBalanceReq) (res *v1.RefreshBalanceRes, err error) {

	// 从JWT token获取用户ID
	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return nil, fmt.Errorf("未登录")
	}

	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	g.Log().Infof(ctx, "刷新用户余额: site_id=%d, user_id=%d", siteId, userId)

	// 查询用户最后一次成功进入游戏的game_id
	lastGameRes, err := s.GetLastBalanceLogGame(ctx, &v1.GetLastBalanceLogGameReq{
		SiteId: int32(siteId),
		UserId: int32(userId),
	})
	if err != nil {
		g.Log().Errorf(ctx, "查询最后一次游戏记录失败: %v", err)
		// 不影响主流程，继续查询余额
	}

	var totalBalance float64
	var gameBalance float64

	// 查询用户余额
	var balanceUser *entity.BalanceUser
	err = dao.BalanceUser.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Scan(&balanceUser)

	if err != nil {
		g.Log().Errorf(ctx, "查询用户余额失败: %v", err)
		return nil, fmt.Errorf("获取余额失败")
	}

	// 如果用户余额记录不存在，返回默认值
	if balanceUser == nil {
		g.Log().Infof(ctx, "用户余额记录不存在: site_id=%d, user_id=%d", siteId, userId)
		return &v1.RefreshBalanceRes{
			Balance:       "0.0000",
			BalanceFrozen: "0.0000",
			Points:        0,
		}, nil
	}

	// 本地余额
	localBalance := balanceUser.Balance
	totalBalance = localBalance

	// 如果有最后一次游戏记录且game_id>0，则需要查询三方游戏余额
	if lastGameRes != nil && lastGameRes.GameId > 0 {
		g.Log().Infof(ctx, "用户最后一次进入游戏: game_id=%d", lastGameRes.GameId)

		// 获取用户信息
		username := middleware.GetUsernameFromContext(ctx)
		siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
		currency := middleware.GetCurrencyFromContext(ctx)
		if currency == "" {
			currency = "MMK(K)"
		}

		g.Log().Infof(ctx, "准备调用 game 服务: user_id=%d, site_id=%d, username=%s, site_code=%s, currency=%s",
			userId, siteId, username, siteCode, currency)

		// 调用 game 服务查询三方游戏余额
		gameClient, conn, err := grpc_client.GetGameClient(ctx)
		if err != nil {
			g.Log().Errorf(ctx, "获取 Game 客户端失败: %v", err)
			// 不影响主流程，继续返回本地余额
		} else {
			defer conn.Close() // 确保连接关闭

			// 创建新的 context，添加用户信息到 metadata
			md := metadata.New(map[string]string{
				"user_id":   gconv.String(userId),
				"site_id":   gconv.String(siteId),
				"username":  username,
				"site_code": siteCode,
				"currency":  currency,
			})
			ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

			// 调用 GetGameBalance 方法
			gameBalanceRes, err := gameClient.GetGameBalance(ctxWithMetadata, &gamev1.GetGameBalanceReq{
				GameId: lastGameRes.GameId,
			})

			if err != nil {
				g.Log().Errorf(ctx, "查询三方游戏余额失败: %v", err)
				// 不影响主流程，继续返回本地余额
			} else if gameBalanceRes != nil && gameBalanceRes.Success {
				gameBalance = gameBalanceRes.Balance
				g.Log().Infof(ctx, "查询三方游戏余额成功: game_id=%d, balance=%.4f", lastGameRes.GameId, gameBalance)
			} else {
				g.Log().Warningf(ctx, "查询三方游戏余额失败: %s", gameBalanceRes.Message)
			}
		}
	} else {
		g.Log().Infof(ctx, "用户没有进入过游戏或game_id=0，只返回本地余额")
	}

	// 总余额 = 本地余额 + 三方游戏余额
	totalBalance = localBalance + gameBalance

	g.Log().Infof(ctx, "刷新用户余额成功: user_id=%d, local_balance=%.4f, game_balance=%.4f, total_balance=%.4f, balance_frozen=%.4f, points=%d",
		userId, localBalance, gameBalance, totalBalance, balanceUser.BalanceFrozen, balanceUser.Points)

	// 格式化余额为字符串，保留数据库原始精度（4位小数）
	// 使用向下取整避免四舍五入，先乘以10000转为整数，再格式化
	balanceInt := int64(totalBalance * 10000)
	balanceStr := fmt.Sprintf("%d.%04d", balanceInt/10000, balanceInt%10000)

	balanceFrozenInt := int64(balanceUser.BalanceFrozen * 10000)
	balanceFrozenStr := fmt.Sprintf("%d.%04d", balanceFrozenInt/10000, balanceFrozenInt%10000)

	res = &v1.RefreshBalanceRes{
		Balance:       balanceStr,
		BalanceFrozen: balanceFrozenStr,
		Points:        int64(balanceUser.Points),
	}

	return res, nil
}

// AdjustBalance 调整用户余额（加款/扣款）
func (s *sBalance) AdjustBalance(ctx context.Context, req *v1.AdjustBalanceReq) (res *v1.AdjustBalanceRes, err error) {
	g.Log().Infof(ctx, "========== Balance服务收到AdjustBalance请求 ==========")
	g.Log().Infof(ctx, "请求参数: site_id=%d, user_id=%d, type=%d, game_name=%s, game_id=%s",
		req.SiteId, req.UserId, req.Type, req.GameName, req.GameId)

	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "AdjustBalance发生panic: %v", r)
			res = &v1.AdjustBalanceRes{
				Success: false,
				Message: fmt.Sprintf("处理失败: %v", r),
			}
			err = nil
		}
	}()

	// 参数验证
	if req.UserId <= 0 {
		g.Log().Errorf(ctx, "用户ID无效: %d", req.UserId)
		return &v1.AdjustBalanceRes{
			Success: false,
			Message: "用户ID无效",
		}, nil
	}

	// 判断是加款还是扣款
	if req.Type == 0 {
		// 扣款 - 转出：从中心账户转账到游戏账户
		g.Log().Infof(ctx, "执行扣款操作（转出到游戏）")
		return s.transferToGame(ctx, req)
	} else {
		// 加款 - 转入：从游戏账户转账到中心账户
		g.Log().Infof(ctx, "执行加款操作（从游戏转入）")
		return s.transferFromGame(ctx, req)
	}
}

//transferToGame
/**
 * @desc：扣款-转出：从中心账户转账到游戏账户（需要配合RabbitMQ和游戏API）
 * @param ctx
 * @param siteId
 * @param req
 * @param amount
 * @return *v1.AdjustBalanceRes
 * @return error
 * @author : Carson
 */
func (s *sBalance) transferToGame(ctx context.Context, req *v1.AdjustBalanceReq) (*v1.AdjustBalanceRes, error) {
	var newBalance float64
	var newBalanceFrozen float64
	var username string
	amount := float64(req.Amount) // 从请求中获取金额

	//操作类型
	tradeType := balanceConsts.LOG_TYPE_TRANSFER_OUT
	//生成流水号
	tradeNo := util.GenerateTradeNo(ctx, tradeType)

	// 事务一：扣除余额、创建pending=1状态的记录
	err := dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询用户余额（加锁）
		var balanceUser *entity.BalanceUser
		err := dao.BalanceUser.Ctx(ctx).
			Where("site_id", req.SiteId).
			Where("user_id", req.UserId).
			LockUpdate().
			Scan(&balanceUser)

		if err != nil {
			return fmt.Errorf("查询用户余额失败")
		}

		if balanceUser == nil {
			return fmt.Errorf("用户余额记录不存在")
		}

		username = balanceUser.Username

		// 检查余额是否足够（四舍五入到2位小数后比较，避免浮点数精度问题）
		balanceRounded := float64(int(balanceUser.Balance*100+0.5)) / 100
		amountRounded := float64(int(amount*100+0.5)) / 100

		if balanceRounded < amountRounded {
			return fmt.Errorf("余额不足，当前余额: %.2f, 需要: %.2f", balanceUser.Balance, amount)
		}

		// 计算新余额：扣除余额，增加冻结金额
		newBalance = balanceUser.Balance - amount
		if newBalance < 0 {
			newBalance = 0 // 避免出现负数
		}
		newBalanceFrozen = balanceUser.BalanceFrozen + amount

		// 更新余额
		_, err = dao.BalanceUser.Ctx(ctx).
			Where("site_id", req.SiteId).
			Where("user_id", req.UserId).
			Update(g.Map{
				"balance":        newBalance,
				"balance_frozen": newBalanceFrozen,
				"updated_at":     gtime.Now(),
			})

		if err != nil {
			return fmt.Errorf("更新余额失败")
		}

		// 记录游戏日志表（status=1待确认）
		gameLog := &entity.BalanceLogGame{
			SiteId:    int(req.SiteId),
			UserId:    int(req.UserId),
			GameId:    gconv.Int(req.GameId),      // 可以从req中传入
			InOrOut:   balanceConsts.TRANSFER_OUT, // 0=转出
			TradeNo:   tradeNo,
			Money:     amount,
			Status:    balanceConsts.StatusPending, // 1=待确认
			Key:       req.GameKey,                 // 保存游戏key
			Remark:    "中心账户 -> " + req.GameName,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}

		_, err = dao.BalanceLogGame.Ctx(ctx).Insert(gameLog)
		if err != nil {
			return fmt.Errorf("记录游戏日志失败")
		}

		// 记录资金日志表（status=1待处理）
		balanceLog := &entity.BalanceLog{
			SiteId:    int(req.SiteId),
			UserId:    int(req.UserId),
			TradeType: tradeType,
			GameId:    gconv.Int(req.GameId),
			TradeNo:   tradeNo,
			Money:     amount,
			Status:    balanceConsts.StatusPending, // 1=处理中
			Remark:    "中心账户 -> " + req.GameName,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}

		_, err = dao.BalanceLog.Ctx(ctx).Insert(balanceLog)
		if err != nil {
			return fmt.Errorf("记录资金日志失败")
		}

		// 记录账变表（status=1待审核）
		balanceChange := &entity.BalanceChange{
			SiteId:        int(req.SiteId),
			UserId:        int(req.UserId),
			Username:      username,
			TradeType:     uint(tradeType),
			TradeNo:       tradeNo,
			BalanceOld:    balanceUser.Balance,
			Money:         amount,
			BalanceNew:    newBalance,
			BalanceFrozen: newBalanceFrozen,
			Status:        balanceConsts.StatusPending, // 1=待审核
			Remark:        "（未确认）中心账户 -> " + req.GameName,
			CreatedAt:     gtime.Now(),
			UpdatedAt:     gtime.Now(),
		}

		_, err = dao.BalanceChange.Ctx(ctx).Insert(balanceChange)
		if err != nil {
			return fmt.Errorf("记录账变日志失败")
		}

		return nil
	})

	if err != nil {
		return &v1.AdjustBalanceRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	balance := fmt.Sprintf("%.2f", newBalance)

	return &v1.AdjustBalanceRes{
		Success: true,
		Message: "扣款请求已提交，等待处理",
		Balance: balance,
		OrderNo: tradeNo,
	}, nil
}

// UpdateTransferStatus 更新转账状态（事务二：成功）
// 由game服务在三方游戏API转账成功后调用
func (s *sBalance) UpdateTransferStatus(ctx context.Context, req *v1.UpdateTransferStatusReq) (*v1.UpdateTransferStatusRes, error) {
	orderNo := req.OrderNo

	err := dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		//幂等判断
		var balanceLog *entity.BalanceLog
		err := dao.BalanceLog.Ctx(ctx).
			Where("trade_no", orderNo).
			Scan(&balanceLog)
		if err != nil {
			return err
		}
		if balanceLog == nil {
			return fmt.Errorf("订单不存在: %s", orderNo)
		}
		if balanceLog.Status == balanceConsts.StatusSuccess || balanceLog.Status == balanceConsts.StatusFail {
			return nil // 幂等：已经处理过，直接认为成功
		}

		siteId := balanceLog.SiteId
		userId := balanceLog.UserId
		amount := balanceLog.Money
		tradeType := balanceLog.TradeType

		// 查询用户余额（加锁）
		var balanceUser *entity.BalanceUser
		err = dao.BalanceUser.Ctx(ctx).
			Where("site_id", siteId).
			Where("user_id", userId).
			LockUpdate().
			Scan(&balanceUser)

		if err != nil {
			return fmt.Errorf("查询用户余额失败")
		}

		if balanceUser == nil {
			return fmt.Errorf("用户余额记录不存在")
		}

		// 根据交易类型区分处理
		var newBalance float64
		var newBalanceFrozen float64
		var remark string

		if tradeType == balanceConsts.LOG_TYPE_TRANSFER_OUT {
			// Type=0: 转出到游戏（扣款）- 减少冻结金额
			g.Log().Infof(ctx, "处理转出成功: 减少冻结金额 %.2f", amount)

			// 检查冻结金额是否足够（四舍五入到2位小数后比较，避免浮点数精度问题）
			frozenRounded := float64(int(balanceUser.BalanceFrozen*100+0.5)) / 100
			amountRounded := float64(int(amount*100+0.5)) / 100

			if frozenRounded < amountRounded {
				return fmt.Errorf("用户冻结金额不足")
			}

			newBalance = balanceUser.Balance
			newBalanceFrozen = balanceUser.BalanceFrozen - amount
			if newBalanceFrozen < 0 {
				newBalanceFrozen = 0 // 避免出现负数
			}
			remark = "（已确认）成功"

		} else if tradeType == balanceConsts.LOG_TYPE_TRANSFER_IN {
			// Type=1: 从游戏转入（加款）- 增加用户余额
			g.Log().Infof(ctx, "处理转入成功: 增加用户余额 %.2f", amount)

			newBalance = balanceUser.Balance + amount
			newBalanceFrozen = balanceUser.BalanceFrozen
			remark = "（已确认）成功"

		} else {
			return fmt.Errorf("未知的交易类型: %d", tradeType)
		}

		// 更新余额
		_, err = dao.BalanceUser.Ctx(ctx).
			Where("site_id", siteId).
			Where("user_id", userId).
			Update(g.Map{
				"balance":        newBalance,
				"balance_frozen": newBalanceFrozen,
				"updated_at":     gtime.Now(),
			})

		if err != nil {
			return fmt.Errorf("更新余额失败")
		}

		// 更新游戏日志状态
		_, err = dao.BalanceLogGame.Ctx(ctx).
			Where("trade_no", orderNo).
			Update(g.Map{
				"status":     balanceConsts.StatusSuccess, // 3=成功
				"updated_at": gtime.Now(),
			})
		if err != nil {
			return err
		}

		// 更新资金日志状态
		_, err = dao.BalanceLog.Ctx(ctx).
			Where("trade_no", orderNo).
			Update(g.Map{
				"status":     balanceConsts.StatusSuccess, // 3=成功
				"updated_at": gtime.Now(),
			})
		if err != nil {
			return err
		}

		// 更新账变表状态
		_, err = dao.BalanceChange.Ctx(ctx).
			Where("trade_no", orderNo).
			Update(g.Map{
				"balance_new":    newBalance,
				"balance_frozen": newBalanceFrozen,
				"remark":         remark,
				"status":         balanceConsts.StatusSuccess, // 3=成功
				"updated_at":     gtime.Now(),
			})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &v1.UpdateTransferStatusRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.UpdateTransferStatusRes{
		Success: true,
		Message: "更新转账状态成功",
	}, nil
}

// CompensateBalance 补偿余额（事务三：失败）
// 由game服务在三方游戏API转账失败后调用
// 转入失败：需要补偿余额（amount > 0）
// 转出失败：只标记失败（amount = 0）
func (s *sBalance) CompensateBalance(ctx context.Context, req *v1.CompensateBalanceReq) (*v1.CompensateBalanceRes, error) {
	siteId := req.SiteId
	userId := req.UserId
	orderNo := req.OrderNo
	amount := float64(req.Amount)

	err := dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

		//幂等判断
		var balanceLog *entity.BalanceLog
		err := dao.BalanceLog.Ctx(ctx).
			Where("trade_no", orderNo).
			Scan(&balanceLog)
		if err != nil {
			return err
		}
		if balanceLog == nil {
			return fmt.Errorf("订单不存在: %s", orderNo)
		}
		if balanceLog.Status == balanceConsts.StatusSuccess || balanceLog.Status == balanceConsts.StatusFail {
			return nil // 幂等：已经处理过，直接认为成功
		}
		// 必须是 pending 才能处理
		if balanceLog.Status != balanceConsts.StatusPending {
			return nil
		}

		tradeType := balanceLog.TradeType
		realAmount := balanceLog.Money
		realSiteId := balanceLog.SiteId
		realUserId := balanceLog.UserId

		// 如果传入的amount为0，使用订单中的金额
		if amount == 0 {
			amount = realAmount
		}

		// 四舍五入到2位小数，避免浮点数精度问题
		amount = float64(int(amount*100+0.5)) / 100
		realAmount = float64(int(realAmount*100+0.5)) / 100

		// 如果传入的siteId为0，使用订单中的siteId
		if siteId == 0 {
			siteId = int32(realSiteId)
		}
		// 如果传入的userId为0，使用订单中的userId
		if userId == 0 {
			userId = int32(realUserId)
		}

		// 查询用户余额（加锁）
		var balanceUser *entity.BalanceUser
		err = dao.BalanceUser.Ctx(ctx).
			Where("site_id", siteId).
			Where("user_id", userId).
			LockUpdate().
			Scan(&balanceUser)

		if err != nil {
			return err
		}

		if balanceUser == nil {
			return fmt.Errorf("用户余额记录不存在")
		}

		var newBalance float64
		var newBalanceFrozen float64
		var remark string

		if tradeType == balanceConsts.LOG_TYPE_TRANSFER_OUT {
			// 转入失败（Type=0）：需要补偿余额
			g.Log().Infof(ctx, "处理转入失败: 补偿余额 %.2f", amount)

			// 补偿余额（加回去）
			newBalance = balanceUser.Balance + amount

			// 检查冻结金额是否足够（四舍五入到2位小数后比较，避免浮点数精度问题）
			frozenRounded := float64(int(balanceUser.BalanceFrozen*100+0.5)) / 100
			amountRounded := float64(int(amount*100+0.5)) / 100

			if frozenRounded < amountRounded {
				return fmt.Errorf("用户冻结金额不足")
			}

			newBalanceFrozen = balanceUser.BalanceFrozen - amount
			if newBalanceFrozen < 0 {
				newBalanceFrozen = 0 // 避免出现负数
			}
			remark = "（已确认）金额补偿"

		} else if tradeType == balanceConsts.LOG_TYPE_TRANSFER_IN {
			// 转出失败（Type=1）：只标记失败，不动余额
			g.Log().Infof(ctx, "处理转出失败: 只标记失败，不动余额")

			newBalance = balanceUser.Balance
			newBalanceFrozen = balanceUser.BalanceFrozen
			remark = "（已确认）转出失败"

		} else {
			return fmt.Errorf("未知的交易类型: %d", tradeType)
		}

		_, err = dao.BalanceUser.Ctx(ctx).
			Where("site_id", siteId).
			Where("user_id", userId).
			Update(g.Map{
				"balance":        newBalance,
				"balance_frozen": newBalanceFrozen,
				"updated_at":     gtime.Now(),
			})
		if err != nil {
			return err
		}

		// 更新游戏日志状态为失败
		_, err = dao.BalanceLogGame.Ctx(ctx).
			Where("trade_no", orderNo).
			Update(g.Map{
				"status":     balanceConsts.StatusFail, // 3=失败
				"updated_at": gtime.Now(),
			})
		if err != nil {
			return err
		}

		// 更新资金日志状态为失败
		_, err = dao.BalanceLog.Ctx(ctx).
			Where("trade_no", orderNo).
			Update(g.Map{
				"status":     balanceConsts.StatusFail, // 3=失败
				"updated_at": gtime.Now(),
			})
		if err != nil {
			return err
		}

		// 更新账变表状态为失败
		_, err = dao.BalanceChange.Ctx(ctx).
			Where("trade_no", orderNo).
			Update(g.Map{
				"balance_new":    newBalance,
				"balance_frozen": newBalanceFrozen,
				"remark":         remark,
				"status":         balanceConsts.StatusFail, // 3=失败
				"updated_at":     gtime.Now(),
			})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &v1.CompensateBalanceRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.CompensateBalanceRes{
		Success: true,
		Message: "处理失败订单成功",
	}, nil
}

//GetLastBalanceLogGame
/**
 * @desc：获取最后一次成功进入游戏的game_id和用户余额
 * @param ctx
 * @param req
 * @return *v1.GetLastBalanceLogGameRes
 * @return error
 * @author : Carson
 */
func (s *sBalance) GetLastBalanceLogGame(ctx context.Context, req *v1.GetLastBalanceLogGameReq) (*v1.GetLastBalanceLogGameRes, error) {

	siteId := req.SiteId
	userId := req.UserId

	// 查询用户余额
	var balanceUser *entity.BalanceUser
	err := dao.BalanceUser.Ctx(ctx).
		Where("site_id", req.SiteId).
		Where("user_id", req.UserId).
		Scan(&balanceUser)

	if err != nil {
		return nil, fmt.Errorf("查询用户余额失败")
	}

	// 如果用户余额记录不存在，返回game_id=0，余额为0
	if balanceUser == nil {
		g.Log().Infof(ctx, "用户余额记录不存在: site_id=%d, user_id=%d, 返回默认值", req.SiteId, req.UserId)
		return &v1.GetLastBalanceLogGameRes{
			GameId: 0,
			Amount: 0,
		}, nil
	}

	// 查询最后一次成功的游戏记录（status=2）
	var last *entity.BalanceLogGame
	err = dao.BalanceLogGame.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", balanceConsts.StatusSuccess).   // 只查询成功的记录
		Where("in_or_out", balanceConsts.TRANSFER_OUT). // 转出：从中心账户转出到游戏账户
		OrderDesc("id").
		Limit(1).
		Scan(&last)
	if err != nil {
		return nil, err
	}

	// 如果没有游戏记录，返回game_id=0，余额为用户中心钱包余额
	if last == nil {
		return &v1.GetLastBalanceLogGameRes{
			GameId: 0,
			Amount: float32(balanceUser.Balance),
		}, nil
	}

	// 如果有游戏记录，返回最后一次游戏的game_id，余额为用户中心钱包余额
	return &v1.GetLastBalanceLogGameRes{
		GameId: int32(last.GameId),
		Amount: float32(balanceUser.Balance),
	}, nil
}

//===================================================================================================
//transferFromGame
/**
 * @desc：加款-转入：从游戏账户转账到中心账户
 * @param ctx
 * @param siteId
 * @param req
 * @param amount
 * @return *v1.AdjustBalanceRes
 * @return error
 * @author : Carson
 */
func (s *sBalance) transferFromGame(ctx context.Context, req *v1.AdjustBalanceReq) (*v1.AdjustBalanceRes, error) {
	var username string
	amount := float64(req.Amount)

	//操作类型
	tradeType := balanceConsts.LOG_TYPE_TRANSFER_IN
	//生成流水号
	tradeNo := util.GenerateTradeNo(ctx, tradeType)

	// 开启事务：只创建pending订单，不修改用户余额
	err := dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询用户余额（加锁）
		var balanceUser *entity.BalanceUser
		err := dao.BalanceUser.Ctx(ctx).
			Where("site_id", req.SiteId).
			Where("user_id", req.UserId).
			LockUpdate().
			Scan(&balanceUser)

		if err != nil {
			return fmt.Errorf("查询用户余额失败")
		}

		if balanceUser == nil {
			return fmt.Errorf("用户余额记录不存在")
		}

		username = balanceUser.Username

		// 注意：转出时不修改用户余额，等三方API成功后再加余额

		// 记录游戏日志表（status=1待处理）
		gameLog := &entity.BalanceLogGame{
			SiteId:    int(req.SiteId),
			UserId:    int(req.UserId),
			GameId:    gconv.Int(req.GameId),
			InOrOut:   balanceConsts.TRANSFER_IN, // 1=转入
			TradeNo:   tradeNo,
			Money:     amount,
			Status:    balanceConsts.StatusPending, // 1=pending
			Remark:    req.GameName + " -> 中心账户",
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}

		_, err = dao.BalanceLogGame.Ctx(ctx).Insert(gameLog)
		if err != nil {
			return fmt.Errorf("记录游戏日志失败")
		}

		// 记录资金日志表（status=1待处理）
		balanceLog := &entity.BalanceLog{
			SiteId:    int(req.SiteId),
			UserId:    int(req.UserId),
			TradeType: tradeType,
			GameId:    gconv.Int(req.GameId),
			TradeNo:   tradeNo,
			Money:     amount,
			Status:    balanceConsts.StatusPending, // 1=处理中
			Remark:    req.GameName + " -> 中心账户",
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}

		_, err = dao.BalanceLog.Ctx(ctx).Insert(balanceLog)
		if err != nil {
			return fmt.Errorf("记录资金日志失败")
		}

		// 记录账变表（status=1待审核）
		balanceChange := &entity.BalanceChange{
			SiteId:        int(req.SiteId),
			UserId:        int(req.UserId),
			Username:      username,
			TradeType:     uint(tradeType),
			TradeNo:       tradeNo,
			BalanceOld:    balanceUser.Balance,
			Money:         amount,
			BalanceNew:    balanceUser.Balance, // 暂时不变，等成功后再加
			BalanceFrozen: balanceUser.BalanceFrozen,
			Status:        balanceConsts.StatusPending, // 1=待审核
			Remark:        "（未确认）" + req.GameName + " -> 中心账户",
			CreatedAt:     gtime.Now(),
			UpdatedAt:     gtime.Now(),
		}

		_, err = dao.BalanceChange.Ctx(ctx).Insert(balanceChange)
		if err != nil {
			return fmt.Errorf("记录账变日志失败")
		}

		return nil
	})

	if err != nil {
		return &v1.AdjustBalanceRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.AdjustBalanceRes{
		Success: true,
		Message: "转出请求已提交，等待处理",
		Balance: "0.00", // 余额暂时不变
		OrderNo: tradeNo,
	}, nil
}

// GetPaymentList 获取支付渠道列表
func (s *sBalance) GetPaymentList(ctx context.Context, req *v1.GetPaymentListReq) (*v1.GetPaymentListRes, error) {
	// 从context获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	g.Log().Infof(ctx, "查询支付渠道列表: site_id=%d", siteId)

	// 查询启用状态的支付渠道，按sort从大到小排序
	var payments []*entity.SitePayment
	err := dao.SitePayment.Ctx(ctx).
		Where("site_id", siteId).
		Where("status", balanceConsts.ChannelStatusEnabled).
		OrderDesc("sort").
		Scan(&payments)

	if err != nil {
		g.Log().Errorf(ctx, "查询支付渠道失败: %v", err)
		return nil, fmt.Errorf("查询支付渠道失败")
	}

	// 转换为响应格式
	list := make([]*v1.PaymentInfo, 0, len(payments))
	for _, p := range payments {
		// 解析moneyList字段（用逗号分隔）
		var moneyList []string
		if p.MoneyList != "" {
			moneyList = strings.Split(p.MoneyList, ",")
		}

		list = append(list, &v1.PaymentInfo{
			Id:        int32(p.Id),
			SiteId:    int32(p.SiteId),
			PaymentId: int32(p.PaymentId),
			Name:      p.Name,
			Code:      p.Code,
			Gateway:   int32(p.Gateway),
			EachMin:   float32(p.EachMin),
			EachMax:   float32(p.EachMax),
			Sort:      int32(p.Sort),
			MoneyList: moneyList,
			IsInput:   int32(p.IsInput),
			Remark:    p.Remark,
			Logo:      p.Logo,
			Status:    int32(p.Status),
		})
	}

	g.Log().Infof(ctx, "查询支付渠道成功: 共%d条", len(list))

	return &v1.GetPaymentListRes{
		List: list,
	}, nil
}

// GetWithdrawList 获取提现渠道列表
func (s *sBalance) GetWithdrawList(ctx context.Context, req *v1.GetWithdrawListReq) (*v1.GetWithdrawListRes, error) {
	// 从context获取站点ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	g.Log().Infof(ctx, "查询提现渠道列表: site_id=%d", siteId)

	// 查询启用状态的提现渠道，按sort从大到小排序
	var withdraws []*entity.SiteWithdraw
	err := dao.SiteWithdraw.Ctx(ctx).
		Where("site_id", siteId).
		Where("status", balanceConsts.ChannelStatusEnabled).
		OrderDesc("sort").
		Scan(&withdraws)

	if err != nil {
		g.Log().Errorf(ctx, "查询提现渠道失败: %v", err)
		return nil, fmt.Errorf("查询提现渠道失败")
	}

	// 转换为响应格式
	list := make([]*v1.WithdrawInfo, 0, len(withdraws))
	for _, w := range withdraws {
		// 解析moneyList字段（用逗号分隔）
		var moneyList []string
		if w.MoneyList != "" {
			moneyList = strings.Split(w.MoneyList, ",")
		}

		list = append(list, &v1.WithdrawInfo{
			Id:         int32(w.Id),
			SiteId:     int32(w.SiteId),
			WithdrawId: int32(w.WithdrawId),
			Name:       w.Name,
			Code:       w.Code,
			Gateway:    int32(w.Gateway),
			EachMin:    float32(w.EachMin),
			EachMax:    float32(w.EachMax),
			Sort:       int32(w.Sort),
			MoneyList:  moneyList,
			IsInput:    int32(w.IsInput),
			Remark:     w.Remark,
			Logo:       w.Logo,
			Status:     int32(w.Status),
		})
	}

	g.Log().Infof(ctx, "查询提现渠道成功: 共%d条", len(list))

	return &v1.GetWithdrawListRes{
		List: list,
	}, nil
}

// AddBankCard 添加银行卡
func (s *sBalance) AddBankCard(ctx context.Context, req *v1.AddBankCardReq) (*v1.AddBankCardRes, error) {
	// 从context获取站点ID和用户ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return &v1.AddBankCardRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}

	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return &v1.AddBankCardRes{
			Success: false,
			Message: "未获取用户ID",
		}, nil
	}

	g.Log().Infof(ctx, "添加银行卡: site_id=%d, user_id=%d, type=%d", siteId, userId, req.Type)

	// 参数验证
	if req.Type != 1 && req.Type != 2 {
		return &v1.AddBankCardRes{
			Success: false,
			Message: "类型无效，必须是1(GCASH)或2(Maya)",
		}, nil
	}

	if req.CardAccount == "" {
		return &v1.AddBankCardRes{
			Success: false,
			Message: "账户名不能为空",
		}, nil
	}

	if req.CardNo == "" {
		return &v1.AddBankCardRes{
			Success: false,
			Message: "卡号不能为空",
		}, nil
	}

	// 检查该类型是否已存在银行卡
	var existingCard *entity.UserBank
	err := dao.UserBank.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("type", req.Type).
		Scan(&existingCard)

	if err != nil {
		g.Log().Errorf(ctx, "查询银行卡失败: %v", err)
		return &v1.AddBankCardRes{
			Success: false,
			Message: "查询银行卡失败",
		}, nil
	}

	// 如果该类型已存在银行卡，返回错误
	if existingCard != nil {
		typeName := "GCASH"
		if req.Type == 2 {
			typeName = "Maya"
		}
		return &v1.AddBankCardRes{
			Success: false,
			Message: fmt.Sprintf("%s类型的银行卡已存在，每种类型只能添加一张", typeName),
		}, nil
	}

	// 新增银行卡时，先将所有银行卡的默认状态改为0
	_, err = dao.UserBank.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Update(g.Map{"is_default": 0})

	if err != nil {
		g.Log().Errorf(ctx, "更新默认状态失败: %v", err)
		return &v1.AddBankCardRes{
			Success: false,
			Message: "更新默认状态失败",
		}, nil
	}

	// 插入新银行卡，is_default固定为1
	bankCard := &entity.UserBank{
		SiteId:      int(siteId),
		UserId:      int(userId),
		Type:        int(req.Type),
		CardAccount: req.CardAccount,
		CardNo:      req.CardNo,
		IsDefault:   1, // 新增的银行卡默认为默认卡
		CreatedAt:   gtime.Now(),
		UpdatedAt:   gtime.Now(),
	}

	_, err = dao.UserBank.Ctx(ctx).Insert(bankCard)
	if err != nil {
		g.Log().Errorf(ctx, "添加银行卡失败: %v", err)
		return &v1.AddBankCardRes{
			Success: false,
			Message: "添加银行卡失败",
		}, nil
	}

	g.Log().Infof(ctx, "添加银行卡成功")

	return &v1.AddBankCardRes{
		Success: true,
		Message: "添加成功",
	}, nil
}

// GetBankCardList 查询银行卡列表
func (s *sBalance) GetBankCardList(ctx context.Context, req *v1.GetBankCardListReq) (*v1.GetBankCardListRes, error) {
	// 从context获取站点ID和用户ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}

	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return nil, fmt.Errorf("未获取用户ID")
	}

	g.Log().Infof(ctx, "查询银行卡列表: site_id=%d, user_id=%d", siteId, userId)

	// 查询银行卡列表
	var bankCards []*entity.UserBank
	err := dao.UserBank.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		OrderDesc("is_default").
		OrderDesc("created_at").
		Scan(&bankCards)

	if err != nil {
		g.Log().Errorf(ctx, "查询银行卡列表失败: %v", err)
		return nil, fmt.Errorf("查询银行卡列表失败")
	}

	// 转换为响应格式
	list := make([]*v1.BankCardInfo, 0, len(bankCards))
	for _, card := range bankCards {
		list = append(list, &v1.BankCardInfo{
			Id:          int32(card.Id),
			Type:        int32(card.Type),
			CardAccount: card.CardAccount,
			CardNo:      card.CardNo,
			IsDefault:   int32(card.IsDefault),
			CreatedAt:   card.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	g.Log().Infof(ctx, "查询银行卡列表成功: 共%d条", len(list))

	return &v1.GetBankCardListRes{
		List: list,
	}, nil
}

// DeleteBankCard 删除银行卡
func (s *sBalance) DeleteBankCard(ctx context.Context, req *v1.DeleteBankCardReq) (*v1.DeleteBankCardRes, error) {
	// 从context获取站点ID和用户ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return &v1.DeleteBankCardRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}

	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return &v1.DeleteBankCardRes{
			Success: false,
			Message: "未获取用户ID",
		}, nil
	}

	g.Log().Infof(ctx, "删除银行卡: site_id=%d, user_id=%d, id=%d", siteId, userId, req.Id)

	// 参数验证
	if req.Id <= 0 {
		return &v1.DeleteBankCardRes{
			Success: false,
			Message: "银行卡ID无效",
		}, nil
	}

	// 删除银行卡（只能删除自己的）
	result, err := dao.UserBank.Ctx(ctx).
		Where("id", req.Id).
		Where("site_id", siteId).
		Where("user_id", userId).
		Delete()

	if err != nil {
		g.Log().Errorf(ctx, "删除银行卡失败: %v", err)
		return &v1.DeleteBankCardRes{
			Success: false,
			Message: "删除银行卡失败",
		}, nil
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return &v1.DeleteBankCardRes{
			Success: false,
			Message: "银行卡不存在或无权删除",
		}, nil
	}

	g.Log().Infof(ctx, "删除银行卡成功")

	return &v1.DeleteBankCardRes{
		Success: true,
		Message: "删除成功",
	}, nil
}

// SetDefaultBankCard 设置默认银行卡
func (s *sBalance) SetDefaultBankCard(ctx context.Context, req *v1.SetDefaultBankCardReq) (*v1.SetDefaultBankCardRes, error) {
	// 从context获取站点ID和用户ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return &v1.SetDefaultBankCardRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}

	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return &v1.SetDefaultBankCardRes{
			Success: false,
			Message: "未获取用户ID",
		}, nil
	}

	g.Log().Infof(ctx, "设置默认银行卡: site_id=%d, user_id=%d, id=%d", siteId, userId, req.Id)

	// 参数验证
	if req.Id <= 0 {
		return &v1.SetDefaultBankCardRes{
			Success: false,
			Message: "银行卡ID无效",
		}, nil
	}

	// 验证银行卡是否存在且属于当前用户
	var bankCard *entity.UserBank
	err := dao.UserBank.Ctx(ctx).
		Where("id", req.Id).
		Where("site_id", siteId).
		Where("user_id", userId).
		Scan(&bankCard)

	if err != nil {
		g.Log().Errorf(ctx, "查询银行卡失败: %v", err)
		return &v1.SetDefaultBankCardRes{
			Success: false,
			Message: "查询银行卡失败",
		}, nil
	}

	if bankCard == nil {
		return &v1.SetDefaultBankCardRes{
			Success: false,
			Message: "银行卡不存在或无权操作",
		}, nil
	}

	// 先将所有银行卡的默认状态改为0
	_, err = dao.UserBank.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Update(g.Map{"is_default": 0})

	if err != nil {
		g.Log().Errorf(ctx, "更新默认状态失败: %v", err)
		return &v1.SetDefaultBankCardRes{
			Success: false,
			Message: "更新默认状态失败",
		}, nil
	}

	// 将指定银行卡设置为默认
	_, err = dao.UserBank.Ctx(ctx).
		Where("id", req.Id).
		Update(g.Map{"is_default": 1})

	if err != nil {
		g.Log().Errorf(ctx, "设置默认银行卡失败: %v", err)
		return &v1.SetDefaultBankCardRes{
			Success: false,
			Message: "设置默认银行卡失败",
		}, nil
	}

	g.Log().Infof(ctx, "设置默认银行卡成功")

	return &v1.SetDefaultBankCardRes{
		Success: true,
		Message: "设置成功",
	}, nil
}

// GetDepositNonce 获取充值 nonce
func (s *sBalance) GetDepositNonce(ctx context.Context, req *v1.GetDepositNonceReq) (*v1.GetDepositNonceRes, error) {
	// 1. 获取用户信息
	userID, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return nil, gerror.New("未登录")
	}

	siteID, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, gerror.New("未获取站点ID")
	}

	// 2. 生成 nonce
	nonce := uuid.New().String()
	timestamp := time.Now().Unix()
	expiresIn := int64(60) // 60秒过期

	g.Log().Infof(ctx, "生成充值 nonce: %s, timestamp: %d, expiresIn: %d", nonce, timestamp, expiresIn)

	// 3. 使用 NonceService 存储 nonce
	nonceService := &common.NonceService{}
	err := nonceService.Generate(ctx, "deposit", int64(siteID), int64(userID), nonce, expiresIn)
	if err != nil {
		return nil, err
	}

	return &v1.GetDepositNonceRes{
		Nonce:     nonce,
		Timestamp: timestamp,
		ExpiresIn: expiresIn,
	}, nil
}

// GetWithdrawNonce 获取提现 nonce
func (s *sBalance) GetWithdrawNonce(ctx context.Context, req *v1.GetWithdrawNonceReq) (*v1.GetWithdrawNonceRes, error) {
	// 1. 获取用户信息
	userID, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return nil, gerror.New("未登录")
	}

	siteID, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, gerror.New("未获取站点ID")
	}

	// 2. 生成 nonce
	nonce := uuid.New().String()
	timestamp := time.Now().Unix()
	expiresIn := int64(60) // 60秒过期

	g.Log().Infof(ctx, "生成提现 nonce: %s, timestamp: %d, expiresIn: %d", nonce, timestamp, expiresIn)

	// 3. 使用 NonceService 存储 nonce
	nonceService := &common.NonceService{}
	err := nonceService.Generate(ctx, "withdraw", int64(siteID), int64(userID), nonce, expiresIn)
	if err != nil {
		return nil, err
	}

	return &v1.GetWithdrawNonceRes{
		Nonce:     nonce,
		Timestamp: timestamp,
		ExpiresIn: expiresIn,
	}, nil
}

// CreateDepositOrder 创建充值订单
func (s *sBalance) CreateDepositOrder(ctx context.Context, req *v1.CreateDepositOrderReq) (*v1.CreateDepositOrderRes, error) {
	// 1. 获取用户信息
	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return nil, gerror.New("未登录")
	}

	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, gerror.New("未获取站点ID")
	}

	username := middleware.GetUsernameFromContext(ctx)

	//==========防重放验证====================
	// 2. 验证 nonce（防重放）
	nonceService := &common.NonceService{}
	err := nonceService.ValidateAndConsume(ctx, "deposit", int64(siteId), int64(userId), req.Nonce)
	if err != nil {
		return nil, err
	}
	//==========防重放验证====================

	g.Log().Infof(ctx, "CreateDepositOrder 开始: siteId=%d, userId=%d, username=%s, paymentId=%d, money=%.2f",
		siteId, userId, username, req.PaymentId, req.Money)

	// 3. 调用 Domain 层创建充值订单
	depositDomain := &deposit.DepositDomain{}
	domainResult, err := depositDomain.CreateDepositOrder(ctx, &deposit.CreateDepositOrderRequest{
		SiteId:       int64(siteId),
		UserId:       int64(userId),
		Username:     username,
		PaymentId:    int64(req.PaymentId),
		Money:        req.Money,
		OrderType:    req.OrderType,
		BankCode:     req.BankCode,
		Mobile:       int64(req.Mobile),
		UserCardNo:   req.UserCardNo,
		UserCardName: req.UserCardName,
	})

	if err != nil {
		g.Log().Errorf(ctx, "创建充值订单失败: %v", err)
		return nil, err
	}

	// 4. 调用支付平台
	payment := s.getPaymentPlatform(ctx, domainResult.SitePayment.Code)
	if payment == nil {
		return nil, gerror.Newf("不支持的支付平台: %s", domainResult.SitePayment.Code)
	}

	// 构建订单参数
	orderParams := &PaymentOrderParams{
		SiteId:       int(siteId),
		UserId:       int(userId),
		Username:     username,
		PaymentId:    int(req.PaymentId),
		MerchantNo:   domainResult.SitePayment.MerchantNo,
		Md5Key:       domainResult.SitePayment.Md5Key,
		TradeNo:      domainResult.TradeNo,
		Money:        req.Money,
		OrderType:    req.OrderType,
		BankCode:     req.BankCode,
		Mobile:       int(req.Mobile),
		UserCardNo:   req.UserCardNo,
		UserCardName: req.UserCardName,
		Gateway:      domainResult.SitePayment.Gateway,
	}

	result, err := s.createPaymentOrder(ctx, payment, orderParams)
	if err != nil {
		g.Log().Errorf(ctx, "调用支付平台失败: %v", err)
		return nil, err
	}

	// 5. 更新订单信息
	err = depositDomain.UpdateDepositOrderInfo(ctx, domainResult.TradeNo, result.SysTradeNo, result.ChannelCardNo, result.ChannelName, result.ImageUrl)
	if err != nil {
		g.Log().Warningf(ctx, "更新订单信息失败: %v", err)
	}

	g.Log().Infof(ctx, "CreateDepositOrder 成功: tradeNo=%s, sysTradeNo=%s", domainResult.TradeNo, result.SysTradeNo)

	// 6. 返回结果
	return &v1.CreateDepositOrderRes{
		OrderSn:       result.OrderSn,
		SysTradeNo:    result.SysTradeNo,
		ChannelCardNo: result.ChannelCardNo,
		ChannelName:   result.ChannelName,
		ImageUrl:      result.ImageUrl,
		Result:        result.Result,
	}, nil
}

// generateTradeNo 生成交易流水号
func (s *sBalance) generateTradeNo(prefix string) string {
	now := time.Now()
	// 格式: 前缀 + 月日时分秒 + 9位随机数
	// 使用 time.Now().UnixNano() 作为随机种子
	randomNum := now.UnixNano() % 1000000000
	return fmt.Sprintf("%s%s%09d", prefix, now.Format("0102150405"), randomNum)
}

// PaymentOrderParams 支付订单参数
type PaymentOrderParams struct {
	SiteId       int
	UserId       int
	Username     string
	PaymentId    int
	MerchantNo   string
	Md5Key       string
	TradeNo      string
	Money        float64
	OrderType    string
	BankCode     string
	Mobile       int
	UserCardNo   string
	UserCardName string
	Gateway      int
}

// PaymentOrderResult 支付订单结果
type PaymentOrderResult struct {
	Code          int
	OrderSn       string
	SysTradeNo    string
	ChannelCardNo string
	ChannelName   string
	ImageUrl      string
	Result        string
}

// getPaymentPlatform 获取支付平台（简化版，直接返回 MG）
func (s *sBalance) getPaymentPlatform(ctx context.Context, paymentCode string) interface{} {
	// 这里简化处理，直接使用 MG 平台
	// 实际应该通过工厂模式创建
	return &MGPayment{}
}

// createPaymentOrder 创建支付订单
func (s *sBalance) createPaymentOrder(ctx context.Context, payment interface{}, params *PaymentOrderParams) (*PaymentOrderResult, error) {
	// 调用 MG 支付平台
	mgPayment, ok := payment.(*MGPayment)
	if !ok {
		return nil, gerror.New("支付平台类型错误")
	}

	return mgPayment.CreateOrder(ctx, params)
}

// MGPayment MG 支付平台（简化版）
type MGPayment struct{}

// CreateOrder 创建订单

func (m *MGPayment) CreateOrder(ctx context.Context, params *PaymentOrderParams) (*PaymentOrderResult, error) {
	// 导入 MG 支付平台
	// 这里需要调用 internal/logic/frontend/payment/platform/mg 包
	// 由于循环导入问题，这里先用简化实现

	g.Log().Infof(ctx, "MG CreateOrder: tradeNo=%s, money=%.2f", params.TradeNo, params.Money)

	// TODO: 实际调用 MG 支付接口
	// 这里返回模拟数据
	return &PaymentOrderResult{
		Code:          200,
		OrderSn:       params.TradeNo,
		SysTradeNo:    fmt.Sprintf("sys_%s", uuid.New().String()),
		ChannelCardNo: "09678682672",
		ChannelName:   "KBZ",
		ImageUrl:      "http://localhost:19000/uploads/jh/2026/02/2026-02-01 01.40.28_1769881256.jpg",
		Result:        "ok",
	}, nil
}

// GetBalanceLog 获取账户明细
func (s *sBalance) GetBalanceLog(ctx context.Context, req *v1.GetBalanceLogReq) (*v1.GetBalanceLogRes, error) {
	// 从context获取站点ID和用户ID
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return nil, gerror.New("未获取站点ID")
	}

	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return nil, gerror.New("未登录")
	}

	g.Log().Infof(ctx, "查询账户明细: site_id=%d, user_id=%d, start=%s, end=%s, type=%d, status=%d, page=%d, size=%d",
		siteId, userId, req.Start, req.End, req.Type, req.Status, req.Page, req.Size)

	// 参数验证
	if req.Start == "" || req.End == "" {
		return nil, gerror.New("开始时间和结束时间不能为空")
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	// 计算统计数据
	totalRecharge := s.getTotalRecharge(ctx, int(siteId), int(userId), req.Start, req.End)
	totalWithdraw := s.getTotalWithdraw(ctx, int(siteId), int(userId), req.Start, req.End)
	totalBonus := s.getTotalBonus(ctx, int(siteId), int(userId), req.Start, req.End)
	totalRebate := s.getTotalRebate(ctx, int(siteId), int(userId), req.Start, req.End)

	// 构建查询条件
	query := dao.BalanceLog.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		WhereNotIn("trade_type", g.Slice{balanceConsts.LOG_TYPE_TRANSFER_IN, balanceConsts.LOG_TYPE_TRANSFER_OUT})

	// 状态筛选
	if req.Status > 0 {
		query = query.Where("status", req.Status)
	}

	// 类型筛选
	if req.Type > 0 {
		switch req.Type {
		case 1: // 充值
			query = query.WhereIn("trade_type", g.Slice{
				balanceConsts.LOG_TYPE_PAYMENT,
				balanceConsts.LOG_TYPE_TRANSFER,
				balanceConsts.LOG_TYPE_MANUAL_ADD,
			})
		case 2: // 提款
			query = query.WhereIn("trade_type", g.Slice{
				balanceConsts.LOG_TYPE_WITHDRAW,
				balanceConsts.LOG_TYPE_MANUAL_CUT,
			})
		case 3: // 红利
			query = query.Where("trade_type", balanceConsts.LOG_TYPE_BONUS)
		case 4: // 返水
			query = query.Where("trade_type", balanceConsts.LOG_TYPE_REBATE)
		case 5: // 转账
			query = query.WhereIn("trade_type", g.Slice{
				balanceConsts.LOG_TYPE_TRANSFER_IN,
				balanceConsts.LOG_TYPE_TRANSFER_OUT,
			})
		}
	}

	// 时间范围筛选
	if req.Start != "" && req.End != "" {
		query = query.WhereBetween("updated_at", req.Start, req.End)
	}

	// 获取总数
	count, err := query.Count()
	if err != nil {
		g.Log().Errorf(ctx, "查询账户明细总数失败: %v", err)
		return nil, gerror.New("查询失败")
	}

	// 分页查询
	skip := (req.Page - 1) * req.Size
	var balanceLogs []*entity.BalanceLog
	err = query.
		Fields("trade_type", "trade_no", "money", "status", "remark", "updated_at").
		Offset(int(skip)).
		Limit(int(req.Size)).
		OrderDesc("updated_at").
		Scan(&balanceLogs)

	if err != nil {
		g.Log().Errorf(ctx, "查询账户明细列表失败: %v", err)
		return nil, gerror.New("查询失败")
	}

	// 转换为响应格式
	list := make([]*v1.BalanceLogItem, 0, len(balanceLogs))
	for _, log := range balanceLogs {
		// 格式化时间
		timeStr := ""
		if log.UpdatedAt != nil {
			timeStr = log.UpdatedAt.String()
		}

		item := &v1.BalanceLogItem{
			TradeNo: log.TradeNo,
			Money:   log.Money,
			Remark:  log.Remark,
			Type:    s.getFrontTypeByType(log.TradeType),
			Symbol:  s.getSymbol(log.TradeType),
			Status:  int32(log.Status),
			Time:    timeStr,
		}

		// 获取交易类型（转入/转出）
		item.ChangeType = s.getChangeType(log.TradeType)

		list = append(list, item)
	}

	g.Log().Infof(ctx, "查询账户明细成功: 共%d条", len(list))

	return &v1.GetBalanceLogRes{
		TotalRecharge: totalRecharge,
		TotalWithdraw: totalWithdraw,
		TotalBonus:    totalBonus,
		TotalRebate:   totalRebate,
		List:          list,
		Count:         int32(count),
	}, nil
}

// GetUserFlowRequirements 获取用户流水要求列表
func (s *sBalance) GetUserFlowRequirements(ctx context.Context, req *v1.GetUserFlowRequirementsReq) (*v1.GetUserFlowRequirementsRes, error) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		return &v1.GetUserFlowRequirementsRes{}, nil
	}
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetUserFlowRequirementsRes{}, nil
	}

	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	model := dao.BalanceFlowRequirement.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId)
	if req.Status >= 0 {
		model = model.Where("status", req.Status)
	}

	count, err := model.Count()
	if err != nil {
		return nil, fmt.Errorf("查询流水要求总数失败: %v", err)
	}

	var items []*entity.BalanceFlowRequirement
	err = model.Page(int(page), int(size)).
		Order("created_at DESC, id DESC").
		Scan(&items)
	if err != nil {
		return nil, fmt.Errorf("查询流水要求列表失败: %v", err)
	}

	list := make([]*v1.UserFlowRequirementInfo, 0, len(items))
	for _, item := range items {
		startTime := ""
		if item.StartTime != nil {
			startTime = item.StartTime.Format("2006-01-02 15:04:05")
		}
		expireTime := ""
		if item.ExpireTime != nil {
			expireTime = item.ExpireTime.Format("2006-01-02 15:04:05")
		}
		completeTime := ""
		if item.CompleteTime != nil {
			completeTime = item.CompleteTime.Format("2006-01-02 15:04:05")
		}
		createdAt := ""
		if item.CreatedAt != nil {
			createdAt = item.CreatedAt.Format("2006-01-02 15:04:05")
		}

		list = append(list, &v1.UserFlowRequirementInfo{
			Id:            int64(item.Id),
			SourceType:    int32(item.SourceType),
			TradeNo:       item.TradeNo,
			Amount:        item.Amount,
			BonusAmount:   item.BonusAmount,
			FlowMultiple:  item.FlowMultiple,
			RequiredFlow:  item.RequiredFlow,
			CompletedFlow: item.CompletedFlow,
			RemainingFlow: item.RemainingFlow,
			Status:        int32(item.Status),
			Progress:      item.Progress,
			StartTime:     startTime,
			ExpireTime:    expireTime,
			CompleteTime:  completeTime,
			Remark:        item.Remark,
			CreatedAt:     createdAt,
		})
	}

	return &v1.GetUserFlowRequirementsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// getTotalRecharge 获取充值总计
func (s *sBalance) getTotalRecharge(ctx context.Context, siteId, userId int, startTime, endTime string) float64 {
	result, err := dao.BalanceLog.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		WhereIn("trade_type", g.Slice{
			balanceConsts.LOG_TYPE_PAYMENT,
			balanceConsts.LOG_TYPE_TRANSFER,
			balanceConsts.LOG_TYPE_MANUAL_ADD,
		}).
		Where("status", balanceConsts.StatusSuccess).
		WhereBetween("updated_at", startTime, endTime).
		Sum("money")

	if err != nil {
		g.Log().Warningf(ctx, "查询充值总计失败: %v", err)
		return 0
	}

	return gconv.Float64(result)
}

// getTotalWithdraw 获取提现总计
func (s *sBalance) getTotalWithdraw(ctx context.Context, siteId, userId int, startTime, endTime string) float64 {
	result, err := dao.BalanceLog.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		WhereIn("trade_type", g.Slice{
			balanceConsts.LOG_TYPE_WITHDRAW,
			balanceConsts.LOG_TYPE_MANUAL_CUT,
		}).
		Where("status", balanceConsts.StatusSuccess).
		WhereBetween("updated_at", startTime, endTime).
		Sum("money")

	if err != nil {
		g.Log().Warningf(ctx, "查询提现总计失败: %v", err)
		return 0
	}

	return gconv.Float64(result)
}

// getTotalBonus 获取红利总计
func (s *sBalance) getTotalBonus(ctx context.Context, siteId, userId int, startTime, endTime string) float64 {
	result, err := dao.BalanceLog.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("trade_type", balanceConsts.LOG_TYPE_BONUS).
		Where("status", balanceConsts.StatusSuccess).
		WhereBetween("updated_at", startTime, endTime).
		Sum("money")

	if err != nil {
		g.Log().Warningf(ctx, "查询红利总计失败: %v", err)
		return 0
	}

	return gconv.Float64(result)
}

// getTotalRebate 获取返水总计
func (s *sBalance) getTotalRebate(ctx context.Context, siteId, userId int, startTime, endTime string) float64 {
	result, err := dao.BalanceLog.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("trade_type", balanceConsts.LOG_TYPE_REBATE).
		Where("status", balanceConsts.StatusSuccess).
		WhereBetween("updated_at", startTime, endTime).
		Sum("money")

	if err != nil {
		g.Log().Warningf(ctx, "查询返水总计失败: %v", err)
		return 0
	}

	return gconv.Float64(result)
}

// getFrontTypeByType 根据交易类型获取前端类型
func (s *sBalance) getFrontTypeByType(tradeType int) int32 {
	switch tradeType {
	case balanceConsts.LOG_TYPE_PAYMENT, balanceConsts.LOG_TYPE_TRANSFER, balanceConsts.LOG_TYPE_MANUAL_ADD:
		return 1 // 充值
	case balanceConsts.LOG_TYPE_WITHDRAW, balanceConsts.LOG_TYPE_MANUAL_CUT:
		return 2 // 提款
	case balanceConsts.LOG_TYPE_BONUS:
		return 3 // 红利
	case balanceConsts.LOG_TYPE_REBATE:
		return 4 // 返水
	case balanceConsts.LOG_TYPE_TRANSFER_IN, balanceConsts.LOG_TYPE_TRANSFER_OUT:
		return 5 // 转账
	default:
		return 0
	}
}

// getSymbol 获取符号
func (s *sBalance) getSymbol(tradeType int) string {
	switch tradeType {
	case balanceConsts.LOG_TYPE_PAYMENT, balanceConsts.LOG_TYPE_TRANSFER, balanceConsts.LOG_TYPE_MANUAL_ADD,
		balanceConsts.LOG_TYPE_BONUS, balanceConsts.LOG_TYPE_REBATE, balanceConsts.LOG_TYPE_TRANSFER_IN:
		return "+"
	case balanceConsts.LOG_TYPE_WITHDRAW, balanceConsts.LOG_TYPE_MANUAL_CUT, balanceConsts.LOG_TYPE_TRANSFER_OUT:
		return "-"
	default:
		return ""
	}
}

// getChangeType 获取交易类型（转入/转出）
func (s *sBalance) getChangeType(tradeType int) int32 {
	switch tradeType {
	case balanceConsts.LOG_TYPE_PAYMENT, balanceConsts.LOG_TYPE_TRANSFER, balanceConsts.LOG_TYPE_MANUAL_ADD,
		balanceConsts.LOG_TYPE_BONUS, balanceConsts.LOG_TYPE_REBATE, balanceConsts.LOG_TYPE_TRANSFER_IN:
		return 1 // 转入
	case balanceConsts.LOG_TYPE_WITHDRAW, balanceConsts.LOG_TYPE_MANUAL_CUT, balanceConsts.LOG_TYPE_TRANSFER_OUT:
		return 2 // 转出
	default:
		return 0
	}
}

// recordCallbackFailure 记录回调失败日志（同步）
func (s *sBalance) recordCallbackFailure(ctx context.Context, siteId int, callbackType int, stage string, reason string, req *v1.DepositCallbackReq, err error) {
	// 准备请求数据JSON
	requestData := gconv.String(g.Map{
		"trade_status":  req.TradeStatus,
		"order_sn":      req.OrderSn,
		"actual_amount": req.ActualAmount,
		"sign":          req.Sign,
	})

	// 准备错误信息
	errorMessage := reason
	if err != nil {
		errorMessage = fmt.Sprintf("%s: %v", reason, err)
	}

	// 插入失败日志
	_, insertErr := dao.BalanceCallbackLog.Ctx(ctx).Insert(g.Map{
		"site_id":        siteId,
		"callback_type":  callbackType,
		"order_sn":       req.OrderSn,
		"trade_status":   req.TradeStatus,
		"actual_amount":  req.ActualAmount,
		"failure_reason": reason,
		"failure_stage":  stage,
		"request_data":   requestData,
		"error_message":  errorMessage,
		"retry_count":    0,
		"is_resolved":    0,
		"created_at":     gtime.Now(),
		"updated_at":     gtime.Now(),
	})

	if insertErr != nil {
		g.Log().Errorf(ctx, "记录回调失败日志失败: %v", insertErr)
	} else {
		g.Log().Infof(ctx, "回调失败日志已记录: order_sn=%s, stage=%s, reason=%s", req.OrderSn, stage, reason)
	}
}

// DepositCallback 充值异步回调
func (s *sBalance) DepositCallback(ctx context.Context, req *v1.DepositCallbackReq) (*v1.DepositCallbackRes, error) {
	startTime := time.Now()

	g.Log().Infof(ctx, "充值异步回调开始: order_sn=%s, trade_status=%s, actual_amount=%.2f",
		req.OrderSn, req.TradeStatus, req.ActualAmount)

	// 1. 检查交易状态
	if req.TradeStatus != "success" {
		g.Log().Warningf(ctx, "充值失败，通知状态：%s, order_sn=%s", req.TradeStatus, req.OrderSn)
		// 记录失败日志
		s.recordCallbackFailure(ctx, 0, 1, "check_status", "支付状态非成功: "+req.TradeStatus, req, nil)
		return &v1.DepositCallbackRes{
			Success: false,
			Message: "支付失败，通知状态：" + req.TradeStatus,
		}, nil
	}

	// 2. 查找订单
	var paymentLog *entity.BalanceLog
	err := dao.BalanceLog.Ctx(ctx).Where("trade_no", req.OrderSn).Scan(&paymentLog)
	if err != nil {
		g.Log().Errorf(ctx, "查询充值订单失败: order_sn=%s, err=%v", req.OrderSn, err)
		s.recordCallbackFailure(ctx, 0, 1, "find_order", "查询订单失败", req, err)
		return &v1.DepositCallbackRes{
			Success: false,
			Message: "查询订单失败",
		}, nil
	}

	if paymentLog == nil {
		g.Log().Warningf(ctx, "充值订单不存在: order_sn=%s", req.OrderSn)
		s.recordCallbackFailure(ctx, 0, 1, "find_order", "订单不存在", req, nil)
		return &v1.DepositCallbackRes{
			Success: false,
			Message: "订单不存在",
		}, nil
	}

	duration1 := time.Since(startTime)
	g.Log().Infof(ctx, "第1阶段-查找订单，总耗时: %.2fms", float64(duration1.Milliseconds()))

	// 3. 检查订单状态
	if paymentLog.Status == balanceConsts.StatusSuccess {
		g.Log().Warningf(ctx, "充值订单已完成: order_sn=%s, 当前状态为:存款成功", req.OrderSn)
		return &v1.DepositCallbackRes{
			Success: true,
			Message: "订单已完成",
		}, nil
	}

	if paymentLog.Status == balanceConsts.StatusFail {
		g.Log().Warningf(ctx, "充值订单已失败: order_sn=%s", req.OrderSn)
		return &v1.DepositCallbackRes{
			Success: false,
			Message: "订单已失败",
		}, nil
	}

	time2 := time.Now()

	// 4. 实际支付金额
	actualAmount := req.ActualAmount
	if actualAmount <= 0 {
		actualAmount = 0
	}

	duration2 := time.Since(time2)
	g.Log().Infof(ctx, "第2阶段-准备金额，总耗时: %.2fms", float64(duration2.Milliseconds()))

	time3 := time.Now()

	// 6. 开始事务处理 - 使用 Ctx(ctx) 保持动态数据库路由
	err = dao.BalanceLog.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 6.0 更新在线支付充值记录状态
		_, err := dao.BalanceLogPayment.Ctx(ctx).TX(tx).
			Where("trade_no", req.OrderSn).
			Data(g.Map{
				"status":       balanceConsts.StatusSuccess, // STATUS_SUCCESS
				"actual_money": actualAmount,
			}).Update()

		if err != nil {
			g.Log().Errorf(ctx, "更新在线支付充值记录失败: order_sn=%s, err=%v", req.OrderSn, err)
			return err
		}

		// 6.1 更新充值记录状态（balance_log表）
		_, err = dao.BalanceLog.Ctx(ctx).TX(tx).
			Where("trade_no", req.OrderSn).
			Data(g.Map{
				"status":       balanceConsts.StatusSuccess, // STATUS_SUCCESS
				"money":        actualAmount,                // 更新金额
				"actual_money": actualAmount,                // 更新实际金额
				"remark":       "异步确认在线支付",
				"updated_at":   gtime.Now(),
			}).Update()

		if err != nil {
			g.Log().Errorf(ctx, "更新充值记录失败: order_sn=%s, err=%v", req.OrderSn, err)
			return err
		}

		// 6.1.2 同时更新balance_log_payment表
		_, err = dao.BalanceLogPayment.Ctx(ctx).TX(tx).
			Where("trade_no", req.OrderSn).
			Data(g.Map{
				"status":       balanceConsts.StatusSuccess, // STATUS_SUCCESS
				"money":        actualAmount,                // 更新金额
				"actual_money": actualAmount,                // 更新实际金额
				"remark":       "异步确认在线支付",
				"updated_at":   gtime.Now(),
			}).Update()

		if err != nil {
			g.Log().Errorf(ctx, "更新充值支付记录失败: order_sn=%s, err=%v", req.OrderSn, err)
			return err
		}

		duration31 := time.Since(time3)
		g.Log().Infof(ctx, "第3.1阶段-更新充值记录，总耗时: %.2fms", float64(duration31.Milliseconds()))

		time32 := time.Now()

		// 6.2 获取用户余额（加锁）
		var balanceUser *entity.BalanceUser
		err = dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("site_id", paymentLog.SiteId).
			Where("user_id", paymentLog.UserId).
			LockUpdate().
			Scan(&balanceUser)

		if err != nil || balanceUser == nil {
			g.Log().Errorf(ctx, "获取用户余额失败: user_id=%d, err=%v", paymentLog.UserId, err)
			return gerror.New("获取用户余额失败")
		}

		duration32 := time.Since(time32)
		g.Log().Infof(ctx, "第3.2阶段-获取用户余额，总耗时: %.2fms", float64(duration32.Milliseconds()))

		time33 := time.Now()

		// 6.3 计算新余额
		oldBalance := balanceUser.Balance
		newBalance := oldBalance + actualAmount

		// 更新用户余额
		_, err = dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("site_id", paymentLog.SiteId).
			Where("user_id", paymentLog.UserId).
			Data(g.Map{
				"balance":    newBalance,
				"updated_at": gtime.Now(),
			}).Update()

		if err != nil {
			g.Log().Errorf(ctx, "更新用户余额失败: user_id=%d, err=%v", paymentLog.UserId, err)
			return err
		}

		duration33 := time.Since(time33)
		g.Log().Infof(ctx, "第3.3阶段-更新用户余额，总耗时: %.2fms", float64(duration33.Milliseconds()))

		time34 := time.Now()

		// 6.4 写入账变记录
		_, err = dao.BalanceChange.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        paymentLog.SiteId,
			"trade_type":     balanceConsts.TradeTypeRecharge, // 2=充值
			"user_id":        paymentLog.UserId,
			"username":       balanceUser.Username,
			"trade_no":       req.OrderSn,
			"balance_old":    oldBalance,
			"money":          actualAmount,
			"balance_new":    newBalance,
			"balance_frozen": balanceUser.BalanceFrozen,
			"status":         balanceConsts.StatusSuccess, // 3=成功
			"remark":         "异步确认在线支付",
			"created_at":     gtime.Now(),
			"updated_at":     gtime.Now(),
		})

		if err != nil {
			g.Log().Errorf(ctx, "写入账变记录失败: order_sn=%s, err=%v", req.OrderSn, err)
			return err
		}

		duration34 := time.Since(time34)
		g.Log().Infof(ctx, "第3.4阶段-写入账变记录，总耗时: %.2fms", float64(duration34.Milliseconds()))

		// 6.5 记录流水要求（充值成功）
		if _, _, err := flow.RecordFlowRequirement(ctx, &flow.RecordFlowRequirementRequest{
			SiteId:      int(paymentLog.SiteId),
			UserId:      int64(paymentLog.UserId),
			Username:    balanceUser.Username,
			SourceType:  balanceConsts.SourceTypeRecharge,
			SourceId:    gconv.Int64(req.OrderSn),
			TradeNo:     req.OrderSn,
			Amount:      actualAmount,
			BonusAmount: 0,
			Remark:      "充值流水要求",
		}); err != nil {
			g.Log().Errorf(ctx, "记录流水要求失败: order_sn=%s, err=%v", req.OrderSn, err)
			return err
		}

		return nil
	})

	if err != nil {
		g.Log().Errorf(ctx, "充值回调事务处理失败: order_sn=%s, err=%v", req.OrderSn, err)
		s.recordCallbackFailure(ctx, paymentLog.SiteId, 1, "transaction", "数据库事务失败", req, err)
		return &v1.DepositCallbackRes{
			Success: false,
			Message: "支付资金处理失败: " + err.Error(),
		}, nil
	}

	duration3 := time.Since(time3)
	g.Log().Infof(ctx, "第3阶段-事务处理，总耗时: %.2fms", float64(duration3.Milliseconds()))

	// 4. 更新 app_service 用户充值次数，并在首充时发放邀请红利
	var incRes *userv1.IncreasePayTimesRes
	userClient, userConn, err := grpc_client.GetUserClient(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取 User 客户端失败: %v", err)
	} else {
		defer userConn.Close()
		ctxWithMeta := metadata.AppendToOutgoingContext(ctx, "site_id", gconv.String(paymentLog.SiteId))
		incRes, err = userClient.IncreasePayTimes(ctxWithMeta, &userv1.IncreasePayTimesReq{
			SiteId: int32(paymentLog.SiteId),
			UserId: int32(paymentLog.UserId),
		})
		if err != nil {
			g.Log().Warningf(ctx, "调用 IncreasePayTimes 失败: %v", err)
		} else if incRes == nil || !incRes.Success {
			msg := ""
			if incRes != nil {
				msg = incRes.Message
			}
			g.Log().Warningf(ctx, "IncreasePayTimes 返回失败: %s", msg)
		} else if incRes.InviterId > 0 {
			model := dao.GetDynamicModel(ctx, "balance_invite_reward", g.DB().Model("balance_invite_reward").Safe().Ctx(ctx))
			count, err := model.Where("site_id", paymentLog.SiteId).Where("invitee_id", paymentLog.UserId).Count()
			if err != nil {
				g.Log().Warningf(ctx, "查询邀请红利发放记录失败: %v", err)
			} else if count > 0 {
				g.Log().Warningf(ctx, "邀请红利已发放，跳过: invitee_id=%d", paymentLog.UserId)
			} else {
				g.Log().Infof(ctx, "邀请红利未发放，开始读取配置: invitee_id=%d, inviter_id=%d",
					paymentLog.UserId, incRes.InviterId)
				inviteCfg, err := getInviteActivityParams(ctx, paymentLog.SiteId)
				if err != nil {
					g.Log().Warningf(ctx, "读取邀请活动配置失败: %v", err)
				} else if inviteCfg == nil {
					g.Log().Warningf(ctx, "邀请活动未配置或已关闭，跳过发放: user_id=%d", incRes.InviterId)
				} else if inviteCfg.TriggerEvent != 1 || inviteCfg.RewardType != 1 {
					g.Log().Warningf(ctx, "邀请活动配置不匹配(仅支持充值总和+固定奖励)，跳过发放: user_id=%d", incRes.InviterId)
				} else {
					g.Log().Infof(ctx, "邀请活动配置已获取: trigger_event=%d, reward_type=%d, min=%.2f, bonus=%.2f",
						inviteCfg.TriggerEvent, inviteCfg.RewardType, inviteCfg.MinRechargeAmount, inviteCfg.InviterRewardAmount)
					minDeposit := inviteCfg.MinRechargeAmount
					bonusAmount := inviteCfg.InviterRewardAmount
					if bonusAmount <= 0 {
						g.Log().Warningf(ctx, "邀请红利未配置或金额为0，跳过发放: user_id=%d", incRes.InviterId)
					} else {
						totalActual, err := dao.BalanceLogPayment.Ctx(ctx).
							Where("site_id", paymentLog.SiteId).
							Where("user_id", paymentLog.UserId).
							Where("status", balanceConsts.StatusSuccess).
							Sum("actual_money")
						if err != nil {
							g.Log().Warningf(ctx, "统计邀请用户累计实付失败: %v", err)
						} else if totalActual < minDeposit {
							g.Log().Warningf(ctx, "累计实付不足%.2f，跳过发放邀请红利: invitee_id=%d, total=%.2f",
								minDeposit, paymentLog.UserId, totalActual)
						} else {
							g.Log().Infof(ctx, "满足邀请红利条件，准备发放: inviter_id=%d, invitee_id=%d, amount=%.2f",
								incRes.InviterId, paymentLog.UserId, bonusAmount)
							grantRes, err := s.GrantInviteBonus(ctx, &v1.GrantInviteBonusReq{
								SiteId: int32(paymentLog.SiteId),
								UserId: int32(incRes.InviterId),
								Amount: bonusAmount,
								Remark: "邀请充值红利",
							})
							if err != nil {
								g.Log().Warningf(ctx, "发放邀请红利失败: %v", err)
							} else if grantRes != nil && !grantRes.Success {
								g.Log().Warningf(ctx, "发放邀请红利失败: %s", grantRes.Message)
							} else {
								_, err = model.Data(g.Map{
									"site_id":      paymentLog.SiteId,
									"inviter_id":   incRes.InviterId,
									"invitee_id":   paymentLog.UserId,
									"bonus_amount": bonusAmount,
									"trade_no":     grantRes.TradeNo,
									"created_at":   gtime.Now(),
								}).Insert()
								if err != nil {
									g.Log().Warningf(ctx, "记录邀请红利发放失败: %v", err)
								}
							}
						}
					}
				}
			}
		}

		// 充值送活动发放（不影响主流程）
		s.tryGrantDepositBonus(ctx, paymentLog.SiteId, paymentLog.UserId, actualAmount, incRes)
	}

	totalDuration := time.Since(startTime)
	g.Log().Infof(ctx, "充值异步回调成功: order_sn=%s, amount=%.2f, 总耗时: %.2fms",
		req.OrderSn, actualAmount, float64(totalDuration.Milliseconds()))

	return &v1.DepositCallbackRes{
		Success: true,
		Message: "支付成功",
	}, nil
}

// Withdraw 申请提现
func (s *sBalance) Withdraw(ctx context.Context, req *v1.WithdrawReq) (*v1.WithdrawRes, error) {
	// 1. 获取用户信息
	userId, exists := middleware.GetUserIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取用户ID")
		return &v1.WithdrawRes{
			Success: false,
			Message: "未登录",
		}, nil
	}

	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		g.Log().Error(ctx, "未获取站点ID")
		return &v1.WithdrawRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}

	username := middleware.GetUsernameFromContext(ctx)

	//==========防重放验证====================
	// 2. 验证 nonce（防重放）
	nonceService := &common.NonceService{}
	err := nonceService.ValidateAndConsume(ctx, "withdraw", int64(siteId), int64(userId), req.Nonce)
	if err != nil {
		return &v1.WithdrawRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	//==========防重放验证====================

	// 3. 调用 Domain 层处理提现业务逻辑
	withdrawDomain := &withdraw.WithdrawDomain{}
	result, err := withdrawDomain.ApplyWithdraw(ctx, &withdraw.ApplyWithdrawRequest{
		SiteId:     int64(siteId),
		UserId:     int64(userId),
		Username:   username,
		Money:      req.Money,
		WithdrawId: int64(req.WithdrawId),
		BankCardId: 0, // 暂时不使用，由 Domain 层查询默认银行卡
	})

	if err != nil {
		g.Log().Errorf(ctx, "提现申请失败: %v", err)
		return &v1.WithdrawRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.WithdrawRes{
		Success: true,
		Message: result.Message,
		TradeNo: result.TradeNo,
	}, nil
}

// GrantInviteBonus 发放邀请红利
func (s *sBalance) GrantInviteBonus(ctx context.Context, req *v1.GrantInviteBonusReq) (*v1.GrantInviteBonusRes, error) {
	if req == nil {
		return &v1.GrantInviteBonusRes{Success: false, Message: "请求参数不能为空"}, nil
	}
	if req.UserId <= 0 {
		return &v1.GrantInviteBonusRes{Success: false, Message: "用户ID不能为空"}, nil
	}
	if req.Amount <= 0 {
		return &v1.GrantInviteBonusRes{Success: false, Message: "红利金额必须大于0"}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		if req.SiteId <= 0 {
			return &v1.GrantInviteBonusRes{Success: false, Message: "未获取站点ID"}, nil
		}
		siteId = uint(req.SiteId)
		// 兜底设置动态数据库组（避免无metadata时路由失败）
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	// 读取流水规则（红利类型）
	rule, err := dao.BalanceFlowRule.GetActiveRule(ctx, int(siteId), balanceConsts.RuleTypeBonus)
	if err != nil {
		return &v1.GrantInviteBonusRes{Success: false, Message: "查询流水规则失败"}, nil
	}
	if rule == nil {
		return &v1.GrantInviteBonusRes{Success: false, Message: "未配置可用的流水规则"}, nil
	}

	remark := req.Remark
	if remark == "" {
		remark = "邀请红利"
	}

	var tradeNo string

	err = dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询并锁定用户余额
		var balanceUser *entity.BalanceUser
		if err := dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("site_id", siteId).
			Where("user_id", req.UserId).
			LockUpdate().
			Scan(&balanceUser); err != nil {
			return err
		}
		if balanceUser == nil || balanceUser.Id == 0 {
			return fmt.Errorf("用户余额记录不存在")
		}

		balanceOld := balanceUser.Balance
		balanceNew := balanceOld + req.Amount
		now := gtime.Now()

		// 更新用户余额
		if _, err := dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("id", balanceUser.Id).
			Update(g.Map{
				"balance":    balanceNew,
				"updated_at": now,
			}); err != nil {
			return err
		}

		// 生成流水号
		tradeNo = util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_BONUS)

		// 记录资金日志
		if _, err := dao.BalanceLog.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":      siteId,
			"user_id":      req.UserId,
			"trade_type":   balanceConsts.LOG_TYPE_BONUS,
			"trade_no":     tradeNo,
			"money":        req.Amount,
			"actual_money": req.Amount,
			"fee":          0,
			"status":       balanceConsts.StatusSuccess,
			"remark":       remark,
			"created_at":   now,
			"updated_at":   now,
		}); err != nil {
			return err
		}

		// 记录账变
		if _, err := dao.BalanceChange.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"trade_type":     balanceConsts.TradeTypeBonus,
			"user_id":        req.UserId,
			"username":       balanceUser.Username,
			"trade_no":       tradeNo,
			"balance_old":    balanceOld,
			"money":          req.Amount,
			"balance_new":    balanceNew,
			"balance_frozen": balanceUser.BalanceFrozen,
			"status":         balanceConsts.StatusSuccess,
			"remark":         remark,
			"created_at":     now,
			"updated_at":     now,
		}); err != nil {
			return err
		}

		// 记录红利日志
		bonusResult, err := dao.BalanceLogBonus.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":              siteId,
			"trade_type":           balanceConsts.BonusTradeTypeInvite,
			"user_id":              req.UserId,
			"username":             balanceUser.Username,
			"trade_no":             tradeNo,
			"activity_recharge_id": 0,
			"money":                req.Amount,
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

		// 记录流水要求
		flowMultiple := rule.FlowMultiple
		expireDays := rule.ExpireDays
		requiredFlow := req.Amount * flowMultiple
		var expireTime *gtime.Time
		if expireDays > 0 {
			expireTime = now.AddDate(0, 0, expireDays)
		}

		if _, err := dao.BalanceFlowRequirement.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"user_id":        req.UserId,
			"username":       balanceUser.Username,
			"source_type":    balanceConsts.SourceTypeBonus,
			"source_id":      bonusId,
			"trade_no":       tradeNo,
			"amount":         0,
			"bonus_amount":   req.Amount,
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

		return nil
	})

	if err != nil {
		g.Log().Errorf(ctx, "发放邀请红利失败: %v", err)
		return &v1.GrantInviteBonusRes{Success: false, Message: err.Error()}, nil
	}

	return &v1.GrantInviteBonusRes{
		Success: true,
		Message: "发放红利成功",
		TradeNo: tradeNo,
	}, nil
}

func (s *sBalance) tryGrantRegisterBonus(ctx context.Context, siteId int32, userId int32) {
	if siteId <= 0 || userId <= 0 {
		return
	}

	model := dao.GetDynamicModel(ctx, "balance_register_reward", g.DB().Model("balance_register_reward").Safe().Ctx(ctx))
	count, err := model.Where("site_id", siteId).Where("user_id", userId).Count()
	if err != nil {
		g.Log().Warningf(ctx, "查询注册送彩金记录失败: %v", err)
		return
	}
	if count > 0 {
		g.Log().Infof(ctx, "注册送彩金已发放，跳过: user_id=%d", userId)
		return
	}

	cfg, err := getRegisterActivityParams(ctx, int(siteId))
	if err != nil {
		g.Log().Warningf(ctx, "读取注册活动配置失败: %v", err)
		return
	}
	if cfg == nil {
		g.Log().Infof(ctx, "注册活动未配置或已关闭，跳过发放: user_id=%d", userId)
		return
	}
	if cfg.RewardAmount <= 0 {
		g.Log().Warningf(ctx, "注册活动奖励金额为0，跳过发放: user_id=%d", userId)
		return
	}

	var registerIP string
	if cfg.LimitIp > 0 {
		var err error
		registerIP, _, err = getRegisterUserInfo(ctx, siteId, userId)
		if err != nil {
			g.Log().Warningf(ctx, "读取注册用户信息失败: %v", err)
			return
		}
		if registerIP == "" {
			g.Log().Warningf(ctx, "注册活动限制IP但未获取到注册IP，跳过发放: user_id=%d", userId)
			return
		}
		ipCount, err := model.Where("site_id", siteId).Where("register_ip", registerIP).Count()
		if err != nil {
			g.Log().Warningf(ctx, "查询注册IP发放次数失败: %v", err)
			return
		}
		if ipCount >= cfg.LimitIp {
			g.Log().Warningf(ctx, "注册IP已达限制次数，跳过发放: ip=%s, limit=%d", registerIP, cfg.LimitIp)
			return
		}
	}

	if cfg.LimitDevice > 0 {
		g.Log().Infof(ctx, "注册活动设备限制已配置(limit_device=%d)，当前仅启用IP限制", cfg.LimitDevice)
	}

	tradeNo, err := s.grantRegisterBonus(ctx, int(siteId), userId, cfg.RewardAmount)
	if err != nil {
		g.Log().Warningf(ctx, "发放注册送彩金失败: %v", err)
		return
	}

	if _, err := model.Data(g.Map{
		"site_id":      siteId,
		"user_id":      userId,
		"register_ip":  registerIP,
		"bonus_amount": cfg.RewardAmount,
		"trade_no":     tradeNo,
		"created_at":   gtime.Now(),
	}).Insert(); err != nil {
		g.Log().Warningf(ctx, "记录注册送彩金失败: %v", err)
	}
}

func (s *sBalance) grantRegisterBonus(ctx context.Context, siteId int, userId int32, amount float64) (string, error) {
	if siteId <= 0 {
		return "", fmt.Errorf("站点ID不能为空")
	}
	if userId <= 0 {
		return "", fmt.Errorf("用户ID不能为空")
	}
	if amount <= 0 {
		return "", fmt.Errorf("红利金额必须大于0")
	}

	// 兜底设置动态数据库组（避免无metadata时路由失败）
	if ctx.Value("dynamic_db_group") == nil {
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	// 优先使用注册活动流水规则，未配置则回退红利规则
	rule, err := dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeRegisterActivity)
	if err != nil {
		return "", fmt.Errorf("查询流水规则失败")
	}
	if rule == nil {
		rule, err = dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeBonus)
		if err != nil {
			return "", fmt.Errorf("查询流水规则失败")
		}
	}
	if rule == nil {
		return "", fmt.Errorf("未配置可用的流水规则")
	}

	var tradeNo string
	now := gtime.Now()

	err = dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询并锁定用户余额
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

		balanceOld := balanceUser.Balance
		balanceNew := balanceOld + amount

		// 更新用户余额
		if _, err := dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("id", balanceUser.Id).
			Update(g.Map{
				"balance":    balanceNew,
				"updated_at": now,
			}); err != nil {
			return err
		}

		// 生成流水号
		tradeNo = util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_BONUS)

		remark := "注册送彩金"

		// 记录资金日志
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

		// 记录账变
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

		// 记录红利日志
		bonusResult, err := dao.BalanceLogBonus.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":              siteId,
			"trade_type":           balanceConsts.BonusTradeTypeRegister,
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

		// 记录流水要求
		flowMultiple := rule.FlowMultiple
		expireDays := rule.ExpireDays
		requiredFlow := amount * flowMultiple
		var expireTime *gtime.Time
		if expireDays > 0 {
			expireTime = now.AddDate(0, 0, expireDays)
		}

		if _, err := dao.BalanceFlowRequirement.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"user_id":        userId,
			"username":       balanceUser.Username,
			"source_type":    balanceConsts.SourceTypeRegisterActivity,
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

		return nil
	})

	if err != nil {
		return "", err
	}

	return tradeNo, nil
}

func (s *sBalance) tryGrantDepositBonus(ctx context.Context, siteId int, userId int, actualAmount float64, incRes *userv1.IncreasePayTimesRes) {
	if siteId <= 0 || userId <= 0 {
		return
	}

	// 兜底设置动态数据库组（避免无metadata时路由失败）
	if ctx.Value("dynamic_db_group") == nil {
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	model := dao.GetDynamicModel(ctx, "balance_deposit_reward", g.DB().Model("balance_deposit_reward").Safe().Ctx(ctx))
	count, err := model.Where("site_id", siteId).Where("user_id", userId).Count()
	if err != nil {
		g.Log().Warningf(ctx, "查询充值送活动发放记录失败: %v", err)
		return
	}
	if count > 0 {
		g.Log().Infof(ctx, "充值送活动已发放，跳过: user_id=%d", userId)
		return
	}

	cfg, err := getDepositActivityParams(ctx, siteId)
	if err != nil {
		g.Log().Warningf(ctx, "读取充值送活动配置失败: %v", err)
		return
	}
	if cfg == nil {
		g.Log().Infof(ctx, "充值送活动未配置或已关闭，跳过发放: user_id=%d", userId)
		return
	}
	if cfg.RewardType != 1 && cfg.RewardType != 2 {
		g.Log().Warningf(ctx, "充值送活动奖励类型不支持: %d", cfg.RewardType)
		return
	}
	if cfg.RewardAmount <= 0 {
		g.Log().Warningf(ctx, "充值送活动奖励金额为0，跳过发放: user_id=%d", userId)
		return
	}

	minAmount := cfg.MinRechargeAmount
	switch cfg.TriggerEvent {
	case 1:
		// 充值总和
		totalActual, err := dao.BalanceLogPayment.Ctx(ctx).
			Where("site_id", siteId).
			Where("user_id", userId).
			Where("status", balanceConsts.StatusSuccess).
			Sum("actual_money")
		if err != nil {
			g.Log().Warningf(ctx, "统计充值送活动累计实付失败: %v", err)
			return
		}
		if minAmount > 0 && totalActual < minAmount {
			g.Log().Warningf(ctx, "累计实付不足%.2f，跳过充值送活动: user_id=%d, total=%.2f",
				minAmount, userId, totalActual)
			return
		}
	case 2:
		// 首充
		if incRes == nil || !incRes.Success {
			g.Log().Warningf(ctx, "充值送活动需要首充信息但未获取到，跳过发放: user_id=%d", userId)
			return
		}
		if incRes.OldPayTimes != 0 || incRes.NewPayTimes != 1 {
			g.Log().Infof(ctx, "充值送活动非首充，跳过发放: user_id=%d, old=%d, new=%d",
				userId, incRes.OldPayTimes, incRes.NewPayTimes)
			return
		}
		if minAmount > 0 && actualAmount < minAmount {
			g.Log().Warningf(ctx, "首充金额不足%.2f，跳过充值送活动: user_id=%d, amount=%.2f",
				minAmount, userId, actualAmount)
			return
		}
	default:
		g.Log().Warningf(ctx, "充值送活动触发类型不支持: %d", cfg.TriggerEvent)
		return
	}

	bonusAmount := cfg.RewardAmount
	if cfg.RewardType == 2 {
		if actualAmount <= 0 {
			g.Log().Warningf(ctx, "充值金额为0，跳过比例奖励: user_id=%d", userId)
			return
		}
		bonusAmount = actualAmount * cfg.RewardAmount
	}
	bonusAmount = float64(int(bonusAmount*100+0.5)) / 100
	if bonusAmount <= 0 {
		g.Log().Warningf(ctx, "充值送活动奖励金额为0，跳过发放: user_id=%d", userId)
		return
	}

	tradeNo, err := s.grantDepositBonus(ctx, siteId, int32(userId), bonusAmount)
	if err != nil {
		g.Log().Warningf(ctx, "发放充值送活动失败: %v", err)
		return
	}

	if _, err := model.Data(g.Map{
		"site_id":      siteId,
		"user_id":      userId,
		"bonus_amount": bonusAmount,
		"trade_no":     tradeNo,
		"created_at":   gtime.Now(),
	}).Insert(); err != nil {
		g.Log().Warningf(ctx, "记录充值送活动发放失败: %v", err)
	}
}

func (s *sBalance) grantDepositBonus(ctx context.Context, siteId int, userId int32, amount float64) (string, error) {
	if siteId <= 0 {
		return "", fmt.Errorf("站点ID不能为空")
	}
	if userId <= 0 {
		return "", fmt.Errorf("用户ID不能为空")
	}
	if amount <= 0 {
		return "", fmt.Errorf("红利金额必须大于0")
	}

	// 兜底设置动态数据库组（避免无metadata时路由失败）
	if ctx.Value("dynamic_db_group") == nil {
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	// 优先使用充值送活动流水规则，未配置则回退红利规则
	rule, err := dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeRechargeActivity)
	if err != nil {
		return "", fmt.Errorf("查询流水规则失败")
	}
	if rule == nil {
		rule, err = dao.BalanceFlowRule.GetActiveRule(ctx, siteId, balanceConsts.RuleTypeBonus)
		if err != nil {
			return "", fmt.Errorf("查询流水规则失败")
		}
	}
	if rule == nil {
		return "", fmt.Errorf("未配置可用的流水规则")
	}

	var tradeNo string
	now := gtime.Now()

	err = dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询并锁定用户余额
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

		balanceOld := balanceUser.Balance
		balanceNew := balanceOld + amount

		// 更新用户余额
		if _, err := dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("id", balanceUser.Id).
			Update(g.Map{
				"balance":    balanceNew,
				"updated_at": now,
			}); err != nil {
			return err
		}

		// 生成流水号
		tradeNo = util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_BONUS)

		remark := "充值送彩金"

		// 记录资金日志
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

		// 记录账变
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

		// 记录红利日志
		bonusResult, err := dao.BalanceLogBonus.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":              siteId,
			"trade_type":           balanceConsts.BonusTradeTypeDeposit,
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

		// 记录流水要求
		flowMultiple := rule.FlowMultiple
		expireDays := rule.ExpireDays
		requiredFlow := amount * flowMultiple
		var expireTime *gtime.Time
		if expireDays > 0 {
			expireTime = now.AddDate(0, 0, expireDays)
		}

		if _, err := dao.BalanceFlowRequirement.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"user_id":        userId,
			"username":       balanceUser.Username,
			"source_type":    balanceConsts.SourceTypeRechargeActivity,
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

		return nil
	})

	if err != nil {
		return "", err
	}

	return tradeNo, nil
}

func getRegisterUserInfo(ctx context.Context, siteId int32, userId int32) (string, string, error) {
	userClient, userConn, err := grpc_client.GetUserClient(ctx)
	if err != nil {
		return "", "", err
	}
	defer userConn.Close()

	ctxWithMeta := metadata.AppendToOutgoingContext(ctx, "site_id", gconv.String(siteId))
	res, err := userClient.GetUserBasicInfo(ctxWithMeta, &userv1.GetUserBasicInfoReq{
		Id: userId,
	})
	if err != nil {
		return "", "", err
	}
	if res == nil || res.User == nil {
		return "", "", fmt.Errorf("用户信息不存在")
	}

	registerIP := strings.TrimSpace(res.User.RegisterIp)
	deviceID := strings.TrimSpace(res.User.Device)
	return registerIP, deviceID, nil
}

// GetValidReferralCount 统计有效邀请人数（按充值总额）
func (s *sBalance) GetValidReferralCount(ctx context.Context, req *v1.GetValidReferralCountReq) (*v1.GetValidReferralCountRes, error) {
	if req == nil || len(req.UserIds) == 0 {
		return &v1.GetValidReferralCountRes{Count: 0}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		if req.SiteId <= 0 {
			return &v1.GetValidReferralCountRes{Count: 0}, nil
		}
		siteId = uint(req.SiteId)
		// 兜底设置动态数据库组（避免无metadata时路由失败）
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	minTotal := req.MinTotal
	if minTotal <= 0 {
		inviteCfg, err := getInviteActivityParams(ctx, int(siteId))
		if err != nil {
			g.Log().Warningf(ctx, "读取邀请活动配置失败: %v", err)
			return &v1.GetValidReferralCountRes{Count: 0}, nil
		}
		if inviteCfg == nil || inviteCfg.TriggerEvent != 1 || inviteCfg.RewardType != 1 {
			return &v1.GetValidReferralCountRes{Count: 0}, nil
		}
		minTotal = inviteCfg.MinRechargeAmount
	}

	userIds := make([]int, 0, len(req.UserIds))
	for _, id := range req.UserIds {
		if id > 0 {
			userIds = append(userIds, int(id))
		}
	}
	if len(userIds) == 0 {
		return &v1.GetValidReferralCountRes{Count: 0}, nil
	}

	var rows []struct {
		UserId int64   `json:"user_id"`
		Total  float64 `json:"total"`
	}

	err := dao.BalanceLogPayment.Ctx(ctx).
		Where("site_id", siteId).
		Where("status", balanceConsts.StatusSuccess).
		WhereIn("user_id", userIds).
		Fields("user_id, IFNULL(SUM(actual_money),0) as total").
		Group("user_id").
		Having("total >= ?", minTotal).
		Scan(&rows)
	if err != nil {
		g.Log().Warningf(ctx, "统计有效邀请人数失败: %v", err)
		return &v1.GetValidReferralCountRes{Count: 0}, nil
	}

	return &v1.GetValidReferralCountRes{Count: int32(len(rows))}, nil
}

// GetInviteBonusSummary 获取邀请红利统计
func (s *sBalance) GetInviteBonusSummary(ctx context.Context, req *v1.GetInviteBonusSummaryReq) (*v1.GetInviteBonusSummaryRes, error) {
	if req == nil || req.UserId <= 0 {
		return &v1.GetInviteBonusSummaryRes{}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		if req.SiteId <= 0 {
			return &v1.GetInviteBonusSummaryRes{}, nil
		}
		siteId = uint(req.SiteId)
		siteCode := registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
		if siteCode != "" {
			ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
		}
	}

	now := gtime.Now()
	todayStart := now.StartOfDay()
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	baseQuery := dao.BalanceLogBonus.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", req.UserId).
		Where("trade_type", balanceConsts.BonusTradeTypeInvite).
		Where("status", balanceConsts.StatusSuccess)

	todayAmount, err := baseQuery.WhereBetween("created_at", todayStart, tomorrowStart).Sum("money")
	if err != nil {
		todayAmount = 0
	}
	yesterdayAmount, err := baseQuery.WhereBetween("created_at", yesterdayStart, todayStart).Sum("money")
	if err != nil {
		yesterdayAmount = 0
	}
	totalAmount, err := baseQuery.Sum("money")
	if err != nil {
		totalAmount = 0
	}

	return &v1.GetInviteBonusSummaryRes{
		TodayAmount:     gconv.Float64(todayAmount),
		YesterdayAmount: gconv.Float64(yesterdayAmount),
		TotalAmount:     gconv.Float64(totalAmount),
	}, nil
}

type vipGradeProgressInfo struct {
	Id           int32
	Name         string
	PaymentLimit float64
	BetLimit     float64
}

// GetVipUpgradeProgress 获取当前用户VIP升级进度
func (s *sBalance) GetVipUpgradeProgress(ctx context.Context, req *v1.GetVipUpgradeProgressReq) (*v1.GetVipUpgradeProgressRes, error) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		return &v1.GetVipUpgradeProgressRes{
			Success: false,
			Message: "未登录",
		}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.GetVipUpgradeProgressRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}

	siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
	if siteCode == "" {
		siteCode = registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
	}
	if siteCode != "" && ctx.Value("dynamic_db_group") == nil {
		ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
	}

	depositAmount := 0.0
	betAmount := 0.0
	var reportTotal *entity.UserReportTotal
	if err := dao.UserReportTotal.Ctx(ctx).
		Where("site_id", siteId).
		Where("user_id", userId).
		Scan(&reportTotal); err != nil {
		g.Log().Errorf(ctx, "查询用户累计统计失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetVipUpgradeProgressRes{
			Success: false,
			Message: "查询用户累计统计失败",
		}, nil
	}
	if reportTotal != nil {
		depositAmount = reportTotal.DepositAmount
		betAmount = reportTotal.BetAmount
	}

	currentGrade, nextGrade, isMaxGrade, err := s.getUserCurrentAndNextVipGrade(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Errorf(ctx, "查询用户等级进度失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.GetVipUpgradeProgressRes{
			Success: false,
			Message: "查询用户等级进度失败",
		}, nil
	}

	if isMaxGrade || nextGrade == nil {
		return &v1.GetVipUpgradeProgressRes{
			Success:          true,
			Message:          "已达到最高等级",
			CurrentGradeId:   currentGrade.Id,
			CurrentGradeName: currentGrade.Name,
			DepositAmount:    depositAmount,
			BetAmount:        betAmount,
			DepositProgress:  100,
			BetProgress:      100,
			Progress:         100,
			IsMaxGrade:       true,
		}, nil
	}

	depositProgress := roundVipProgress(calcVipProgressPercent(depositAmount, nextGrade.PaymentLimit))
	betProgress := roundVipProgress(calcVipProgressPercent(betAmount, nextGrade.BetLimit))
	progress := roundVipProgress(math.Min(depositProgress, betProgress))

	return &v1.GetVipUpgradeProgressRes{
		Success:          true,
		Message:          "获取成功",
		CurrentGradeId:   currentGrade.Id,
		CurrentGradeName: currentGrade.Name,
		NextGradeId:      nextGrade.Id,
		NextGradeName:    nextGrade.Name,
		DepositAmount:    depositAmount,
		BetAmount:        betAmount,
		NextPaymentLimit: nextGrade.PaymentLimit,
		NextBetLimit:     nextGrade.BetLimit,
		DepositProgress:  depositProgress,
		BetProgress:      betProgress,
		Progress:         progress,
		IsMaxGrade:       false,
	}, nil
}

func calcVipProgressPercent(current, target float64) float64 {
	if target <= 0 {
		return 100
	}
	if current <= 0 {
		return 0
	}
	progress := current / target * 100
	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}
	return progress
}

func roundVipProgress(value float64) float64 {
	return math.Round(value*100) / 100
}

func (s *sBalance) getUserCurrentAndNextVipGrade(ctx context.Context, siteId int, userId int) (vipGradeProgressInfo, *vipGradeProgressInfo, bool, error) {
	userClient, userConn, err := grpc_client.GetUserClient(ctx)
	if err != nil {
		return vipGradeProgressInfo{}, nil, false, err
	}
	defer userConn.Close()

	ctxWithMeta := metadata.AppendToOutgoingContext(ctx, "site_id", gconv.String(siteId))

	userRes, err := userClient.GetUserBasicInfo(ctxWithMeta, &userv1.GetUserBasicInfoReq{
		Id: int32(userId),
	})
	if err != nil {
		return vipGradeProgressInfo{}, nil, false, err
	}
	if userRes == nil || userRes.User == nil {
		return vipGradeProgressInfo{}, nil, false, fmt.Errorf("用户信息不存在")
	}

	current := vipGradeProgressInfo{
		Id:   userRes.User.GetGradeId(),
		Name: strings.TrimSpace(userRes.User.GetGradeName()),
	}

	gradesRes, err := userClient.GetUserGrades(ctxWithMeta, &userv1.GetUserGradesReq{
		SiteId: int32(siteId),
	})
	if err != nil {
		return vipGradeProgressInfo{}, nil, false, err
	}
	if gradesRes == nil || len(gradesRes.GetData()) == 0 {
		return vipGradeProgressInfo{}, nil, false, fmt.Errorf("站点等级配置不存在")
	}

	grades := make([]vipGradeProgressInfo, 0, len(gradesRes.GetData()))
	for _, grade := range gradesRes.GetData() {
		if grade == nil || grade.GetId() <= 0 {
			continue
		}
		grades = append(grades, vipGradeProgressInfo{
			Id:           grade.GetId(),
			Name:         strings.TrimSpace(grade.GetName()),
			PaymentLimit: grade.GetPaymentLimit(),
			BetLimit:     grade.GetBetLimit(),
		})
	}
	if len(grades) == 0 {
		return vipGradeProgressInfo{}, nil, false, fmt.Errorf("站点等级配置不存在")
	}

	sort.Slice(grades, func(i, j int) bool {
		return grades[i].Id < grades[j].Id
	})

	var nextGrade *vipGradeProgressInfo
	for i := range grades {
		grade := grades[i]
		if grade.Id == current.Id && grade.Name != "" {
			current.Name = grade.Name
		}
		if grade.Id > current.Id {
			next := grade
			nextGrade = &next
			break
		}
	}

	if current.Name == "" && current.Id > 0 {
		current.Name = fmt.Sprintf("VIP%d", current.Id)
	}

	if nextGrade == nil {
		return current, nil, true, nil
	}

	return current, nextGrade, false, nil
}

// WashCodeRebate 执行洗码返水
func (s *sBalance) WashCodeRebate(ctx context.Context, req *v1.WashCodeRebateReq) (*v1.WashCodeRebateRes, error) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok || userId == 0 {
		return &v1.WashCodeRebateRes{
			Success: false,
			Message: "未登录",
		}, nil
	}

	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return &v1.WashCodeRebateRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}

	siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
	if siteCode == "" {
		siteCode = registry.GetDBConfigManager().GetSiteCodeBySiteID(int64(siteId))
	}
	if siteCode != "" && ctx.Value("dynamic_db_group") == nil {
		ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
	}

	lastBetId, err := s.getLastRebateBetCursor(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Warningf(ctx, "查询上次洗码游标失败，按无历史处理: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		lastBetId = 0
	}

	gameClient, gameConn, err := grpc_client.GetGameClient(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "连接game服务失败: %v", err)
		return &v1.WashCodeRebateRes{
			Success: false,
			Message: "连接游戏服务失败",
		}, nil
	}
	defer gameConn.Close()

	mdMap := map[string]string{
		"site_id": gconv.String(siteId),
		"user_id": gconv.String(userId),
	}
	if siteCode != "" {
		mdMap["site_code"] = siteCode
	}
	if username := middleware.GetUsernameFromContext(ctx); username != "" {
		mdMap["username"] = username
	}
	if currency := middleware.GetCurrencyFromContext(ctx); currency != "" {
		mdMap["currency"] = currency
	}
	ctxWithMeta := metadata.NewOutgoingContext(ctx, metadata.New(mdMap))

	quotaRes, err := gameClient.GetWashCodeQuota(ctxWithMeta, &gamev1.GetWashCodeQuotaReq{
		LastBetId: lastBetId,
	})
	if err != nil {
		g.Log().Errorf(ctx, "获取可洗码额度失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.WashCodeRebateRes{
			Success: false,
			Message: "获取可洗码额度失败",
		}, nil
	}
	if quotaRes == nil || !quotaRes.GetSuccess() {
		msg := "获取可洗码额度失败"
		if quotaRes != nil && quotaRes.GetMessage() != "" {
			msg = quotaRes.GetMessage()
		}
		return &v1.WashCodeRebateRes{
			Success: false,
			Message: msg,
		}, nil
	}

	validBetAmount := quotaRes.GetValidBetAmount()
	maxBetId := quotaRes.GetMaxBetId()
	if validBetAmount <= 0 || maxBetId <= lastBetId {
		return &v1.WashCodeRebateRes{
			Success:        false,
			Message:        "暂无可洗码额度",
			ValidBetAmount: validBetAmount,
			MaxBetId:       maxBetId,
		}, nil
	}

	gradeName, rebatePercent, err := s.getUserEgameRebateConfig(ctx, int(siteId), int(userId))
	if err != nil {
		g.Log().Errorf(ctx, "获取用户等级返水比例失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.WashCodeRebateRes{
			Success: false,
			Message: "获取用户等级返水比例失败",
		}, nil
	}
	if rebatePercent <= 0 {
		return &v1.WashCodeRebateRes{
			Success:        false,
			Message:        "当前等级未配置电子游戏返水比例",
			ValidBetAmount: validBetAmount,
			RebatePercent:  rebatePercent,
			GradeName:      gradeName,
			MaxBetId:       maxBetId,
		}, nil
	}

	rebateAmount := math.Round(validBetAmount*rebatePercent*100) / 10000
	if rebateAmount <= 0 {
		return &v1.WashCodeRebateRes{
			Success:        false,
			Message:        "可返水金额为0",
			ValidBetAmount: validBetAmount,
			RebatePercent:  rebatePercent,
			GradeName:      gradeName,
			MaxBetId:       maxBetId,
		}, nil
	}

	var tradeNo string
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

		latestCursor, err := s.getLastRebateBetCursorWithTx(ctx, tx, int(siteId), int(userId))
		if err != nil {
			return err
		}
		if latestCursor != lastBetId {
			return fmt.Errorf("洗码数据已更新，请重试")
		}

		now := gtime.Now()
		balanceOld := balanceUser.Balance
		balanceNew := balanceOld + rebateAmount
		tradeNo = util.GenerateTradeNo(ctx, balanceConsts.LOG_TYPE_REBATE)

		remark := "洗码返水"
		if strings.TrimSpace(gradeName) != "" {
			remark = fmt.Sprintf("洗码返水(%s)", strings.TrimSpace(gradeName))
		}

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
			"trade_type":   balanceConsts.LOG_TYPE_REBATE,
			"trade_no":     tradeNo,
			"money":        rebateAmount,
			"actual_money": rebateAmount,
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
			"trade_type":     balanceConsts.TradeTypeRebate,
			"user_id":        userId,
			"username":       balanceUser.Username,
			"trade_no":       tradeNo,
			"balance_old":    balanceOld,
			"money":          rebateAmount,
			"balance_new":    balanceNew,
			"balance_frozen": balanceUser.BalanceFrozen,
			"status":         balanceConsts.StatusSuccess,
			"remark":         remark,
			"created_at":     now,
			"updated_at":     now,
		}); err != nil {
			return err
		}

		_, err = dao.BalanceLogRebate.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":          siteId,
			"user_id":          userId,
			"username":         balanceUser.Username,
			"trade_no":         tradeNo,
			"last_bet_id":      maxBetId,
			"valid_bet_amount": validBetAmount,
			"money":            rebateAmount,
			"fee":              0,
			"status":           balanceConsts.StatusSuccess,
			"admin_id":         0,
			"created_at":       now,
			"updated_at":       now,
		})
		if err != nil {
			_, fallbackErr := dao.BalanceLogRebate.Ctx(ctx).TX(tx).Insert(g.Map{
				"site_id":          siteId,
				"user_id":          userId,
				"username":         balanceUser.Username,
				"trade_no":         tradeNo,
				"rebate_date":      now,
				"valid_bet_amount": validBetAmount,
				"money":            rebateAmount,
				"fee":              0,
				"status":           balanceConsts.StatusSuccess,
				"admin_id":         0,
				"created_at":       now,
				"updated_at":       now,
			})
			if fallbackErr != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		g.Log().Errorf(ctx, "执行洗码返水失败: site_id=%d, user_id=%d, err=%v", siteId, userId, err)
		return &v1.WashCodeRebateRes{
			Success:        false,
			Message:        err.Error(),
			ValidBetAmount: validBetAmount,
			RebatePercent:  rebatePercent,
			GradeName:      gradeName,
			MaxBetId:       maxBetId,
		}, nil
	}

	return &v1.WashCodeRebateRes{
		Success:        true,
		Message:        "洗码返水成功",
		TradeNo:        tradeNo,
		ValidBetAmount: validBetAmount,
		RebatePercent:  rebatePercent,
		RebateAmount:   rebateAmount,
		GradeName:      gradeName,
		MaxBetId:       maxBetId,
	}, nil
}

func (s *sBalance) getUserEgameRebateConfig(ctx context.Context, siteId int, userId int) (string, float64, error) {
	userClient, userConn, err := grpc_client.GetUserClient(ctx)
	if err != nil {
		return "", 0, err
	}
	defer userConn.Close()

	ctxWithMeta := metadata.AppendToOutgoingContext(ctx, "site_id", gconv.String(siteId))

	userRes, err := userClient.GetUserBasicInfo(ctxWithMeta, &userv1.GetUserBasicInfoReq{
		Id: int32(userId),
	})
	if err != nil {
		return "", 0, err
	}
	if userRes == nil || userRes.User == nil {
		return "", 0, fmt.Errorf("用户信息不存在")
	}

	gradeId := userRes.User.GetGradeId()
	gradeName := strings.TrimSpace(userRes.User.GetGradeName())

	gradesRes, err := userClient.GetUserGrades(ctxWithMeta, &userv1.GetUserGradesReq{
		SiteId: int32(siteId),
	})
	if err != nil {
		return "", 0, err
	}
	if gradesRes == nil || len(gradesRes.GetData()) == 0 {
		return "", 0, fmt.Errorf("站点等级配置不存在")
	}

	for _, grade := range gradesRes.GetData() {
		if grade.GetId() != gradeId {
			continue
		}
		if strings.TrimSpace(grade.GetName()) != "" {
			gradeName = strings.TrimSpace(grade.GetName())
		}
		return gradeName, grade.GetRebatePercentEgame(), nil
	}

	return gradeName, 0, fmt.Errorf("未找到用户等级配置")
}

func (s *sBalance) getLastRebateBetCursor(ctx context.Context, siteId int, userId int) (uint64, error) {
	return s.getLastRebateBetCursorWithTx(ctx, nil, siteId, userId)
}

func (s *sBalance) getLastRebateBetCursorWithTx(ctx context.Context, tx gdb.TX, siteId int, userId int) (uint64, error) {
	var row struct {
		LastBetId uint64 `json:"last_bet_id"`
	}

	model := dao.BalanceLogRebate.Ctx(ctx)
	if tx != nil {
		model = model.TX(tx)
	}
	model = model.
		Where("site_id", siteId).
		Where("user_id", userId).
		Where("status", balanceConsts.StatusSuccess).
		Fields("last_bet_id").
		OrderDesc("id").
		Limit(1)

	if tx != nil {
		model = model.LockUpdate()
	}

	if err := model.Scan(&row); err != nil {
		if isIgnorableRebateCursorErr(err) {
			return 0, nil
		}
		return 0, err
	}

	return row.LastBetId, nil
}

func isIgnorableRebateCursorErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "last_bet_id") {
		return true
	}
	if strings.Contains(msg, "unknown column") || strings.Contains(msg, "error 1054") {
		return true
	}
	if strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "error 1146") {
		return true
	}
	if strings.Contains(msg, "no rows in result set") || strings.Contains(msg, "record not found") {
		return true
	}
	return false
}
