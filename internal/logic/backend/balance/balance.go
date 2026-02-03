package balance

import (
	"context"
	"fmt"
	v1 "jh_balance_service/api/backend/balance/v1"
	gamev1 "jh_balance_service/api/grpc/game/backend/v1"
	userv1 "jh_balance_service/api/grpc/user/backend/v1"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	withdrawReview "jh_balance_service/internal/domain/withdraw_review"
	"jh_balance_service/internal/grpc_client"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/service/backend"
	"jh_balance_service/internal/tracing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

type (
	sBalance struct{}
)

func init() {
	backend.RegisterBalance(&sBalance{})
}

// GetBalanceChanges 获取账变记录
func (s *sBalance) GetBalanceChanges(ctx context.Context, req *v1.GetBalanceChangesReq) (*v1.GetBalanceChangesRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetBalanceChanges", trace.WithAttributes(
		attribute.String("method", "GetBalanceChanges"),
		attribute.String("username", req.Username),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	siteId := 1
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	query := dao.BalanceChange.Ctx(ctx).Where("site_id = ?", siteId)

	if req.Username != "" {
		query = query.Where("username = ?", req.Username)
	}
	if req.ChangeType > 0 {
		if req.ChangeType == int32(balanceConsts.ChangeTypeIncome) { // 入款
			query = query.WhereIn("trade_type", balanceConsts.IncomeTradeTypes)
		} else if req.ChangeType == int32(balanceConsts.ChangeTypePay) { // 出款
			query = query.WhereIn("trade_type", balanceConsts.PayTradeTypes)
		}
	}
	if req.TradeType > 0 {
		query = query.Where("trade_type = ?", req.TradeType)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	count, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("查询账变记录总数失败: %v", err)
	}

	var changes []*entity.BalanceChange
	offset := (page - 1) * size
	err = query.Order("created_at DESC").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(&changes)
	if err != nil {
		return nil, fmt.Errorf("查询账变记录列表失败: %v", err)
	}

	var list []*v1.BalanceChangeInfo
	for _, change := range changes {
		info := &v1.BalanceChangeInfo{
			Id:            int32(change.Id),
			TradeType:     int32(change.TradeType),
			TradeTypeName: balanceConsts.GetTradeTypeName(int(change.TradeType)),
			UserId:        int32(change.UserId),
			Username:      change.Username,
			TradeNo:       change.TradeNo,
			BalanceOld:    change.BalanceOld,
			Money:         change.Money,
			BalanceNew:    change.BalanceNew,
			BalanceFrozen: change.BalanceFrozen,
			Status:        int32(change.Status),
			StatusName:    balanceConsts.GetStatusName(change.Status),
			Remark:        change.Remark,
			CreatedAt:     change.CreatedAt.String(),
			ChangeType:    int32(balanceConsts.GetChangeTypeByTradeType(int(change.TradeType))),
		}
		list = append(list, info)
	}

	return &v1.GetBalanceChangesRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// GetRechargePayments 获取充值记录
func (s *sBalance) GetRechargePayments(ctx context.Context, req *v1.GetRechargePaymentsReq) (*v1.GetRechargePaymentsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetRechargePayments", trace.WithAttributes(
		attribute.String("method", "GetRechargePayments"),
		attribute.String("username", req.Username),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	siteId := 1
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	query := dao.BalanceLogPayment.Ctx(ctx).Where("site_id = ?", siteId)

	if req.Username != "" {
		query = query.Where("username = ?", req.Username)
	}
	if req.Gateway > 0 {
		query = query.Where("gateway = ?", req.Gateway)
	}
	if req.PaymentId > 0 {
		query = query.Where("payment_id = ?", req.PaymentId)
	}
	if req.AccountId > 0 {
		query = query.Where("payment_account_id = ?", req.AccountId)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.TradeNo != "" {
		query = query.Where("trade_no = ?", req.TradeNo)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	count, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("查询充值记录总数失败: %v", err)
	}

	var payments []*entity.BalanceLogPayment
	offset := (page - 1) * size
	err = query.Order("created_at DESC").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(&payments)
	if err != nil {
		return nil, fmt.Errorf("查询充值记录列表失败: %v", err)
	}

	var list []*v1.RechargePaymentInfo
	for _, payment := range payments {
		info := &v1.RechargePaymentInfo{
			Id:                 int64(payment.Id),
			UserId:             int32(payment.UserId),
			Username:           payment.Username,
			ActivityRechargeId: int32(payment.ActivityRechargeId),
			Gateway:            int32(payment.Gateway),
			GatewayName:        s.getGatewayName(payment.Gateway),
			PaymentId:          int32(payment.PaymentId),
			PaymentName:        "",
			PaymentAccountId:   int32(payment.PaymentAccountId),
			BankValue:          payment.BankValue,
			TradeNo:            payment.TradeNo,
			Money:              payment.Money,
			Fee:                payment.Fee,
			Status:             int32(payment.Status),
			StatusName:         s.getPaymentStatusName(payment.Status),
			AdminId:            int32(payment.AdminId),
			AdminName:          "",
			Remark:             payment.Remark,
			CreatedAt:          payment.CreatedAt.String(),
			UpdatedAt:          payment.UpdatedAt.String(),
		}
		list = append(list, info)
	}

	return &v1.GetRechargePaymentsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// GetRechargeManuals 获取后台加款记录
func (s *sBalance) GetRechargeManuals(ctx context.Context, req *v1.GetRechargeManualsReq) (*v1.GetRechargeManualsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetRechargeManuals", trace.WithAttributes(
		attribute.String("method", "GetRechargeManuals"),
		attribute.String("username", req.Username),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	siteId := 1
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	query := dao.BalanceLogManual.Ctx(ctx).Where("site_id = ?", siteId).Where("trade_type = ?", balanceConsts.TradeTypeManualAdd) // 丢失补款

	if req.Username != "" {
		query = query.Where("username = ?", req.Username)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	count, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("查询后台加款记录总数失败: %v", err)
	}

	var manuals []*entity.BalanceLogManual
	offset := (page - 1) * size
	err = query.Order("created_at DESC").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(&manuals)
	if err != nil {
		return nil, fmt.Errorf("查询后台加款记录列表失败: %v", err)
	}

	var list []*v1.RechargeManualInfo
	for _, manual := range manuals {
		info := &v1.RechargeManualInfo{
			Id:         int64(manual.Id),
			UserId:     int32(manual.UserId),
			Username:   manual.Username,
			TradeNo:    manual.TradeNo,
			Money:      manual.Money,
			Status:     int32(manual.Status),
			StatusName: balanceConsts.GetStatusName(manual.Status),
			AdminId:    int32(manual.AdminId),
			AdminName:  "",
			Remark:     manual.Remark,
			CreatedAt:  manual.CreatedAt.String(),
		}
		list = append(list, info)
	}

	return &v1.GetRechargeManualsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// ConfirmPaymentOrder 确认支付订单
func (s *sBalance) ConfirmPaymentOrder(ctx context.Context, req *v1.ConfirmPaymentOrderReq) (*v1.ConfirmPaymentOrderRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.ConfirmPaymentOrder", trace.WithAttributes(
		attribute.String("method", "ConfirmPaymentOrder"),
		attribute.Int64("id", req.Id),
	))
	defer span.End()

	// 查找支付订单
	var payment entity.BalanceLogPayment
	err := dao.BalanceLogPayment.Ctx(ctx).Where("id = ?", req.Id).Scan(&payment)
	if err != nil {
		return nil, fmt.Errorf("查找支付订单失败: %v", err)
	}

	if payment.Id == 0 {
		return &v1.ConfirmPaymentOrderRes{
			Success: false,
			Message: "支付订单不存在",
		}, nil
	}

	if payment.Status == balanceConsts.StatusSuccess {
		return &v1.ConfirmPaymentOrderRes{
			Success: false,
			Message: "订单已经确认，请勿重复操作",
		}, nil
	}

	// 开启事务 - 使用 Ctx(ctx) 保持动态数据库路由
	err = dao.BalanceLogPayment.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新订单状态
		_, err := dao.BalanceLogPayment.Ctx(ctx).TX(tx).
			Where("id = ?", req.Id).
			Update(g.Map{
				"status":     balanceConsts.StatusSuccess,
				"remark":     req.Remark,
				"updated_at": time.Now(),
			})
		if err != nil {
			return fmt.Errorf("更新订单状态失败: %v", err)
		}

		// TODO: 调用用户服务更新余额
		// 这里需要调用 jh_balance_service 的用户余额更新接口

		return nil
	})

	if err != nil {
		return &v1.ConfirmPaymentOrderRes{
			Success: false,
			Message: fmt.Sprintf("确认订单失败: %v", err),
		}, nil
	}

	return &v1.ConfirmPaymentOrderRes{
		Success: true,
		Message: "订单确认成功",
	}, nil
}

// GetWithdraws 获取提现记录
func (s *sBalance) GetWithdraws(ctx context.Context, req *v1.GetWithdrawsReq) (*v1.GetWithdrawsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetWithdraws", trace.WithAttributes(
		attribute.String("method", "GetWithdraws"),
		attribute.String("username", req.Username),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	siteId := 1
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	query := dao.BalanceLogWithdraw.Ctx(ctx).Where("site_id = ?", siteId)

	if req.Username != "" {
		query = query.Where("username = ?", req.Username)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.TradeNo != "" {
		query = query.Where("trade_no = ?", req.TradeNo)
	}
	if req.StartTime != "" {
		query = query.Where("updated_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("updated_at <= ?", req.EndTime)
	}

	count, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("查询提现记录总数失败: %v", err)
	}

	var withdraws []*entity.BalanceLogWithdraw
	offset := (page - 1) * size
	err = query.Order("updated_at DESC").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(&withdraws)
	if err != nil {
		return nil, fmt.Errorf("查询提现记录列表失败: %v", err)
	}

	var list []*v1.WithdrawInfo
	for _, withdraw := range withdraws {
		info := &v1.WithdrawInfo{
			Id:          int64(withdraw.Id),
			UserId:      int32(withdraw.UserId),
			UserLevelId: int32(withdraw.UserLevelId),
			Username:    withdraw.Username,
			TradeNo:     withdraw.TradeNo,
			Money:       withdraw.Money,
			Fee:         withdraw.Fee,
			BankName:    withdraw.BankName,
			CardAccount: withdraw.CardAccount,
			CardNo:      s.maskCardNo(withdraw.CardNo),
			DepositBank: withdraw.DepositBank,
			Status:      int32(withdraw.Status),
			StatusName:  s.getWithdrawStatusName(withdraw.Status),
			AdminId:     int32(withdraw.AdminId),
			AdminName:   "",
			Remark:      withdraw.Remark,
			CreatedAt:   withdraw.CreatedAt.String(),
			UpdatedAt:   withdraw.UpdatedAt.String(),
		}
		list = append(list, info)
	}

	return &v1.GetWithdrawsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// GetWithdrawManuals 获取后台提现记录
func (s *sBalance) GetWithdrawManuals(ctx context.Context, req *v1.GetWithdrawManualsReq) (*v1.GetWithdrawManualsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetWithdrawManuals", trace.WithAttributes(
		attribute.String("method", "GetWithdrawManuals"),
		attribute.String("username", req.Username),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	siteId := 1
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	query := dao.BalanceLogManual.Ctx(ctx).Where("site_id = ?", siteId).Where("trade_type = ?", balanceConsts.TradeTypeManualCut) // 多出扣款

	if req.Username != "" {
		query = query.Where("username = ?", req.Username)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	count, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("查询后台提现记录总数失败: %v", err)
	}

	var manuals []*entity.BalanceLogManual
	offset := (page - 1) * size
	err = query.Order("created_at DESC").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(&manuals)
	if err != nil {
		return nil, fmt.Errorf("查询后台提现记录列表失败: %v", err)
	}

	var list []*v1.WithdrawManualInfo
	for _, manual := range manuals {
		info := &v1.WithdrawManualInfo{
			Id:         int64(manual.Id),
			UserId:     int32(manual.UserId),
			Username:   manual.Username,
			TradeNo:    manual.TradeNo,
			Money:      manual.Money,
			Status:     int32(manual.Status),
			StatusName: balanceConsts.GetStatusName(manual.Status),
			AdminId:    int32(manual.AdminId),
			AdminName:  "",
			Remark:     manual.Remark,
			CreatedAt:  manual.CreatedAt.String(),
		}
		list = append(list, info)
	}

	return &v1.GetWithdrawManualsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// GetWithdrawReview 获取提现审核信息
func (s *sBalance) GetWithdrawReview(ctx context.Context, req *v1.GetWithdrawReviewReq) (*v1.GetWithdrawReviewRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetWithdrawReview", trace.WithAttributes(
		attribute.String("method", "GetWithdrawReview"),
		attribute.Int64("id", req.Id),
	))
	defer span.End()

	siteId := int64(1)

	// 1. 查找提现记录
	var withdraw entity.BalanceLogWithdraw
	err := dao.BalanceLogWithdraw.Ctx(ctx).Where("id = ?", req.Id).Scan(&withdraw)
	if err != nil {
		return nil, fmt.Errorf("查找提现记录失败: %v", err)
	}

	if withdraw.Id == 0 {
		return nil, fmt.Errorf("提现记录不存在")
	}

	// 2. 查询用户余额信息
	var balanceUser entity.BalanceUser
	err = dao.BalanceUser.Ctx(ctx).
		Where("site_id = ?", siteId).
		Where("user_id = ?", withdraw.UserId).
		Scan(&balanceUser)
	if err != nil {
		return nil, fmt.Errorf("查询用户余额失败: %v", err)
	}

	// 3. 查询总充值金额
	totalRecharge, err := dao.BalanceLogPayment.GetTotalRechargeByUserId(ctx, siteId, int64(withdraw.UserId))
	if err != nil {
		g.Log().Warningf(ctx, "查询总充值金额失败: %v", err)
		totalRecharge = 0
	}

	// 4. 查询今日充值金额
	todayRecharge, err := dao.BalanceLogPayment.GetTodayRechargeByUserId(ctx, siteId, int64(withdraw.UserId))
	if err != nil {
		g.Log().Warningf(ctx, "查询今日充值金额失败: %v", err)
		todayRecharge = 0
	}

	// 5. 查询今日提现金额
	todayWithdraw, err := dao.BalanceLogWithdraw.GetTodayWithdrawByUserId(ctx, siteId, int64(withdraw.UserId))
	if err != nil {
		g.Log().Warningf(ctx, "查询今日提现金额失败: %v", err)
		todayWithdraw = 0
	}

	// 6. 查询总提现次数
	totalWithdrawCount, err := dao.BalanceLogWithdraw.GetTotalWithdrawCountByUserId(ctx, siteId, int64(withdraw.UserId))
	if err != nil {
		g.Log().Warningf(ctx, "查询总提现次数失败: %v", err)
		totalWithdrawCount = 0
	}

	// 7. 从 app 服务获取用户基本信息（注册时间、注册IP、VIP等级）
	var registerTime, registerIp string
	var vipLevel int32

	userClient, userConn, err := grpc_client.GetUserClient(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取 User 客户端失败: %v", err)
	} else {
		defer userConn.Close()

		// 创建带有 metadata 的上下文，传递 site_id
		ctxWithMeta := metadata.AppendToOutgoingContext(ctx, "site_id", gconv.String(siteId))

		userInfoResp, err := userClient.GetUserBasicInfo(ctxWithMeta, &userv1.GetUserBasicInfoReq{
			Id: int32(withdraw.UserId),
		})
		if err != nil {
			g.Log().Warningf(ctx, "调用 GetUserBasicInfo 失败: %v", err)
		} else if userInfoResp != nil && userInfoResp.User != nil {
			registerTime = userInfoResp.User.RegisterTime
			registerIp = userInfoResp.User.RegisterIp
			vipLevel = int32(userInfoResp.User.GradeId)
		}
	}

	// 8. 银行卡脱敏处理（保留前4位和后4位）
	bankCardMasked := s.maskCardNo(withdraw.CardNo)

	// 9. 从 game 服务获取有效投注/总输赢/打码倍数
	validBet := 0.0
	totalWinLose := 0.0
	totalBetMultiple := 0.0

	siteCode, _ := middleware.GetSiteCodeFromContext(ctx)

	endTime := ""
	if withdraw.CreatedAt != nil {
		endTime = withdraw.CreatedAt.Format("Y-m-d H:i:s")
	}
	if endTime == "" {
		endTime = time.Now().Format("2006-01-02 15:04:05")
	}

	lastWithdrawTime, err := dao.BalanceLogWithdraw.GetLastWithdrawTimeByUser(ctx, siteId, int64(withdraw.UserId), withdraw.CreatedAt)
	if err != nil {
		g.Log().Warningf(ctx, "获取上次提现时间失败: %v", err)
	}

	startTime := ""
	if lastWithdrawTime != nil {
		startTime = lastWithdrawTime.Format("Y-m-d H:i:s")
	} else if registerTime != "" {
		startTime = registerTime
	} else {
		startTime = endTime
	}

	if siteCode == "" {
		g.Log().Warningf(ctx, "site_code 为空，跳过 GetUserBetSummary 调用")
	} else {
		gameClient, gameConn, err := grpc_client.GetGameBackendClient(ctx)
		if err != nil {
			g.Log().Warningf(ctx, "获取 Game 后台客户端失败: %v", err)
		} else {
			defer gameConn.Close()

			md := metadata.New(map[string]string{
				"site_id":   gconv.String(siteId),
				"site_code": siteCode,
				"user_id":   gconv.String(withdraw.UserId),
			})
			ctxWithMeta := metadata.NewOutgoingContext(ctx, md)

			summaryRes, err := gameClient.GetUserBetSummary(ctxWithMeta, &gamev1.GetUserBetSummaryReq{
				StartTime: startTime,
				EndTime:   endTime,
			})
			if err != nil {
				g.Log().Warningf(ctx, "调用 GetUserBetSummary 失败: %v", err)
			} else if summaryRes != nil && summaryRes.Success && summaryRes.Summary != nil {
				validBet = gconv.Float64(summaryRes.Summary.ValidAmount)
				totalWinLose = gconv.Float64(summaryRes.Summary.NetAmount)
				totalBetMultiple = gconv.Float64(summaryRes.Summary.TurnoverMultiple)
			}
		}
	}

	// 构建审核信息响应
	reviewInfo := &v1.WithdrawReviewInfo{
		// ========== 第一块：提现核心信息 ==========
		UserId:         int32(withdraw.UserId),
		Username:       withdraw.Username,
		WithdrawAmount: withdraw.Money,
		BankCardMasked: bankCardMasked,
		ApplyTime:      withdraw.CreatedAt.String(),
		CurrentBalance: balanceUser.Balance,
		FrozenBalance:  balanceUser.BalanceFrozen,

		// ========== 第二块：资金合规与流水 ==========
		TotalRecharge: totalRecharge,
		ValidBet:      validBet,
		// RequiredFlow:  0, // 待实现
		// RemainingFlow: 0, // 待实现
		TodayRecharge: todayRecharge,
		TodayWithdraw: todayWithdraw,

		// ========== 第三块：扩展信息 ==========
		RegisterTime:       registerTime,
		RegisterIp:         registerIp,
		VipLevel:           vipLevel,
		TotalWithdrawCount: int32(totalWithdrawCount),
		TotalWinLose:       totalWinLose,
		TotalBetMultiple:   totalBetMultiple,

		// ========== 其他信息（用于后续处理） ==========
		WithdrawId:  int64(withdraw.Id),
		TradeNo:     withdraw.TradeNo,
		BankName:    withdraw.BankName,
		CardAccount: withdraw.CardAccount,
		CardNo:      withdraw.CardNo, // 审核时显示完整卡号
		DepositBank: withdraw.DepositBank,
		Status:      int32(withdraw.Status),
		Fee:         withdraw.Fee,
		Remark:      withdraw.Remark,
	}

	return &v1.GetWithdrawReviewRes{
		Data: reviewInfo,
	}, nil
}

// DealWithWithdraw 处理提现
func (s *sBalance) DealWithWithdraw(ctx context.Context, req *v1.DealWithWithdrawReq) (*v1.DealWithWithdrawRes, error) {
	ctx, span := tracing.StartSpan(ctx, "balance.DealWithWithdraw", trace.WithAttributes(
		attribute.String("method", "DealWithWithdraw"),
		attribute.Int64("id", req.Id),
		attribute.Int("type", int(req.Type)),
	))
	defer span.End()

	// 查找提现记录
	var withdraw entity.BalanceLogWithdraw
	err := dao.BalanceLogWithdraw.Ctx(ctx).Where("id = ?", req.Id).Scan(&withdraw)
	if err != nil {
		return nil, fmt.Errorf("查找提现记录失败: %v", err)
	}

	if withdraw.Id == 0 {
		return &v1.DealWithWithdrawRes{
			Success: false,
			Message: "提现记录不存在",
		}, nil
	}

	if withdraw.Status != balanceConsts.StatusPending {
		return &v1.DealWithWithdrawRes{
			Success: false,
			Message: "该提现记录已处理，请勿重复操作",
		}, nil
	}

	// 创建 Domain 服务
	reviewDomain := withdrawReview.New(ctx)

	// 准备审核请求
	reviewReq := &withdrawReview.ReviewRequest{
		WithdrawID: req.Id,
		Type:       withdrawReview.ReviewType(req.Type),
		Fee:        req.Fee,
		Remark:     req.Remark,
	}

	// 开启事务处理
	var result *withdrawReview.ReviewResult
	err = dao.BalanceLogWithdraw.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var txErr error
		result, txErr = reviewDomain.ProcessWithdrawReview(tx, reviewReq, &withdraw)
		return txErr
	})

	if err != nil {
		return &v1.DealWithWithdrawRes{
			Success: false,
			Message: fmt.Sprintf("处理提现失败: %v", err),
		}, nil
	}

	return &v1.DealWithWithdrawRes{
		Success: result.Success,
		Message: result.Message,
	}, nil
}

