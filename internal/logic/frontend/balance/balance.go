package balance

import (
	"context"
	"fmt"
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
	gamev1 "jh_balance_service/api/grpc/game/v1"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/grpc_client"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
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
		BalanceStatus: 1, // 1=正常
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
			Balance:       "0.00",
			BalanceFrozen: "0.00",
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
				gameBalance = float64(gameBalanceRes.Balance)
				g.Log().Infof(ctx, "查询三方游戏余额成功: game_id=%d, balance=%.2f", lastGameRes.GameId, gameBalance)
			} else {
				g.Log().Warningf(ctx, "查询三方游戏余额失败: %s", gameBalanceRes.Message)
			}
		}
	} else {
		g.Log().Infof(ctx, "用户没有进入过游戏或game_id=0，只返回本地余额")
	}

	// 总余额 = 本地余额 + 三方游戏余额
	totalBalance = localBalance + gameBalance

	// 格式化余额（保留2位小数）
	balance := fmt.Sprintf("%.2f", totalBalance)
	balanceFrozen := fmt.Sprintf("%.2f", balanceUser.BalanceFrozen)

	g.Log().Infof(ctx, "刷新用户余额成功: user_id=%d, local_balance=%.2f, game_balance=%.2f, total_balance=%s, balance_frozen=%s, points=%d",
		userId, localBalance, gameBalance, balance, balanceFrozen, balanceUser.Points)

	res = &v1.RefreshBalanceRes{
		Balance:       balance,
		BalanceFrozen: balanceFrozen,
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
	err := dao.BalanceUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
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

		// 检查余额是否足够
		if balanceUser.Balance < amount {
			return fmt.Errorf("余额不足，当前余额: %.2f, 需要: %.2f", balanceUser.Balance, amount)
		}

		// 计算新余额：扣除余额，增加冻结金额
		newBalance = balanceUser.Balance - amount
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

	err := dao.BalanceUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

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

			if balanceUser.BalanceFrozen < amount {
				return fmt.Errorf("用户冻结金额不足")
			}

			newBalance = balanceUser.Balance
			newBalanceFrozen = balanceUser.BalanceFrozen - amount
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
				"status":     balanceConsts.StatusSuccess, // 2=成功
				"updated_at": gtime.Now(),
			})
		if err != nil {
			return err
		}

		// 更新资金日志状态
		_, err = dao.BalanceLog.Ctx(ctx).
			Where("trade_no", orderNo).
			Update(g.Map{
				"status":     balanceConsts.StatusSuccess, // 2=成功
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
				"status":         balanceConsts.StatusSuccess, // 2=成功
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

	err := dao.BalanceUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {

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

			if balanceUser.BalanceFrozen < amount {
				return fmt.Errorf("用户冻结金额不足")
			}
			newBalanceFrozen = balanceUser.BalanceFrozen - amount
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
	err := dao.BalanceUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
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

	// 查询status=1的支付渠道，按sort从大到小排序
	var payments []*entity.SitePayment
	err := dao.SitePayment.Ctx(ctx).
		Where("site_id", siteId).
		Where("status", 1).
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

	// 查询status=1的提现渠道，按sort从大到小排序
	var withdraws []*entity.SiteWithdraw
	err := dao.SiteWithdraw.Ctx(ctx).
		Where("site_id", siteId).
		Where("status", 1).
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
	// 生成 UUID 作为 nonce
	nonce := uuid.New().String()

	// 获取当前时间戳
	timestamp := time.Now().Unix()

	// 设置过期时间（60秒）
	expiresIn := int64(60)

	g.Log().Infof(ctx, "生成 nonce: %s, timestamp: %d, expiresIn: %d", nonce, timestamp, expiresIn)

	// 将 nonce 存储到 Redis，设置过期时间
	// 使用 user_id 作为 key 的一部分，确保每个用户只能有一个有效的 nonce
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

	redisKey := fmt.Sprintf("deposit:nonce:%d:%d", siteID, userID)

	g.Log().Infof(ctx, "存储 nonce 到 Redis: key=%s, value=%s, expiresIn=%d", redisKey, nonce, expiresIn)

	// 存储 nonce 到 Redis
	err := g.Redis().SetEX(ctx, redisKey, nonce, expiresIn)
	if err != nil {
		g.Log().Errorf(ctx, "存储充值 nonce 失败: %v, key=%s, value=%s, expiresIn=%d", err, redisKey, nonce, expiresIn)
		return nil, gerror.Newf("生成 nonce 失败: %v", err)
	}

	g.Log().Infof(ctx, "Redis 存储成功")

	// 验证是否真的存储成功
	val, err := g.Redis().Get(ctx, redisKey)
	if err != nil {
		g.Log().Errorf(ctx, "验证 Redis 存储失败: %v", err)
	} else {
		g.Log().Infof(ctx, "验证 Redis 存储成功: key=%s, value=%s", redisKey, val.String())
	}

	result := &v1.GetDepositNonceRes{
		Nonce:     nonce,
		Timestamp: timestamp,
		ExpiresIn: expiresIn,
	}

	g.Log().Infof(ctx, "返回结果: nonce=%s, timestamp=%d, expiresIn=%d", result.Nonce, result.Timestamp, result.ExpiresIn)

	return result, nil
}

// GetWithdrawNonce 获取提现 nonce
func (s *sBalance) GetWithdrawNonce(ctx context.Context, req *v1.GetWithdrawNonceReq) (*v1.GetWithdrawNonceRes, error) {
	// 生成 UUID 作为 nonce
	nonce := uuid.New().String()

	// 获取当前时间戳
	timestamp := time.Now().Unix()

	// 设置过期时间（60秒）
	expiresIn := int64(60)

	g.Log().Infof(ctx, "生成 nonce: %s, timestamp: %d, expiresIn: %d", nonce, timestamp, expiresIn)

	// 将 nonce 存储到 Redis，设置过期时间
	// 使用 user_id 作为 key 的一部分，确保每个用户只能有一个有效的 nonce
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

	redisKey := fmt.Sprintf("withdraw:nonce:%d:%d", siteID, userID)

	g.Log().Infof(ctx, "存储 nonce 到 Redis: key=%s, value=%s, expiresIn=%d", redisKey, nonce, expiresIn)

	// 存储 nonce 到 Redis
	err := g.Redis().SetEX(ctx, redisKey, nonce, expiresIn)
	if err != nil {
		g.Log().Errorf(ctx, "存储提现 nonce 失败: %v, key=%s, value=%s, expiresIn=%d", err, redisKey, nonce, expiresIn)
		return nil, gerror.Newf("生成 nonce 失败: %v", err)
	}

	g.Log().Infof(ctx, "Redis 存储成功")

	// 验证是否真的存储成功
	val, err := g.Redis().Get(ctx, redisKey)
	if err != nil {
		g.Log().Errorf(ctx, "验证 Redis 存储失败: %v", err)
	} else {
		g.Log().Infof(ctx, "验证 Redis 存储成功: key=%s, value=%s", redisKey, val.String())
	}

	result := &v1.GetWithdrawNonceRes{
		Nonce:     nonce,
		Timestamp: timestamp,
		ExpiresIn: expiresIn,
	}

	g.Log().Infof(ctx, "返回结果: nonce=%s, timestamp=%d, expiresIn=%d", result.Nonce, result.Timestamp, result.ExpiresIn)

	return result, nil
}