// 辅助方法
func (s *sBalance) getPaymentStatusName(status int) string {
	names := map[int]string{
		balanceConsts.StatusPending: "未付",
		balanceConsts.StatusSuccess: "成功",
		balanceConsts.StatusFail:    "失败",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return "未知"
}

func (s *sBalance) getWithdrawStatusName(status int) string {
	names := map[int]string{
		balanceConsts.StatusPending:    "待审核",
		balanceConsts.StatusProcessing: "处理中",
		balanceConsts.StatusSuccess:    "成功",
		balanceConsts.StatusFail:       "失败",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return "未知"
}

func (s *sBalance) getGatewayName(gateway int) string {
	names := map[int]string{
		1: "在线支付",
		2: "银行转账",
		3: "支付宝",
		4: "微信支付",
	}
	if name, ok := names[gateway]; ok {
		return name
	}
	return "未知"
}

func (s *sBalance) maskCardNo(cardNo string) string {
	if len(cardNo) <= 8 {
		return cardNo
	}
	return cardNo[:4] + "****" + cardNo[len(cardNo)-4:]
}

// QueryUserBalance 查询用户余额
func (s *sBalance) QueryUserBalance(ctx context.Context, req *v1.QueryUserBalanceReq) (*v1.QueryUserBalanceRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.QueryUserBalance", trace.WithAttributes(
		attribute.String("method", "QueryUserBalance"),
		attribute.String("username", string(req.UserId)),
	))
	defer span.End()

	if req.UserId == 0 {
		return nil, fmt.Errorf("用户id不能为空")
	}

	siteId := 1

	userId := req.UserId

	// 查询用户余额
	var balance entity.BalanceUser
	err := dao.BalanceUser.Ctx(ctx).Where("site_id = ? AND user_id = ?", siteId, userId).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("查询用户余额失败: %v", err)
	}

	if balance.Id == 0 {
		return nil, fmt.Errorf("用户不存在")
	}

	balanceInfo := &v1.UserBalanceInfo{
		UserId:         int32(balance.UserId),
		Username:       balance.Username,
		Balance:        balance.Balance,
		BalanceFrozen:  balance.BalanceFrozen,
		Points:         float64(balance.Points),
		LastUpdateTime: balance.UpdatedAt.String(),
	}

	return &v1.QueryUserBalanceRes{
		Data: balanceInfo,
	}, nil
}

// QueryGameBalance 查询游戏余额
func (s *sBalance) QueryGameBalance(ctx context.Context, req *v1.QueryGameBalanceReq) (*v1.QueryGameBalanceRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.QueryGameBalance", trace.WithAttributes(
		attribute.String("method", "QueryGameBalance"),
		attribute.Int("game_id", int(req.GameId)),
		attribute.Int("user_id", int(req.UserId)),
	))
	defer span.End()

	if req.GameId <= 0 {
		return nil, fmt.Errorf("游戏ID不能为空")
	}
	if req.UserId <= 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// TODO: 调用游戏服务查询第三方余额
	// 暂时写死数据
	gameBalanceInfo := &v1.GameBalanceInfo{
		GameId:         req.GameId,
		GameName:       "测试游戏",
		UserId:         req.UserId,
		Username:       "testuser",
		Balance:        1000.00,
		LastUpdateTime: "2024-01-01 10:00:00",
	}

	return &v1.QueryGameBalanceRes{
		Data: gameBalanceInfo,
	}, nil
}

// ManualUserBalance 手动操作用户余额
func (s *sBalance) ManualUserBalance(ctx context.Context, req *v1.ManualUserBalanceReq) (*v1.ManualUserBalanceRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.ManualUserBalance", trace.WithAttributes(
		attribute.String("method", "ManualUserBalance"),
		attribute.Int("user_id", int(req.UserId)),
		attribute.Int("type", int(req.Type)),
		attribute.Float64("money", req.Money),
	))
	defer span.End()

	if req.UserId <= 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// 验证操作类型是否有效（1-6为有效类型）
	validTypes := map[int32]bool{
		1: true, // TYPE_ADD_RECHARGE
		2: true, // TYPE_ADD_BONUS
		3: true, // TYPE_ADD_REBATE
		4: true, // TYPE_ADD_DIRECT
		5: true, // TYPE_CUT_WITHDRAW
		6: true, // TYPE_CUT_DIRECT
	}
	if !validTypes[req.Type] {
		return nil, fmt.Errorf("操作类型无效")
	}

	if req.Money <= 0 {
		return nil, fmt.Errorf("操作金额必须大于0")
	}

	siteId := 1

	// 判断是加款还是扣款
	isAddType := req.Type == 1 || req.Type == 2 || req.Type == 3 || req.Type == 4
	isDeductType := req.Type == 5 || req.Type == 6

	// 开启事务 - 使用 Ctx(ctx) 保持动态数据库路由
	err := dao.BalanceUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询用户当前余额
		var balance entity.BalanceUser
		err := dao.BalanceUser.Ctx(ctx).TX(tx).Where("site_id = ? AND user_id = ?", siteId, req.UserId).Scan(&balance)
		if err != nil {
			return fmt.Errorf("查询用户余额失败: %v", err)
		}

		if balance.Id == 0 {
			return fmt.Errorf("用户不存在")
		}

		balanceOld := balance.Balance
		var balanceNew float64
		var tradeType int
		var remark string

		if isAddType { // 加款
			balanceNew = balanceOld + req.Money
			tradeType = int(req.Type)
			remark = balanceConsts.GetTypeName(int(req.Type))
		} else if isDeductType { // 扣款
			if balanceOld < req.Money {
				return fmt.Errorf("余额不足")
			}
			balanceNew = balanceOld - req.Money
			tradeType = int(req.Type)
			remark = balanceConsts.GetTypeName(int(req.Type))
		}

		if req.Remark != "" {
			remark = req.Remark
		}

		// 获取当前时间
		now := time.Now()

		// 更新用户余额
		_, err = dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("id = ?", balance.Id).
			Update(g.Map{
				"balance":    balanceNew,
				"updated_at": now,
			})
		if err != nil {
			return fmt.Errorf("更新用户余额失败: %v", err)
		}

		// 生成流水号
		tradeNo := fmt.Sprintf("M%d%d", req.UserId, time.Now().Unix())

		// 记录账变
		_, err = dao.BalanceChange.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"trade_type":     balanceConsts.GetBalanceChangeTradeType(int(req.Type)),
			"user_id":        req.UserId,
			"username":       "testuser", // TODO: 从用户服务获取
			"trade_no":       tradeNo,
			"balance_old":    balanceOld,
			"money":          req.Money,
			"balance_new":    balanceNew,
			"balance_frozen": balance.BalanceFrozen,
			"status":         balanceConsts.StatusSuccess,
			"remark":         remark,
			"created_at":     now,
			"updated_at":     now,
		})
		if err != nil {
			return fmt.Errorf("记录账变失败: %v", err)
		}

		// 记录手动操作日志
		_, err = dao.BalanceLogManual.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":    siteId,
			"trade_type": tradeType,
			"user_id":    req.UserId,
			"username":   "testuser", // TODO: 从用户服务获取
			"trade_no":   tradeNo,
			"money":      req.Money,
			"status":     balanceConsts.StatusSuccess,
			"admin_id":   1, // TODO: 从上下文获取管理员ID
			"remark":     remark,
			"created_at": now,
			"updated_at": now,
		})
		if err != nil {
			return fmt.Errorf("记录操作日志失败: %v", err)
		}

		return nil
	})

	if err != nil {
		return &v1.ManualUserBalanceRes{
			Success: false,
			Message: fmt.Sprintf("操作失败: %v", err),
		}, nil
	}

	return &v1.ManualUserBalanceRes{
		Success: true,
		Message: "操作成功",
	}, nil
}

// 辅助方法
func (s *sBalance) getPaymentAccountStatusName(status int) string {
	names := map[int]string{
		0: "禁用",
		1: "启用",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return "未知"
}

func (s *sBalance) maskSensitiveData(data string) string {
	if len(data) <= 8 {
		return "****"
	}
	return data[:4] + "****" + data[len(data)-4:]
}

// GetChangeList 获取交易类型列表
func (s *sBalance) GetChangeList(ctx context.Context, req *v1.GetChangeListReq) (*v1.GetChangeListRes, error) {
	ctx, span := tracing.StartSpan(ctx, "balance.GetChangeList", trace.WithAttributes(
		attribute.String("method", "GetChangeList"),
	))
	defer span.End()

	// 使用常量定义的交易类型列表
	changeList := make(map[int32]string)
	for k, v := range balanceConsts.TradeTypeNames {
		changeList[int32(k)] = v
	}

	return &v1.GetChangeListRes{
		List: changeList,
	}, nil
}

// GetManualList 获取操作类型列表
func (s *sBalance) GetManualList(ctx context.Context, req *v1.GetManualListReq) (*v1.GetManualListRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetManualList", trace.WithAttributes(
		attribute.String("method", "GetManualList"),
	))
	defer span.End()

	// 需要过滤的类型（第三方相关的操作类型）
	excludeTypes := map[int]bool{
		7:  true, // TYPE_GAME_ADD_DIRECT
		8:  true, // TYPE_GAME_CUT_DIRECT
		9:  true, // TYPE_GAME_TRANSFER_IN
		10: true, // TYPE_GAME_TRANSFER_OUT
	}

	// 返回操作类型列表（过滤掉第三方相关的类型）
	manualList := make(map[int32]string)
	for typeCode, typeName := range balanceConsts.TypeNames {
		if !excludeTypes[typeCode] {
			manualList[int32(typeCode)] = typeName
		}
	}

	return &v1.GetManualListRes{
		List: manualList,
	}, nil
}

// GetSitePayments 获取站点支付列表
func (s *sBalance) GetSitePayments(ctx context.Context, req *v1.GetSitePaymentsReq) (*v1.GetSitePaymentsRes, error) {
	ctx, span := tracing.StartSpan(ctx, "balance.GetSitePayments", trace.WithAttributes(
		attribute.String("method", "GetSitePayments"),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	// 构建查询
	model := dao.SitePayment.Ctx(ctx)

	// 条件筛选
	if req.PaymentId > 0 {
		model = model.Where("payment_id", req.PaymentId)
	}
	if req.Status >= 0 {
		model = model.Where("status", req.Status)
	}

	// 查询总数
	count, err := model.Count()
	if err != nil {
		g.Log().Errorf(ctx, "查询站点支付总数失败: %v", err)
		return nil, fmt.Errorf("查询站点支付总数失败")
	}

	// 查询列表
	var sitePayments []*entity.SitePayment
	err = model.Page(int(page), int(size)).
		Order("sort ASC, id DESC").
		Scan(&sitePayments)
	if err != nil {
		g.Log().Errorf(ctx, "查询站点支付列表失败: %v", err)
		return nil, fmt.Errorf("查询站点支付列表失败")
	}

	// 转换为响应格式
	list := make([]*v1.SitePaymentInfo, 0, len(sitePayments))
	for _, sp := range sitePayments {
		list = append(list, &v1.SitePaymentInfo{
			Id:         int32(sp.Id),
			SiteId:     int32(sp.SiteId),
			PaymentId:  int32(sp.PaymentId),
			Name:       sp.Name,
			Code:       sp.Code,
			Gateway:    int32(sp.Gateway),
			ApiUrl:     sp.ApiUrl,
			MerchantNo: sp.MerchantNo,
			Md5Key:     sp.Md5Key,
			EachMin:    sp.EachMin,
			EachMax:    sp.EachMax,
			Sort:       int32(sp.Sort),
			MoneyList:  sp.MoneyList,
			IsInput:    int32(sp.IsInput),
			Logo:       sp.Logo,
			Status:     int32(sp.Status),
			CreatedAt:  sp.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  sp.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &v1.GetSitePaymentsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// UpdateSitePayment 更新站点支付
func (s *sBalance) UpdateSitePayment(ctx context.Context, req *v1.UpdateSitePaymentReq) (*v1.UpdateSitePaymentRes, error) {
	ctx, span := tracing.StartSpan(ctx, "balance.UpdateSitePayment", trace.WithAttributes(
		attribute.String("method", "UpdateSitePayment"),
		attribute.Int("id", int(req.Id)),
	))
	defer span.End()

	if req.Id <= 0 {
		return &v1.UpdateSitePaymentRes{
			Success: false,
			Message: "站点支付ID不能为空",
		}, nil
	}

	// 构建更新数据
	data := g.Map{
		"updated_at": time.Now(),
	}

	if req.Status >= 0 {
		data["status"] = req.Status
	}
	if req.Sort >= 0 {
		data["sort"] = req.Sort
	}
	if req.EachMin >= 0 {
		data["each_min"] = req.EachMin
	}
	if req.EachMax >= 0 {
		data["each_max"] = req.EachMax
	}
	if req.MoneyList != "" {
		data["moneyList"] = req.MoneyList
	}
	if req.IsInput >= 0 {
		data["is_input"] = req.IsInput
	}
	if req.Remark != "" {
		data["remark"] = req.Remark
	}
	if req.Logo != "" {
		data["logo"] = req.Logo
	}

	// 更新数据
	_, err := dao.SitePayment.Ctx(ctx).
		Where("id", req.Id).
		Data(data).
		Update()
	if err != nil {
		g.Log().Errorf(ctx, "更新站点支付失败: %v", err)
		return &v1.UpdateSitePaymentRes{
			Success: false,
			Message: "更新站点支付失败",
		}, nil
	}

	return &v1.UpdateSitePaymentRes{
		Success: true,
		Message: "更新成功",
	}, nil
}

// GetSiteWithdraws 获取站点提现列表
func (s *sBalance) GetSiteWithdraws(ctx context.Context, req *v1.GetSiteWithdrawsReq) (*v1.GetSiteWithdrawsRes, error) {
	ctx, span := tracing.StartSpan(ctx, "balance.GetSiteWithdraws", trace.WithAttributes(
		attribute.String("method", "GetSiteWithdraws"),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	// 构建查询
	model := dao.SiteWithdraw.Ctx(ctx)

	// 条件筛选
	if req.WithdrawId > 0 {
		model = model.Where("withdraw_id", req.WithdrawId)
	}
	if req.Status >= 0 {
		model = model.Where("status", req.Status)
	}

	// 查询总数
	count, err := model.Count()
	if err != nil {
		g.Log().Errorf(ctx, "查询站点提现总数失败: %v", err)
		return nil, fmt.Errorf("查询站点提现总数失败")
	}

	// 查询列表
	var siteWithdraws []*entity.SiteWithdraw
	err = model.Page(int(page), int(size)).
		Order("sort ASC, id DESC").
		Scan(&siteWithdraws)
	if err != nil {
		g.Log().Errorf(ctx, "查询站点提现列表失败: %v", err)
		return nil, fmt.Errorf("查询站点提现列表失败")
	}

	// 转换为响应格式
	list := make([]*v1.SiteWithdrawInfo, 0, len(siteWithdraws))
	for _, sw := range siteWithdraws {
		list = append(list, &v1.SiteWithdrawInfo{
			Id:         int32(sw.Id),
			SiteId:     int32(sw.SiteId),
			WithdrawId: int32(sw.WithdrawId),
			Name:       sw.Name,
			Code:       sw.Code,
			Gateway:    int32(sw.Gateway),
			ApiUrl:     sw.ApiUrl,
			MerchantNo: sw.MerchantNo,
			Md5Key:     sw.Md5Key,
			EachMin:    sw.EachMin,
			EachMax:    sw.EachMax,
			Sort:       int32(sw.Sort),
			MoneyList:  sw.MoneyList,
			IsInput:    int32(sw.IsInput),
			Logo:       sw.Logo,
			Status:     int32(sw.Status),
			CreatedAt:  sw.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  sw.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &v1.GetSiteWithdrawsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// UpdateSiteWithdraw 更新站点提现
func (s *sBalance) UpdateSiteWithdraw(ctx context.Context, req *v1.UpdateSiteWithdrawReq) (*v1.UpdateSiteWithdrawRes, error) {
	ctx, span := tracing.StartSpan(ctx, "balance.UpdateSiteWithdraw", trace.WithAttributes(
		attribute.String("method", "UpdateSiteWithdraw"),
		attribute.Int("id", int(req.Id)),
	))
	defer span.End()

	if req.Id <= 0 {
		return &v1.UpdateSiteWithdrawRes{
			Success: false,
			Message: "站点提现ID不能为空",
		}, nil
	}

	// 构建更新数据
	data := g.Map{
		"updated_at": time.Now(),
	}

	if req.Status >= 0 {
		data["status"] = req.Status
	}
	if req.Sort >= 0 {
		data["sort"] = req.Sort
	}
	if req.EachMin >= 0 {
		data["each_min"] = req.EachMin
	}
	if req.EachMax >= 0 {
		data["each_max"] = req.EachMax
	}
	if req.MoneyList != "" {
		data["moneyList"] = req.MoneyList
	}
	if req.IsInput >= 0 {
		data["is_input"] = req.IsInput
	}
	if req.Remark != "" {
		data["remark"] = req.Remark
	}
	if req.Logo != "" {
		data["logo"] = req.Logo
	}

	// 更新数据
	_, err := dao.SiteWithdraw.Ctx(ctx).
		Where("id", req.Id).
		Data(data).
		Update()
	if err != nil {
		g.Log().Errorf(ctx, "更新站点提现失败: %v", err)
		return &v1.UpdateSiteWithdrawRes{
			Success: false,
			Message: "更新站点提现失败",
		}, nil
	}

	return &v1.UpdateSiteWithdrawRes{
		Success: true,
		Message: "更新成功",
	}, nil
}
