package balance

import (
	"context"
	"fmt"
	v1 "jh_balance_service/api/backend/balance/v1"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/service/backend"
	"jh_balance_service/internal/tracing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	// 开启事务
	err = dao.BalanceLogPayment.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 更新订单状态
		_, err := dao.BalanceLogPayment.Ctx(ctx).TX(tx).
			Where("id = ?", req.Id).
			Update(g.Map{
				"status":     balanceConsts.StatusSuccess,
				"remark":     req.Remark,
				"updated_at": "NOW()",
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

	// 查找提现记录
	var withdraw entity.BalanceLogWithdraw
	err := dao.BalanceLogWithdraw.Ctx(ctx).Where("id = ?", req.Id).Scan(&withdraw)
	if err != nil {
		return nil, fmt.Errorf("查找提现记录失败: %v", err)
	}

	if withdraw.Id == 0 {
		return nil, fmt.Errorf("提现记录不存在")
	}

	// TODO: 调用用户服务获取用户信息和余额
	// 这里需要调用 jh_balance_service 的用户信息接口

	reviewInfo := &v1.WithdrawReviewInfo{
		Id:                int64(withdraw.Id),
		UserId:            int32(withdraw.UserId),
		Username:          withdraw.Username,
		TradeNo:           withdraw.TradeNo,
		Money:             withdraw.Money,
		Fee:               withdraw.Fee,
		BankName:          withdraw.BankName,
		CardAccount:       withdraw.CardAccount,
		CardNo:            withdraw.CardNo, // 审核时显示完整卡号
		DepositBank:       withdraw.DepositBank,
		Status:            int32(withdraw.Status),
		Remark:            withdraw.Remark,
		CreatedAt:         withdraw.CreatedAt.String(),
		UserRealname:      "",
		UserMobile:        "",
		UserBalance:       0,
		UserBalanceFrozen: 0,
		TotalRecharge:     0,
		TotalWithdraw:     0,
		WithdrawCount:     0,
	}

	return &v1.GetWithdrawReviewRes{
		Data: reviewInfo,
	}, nil
}

// DealWithWithdraw 处理提现
func (s *sBalance) DealWithWithdraw(ctx context.Context, req *v1.DealWithWithdrawReq) (*v1.DealWithWithdrawRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

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

	var message string
	// 开启事务处理
	err = dao.BalanceLogWithdraw.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var newStatus int

		switch req.Type {
		case 1: // 确认提现
			newStatus = balanceConsts.StatusSuccess
			message = "提现确认成功"
			// TODO: 调用第三方支付API
		case 0: // 拒绝提现
			newStatus = balanceConsts.StatusFail
			message = "提现已拒绝"
			// TODO: 调用用户服务解冻余额
		case 2: // 补单提现
			newStatus = balanceConsts.StatusSuccess
			message = "补单提现成功"
			// TODO: 调用第三方支付API
		default:
			return fmt.Errorf("无效的处理类型")
		}

		// 更新提现记录
		_, err := dao.BalanceLogWithdraw.Ctx(ctx).TX(tx).
			Where("id = ?", req.Id).
			Update(g.Map{
				"status":     newStatus,
				"fee":        req.Fee,
				"remark":     req.Remark,
				"updated_at": "NOW()",
			})
		if err != nil {
			return fmt.Errorf("更新提现记录失败: %v", err)
		}

		return nil
	})

	if err != nil {
		return &v1.DealWithWithdrawRes{
			Success: false,
			Message: fmt.Sprintf("处理提现失败: %v", err),
		}, nil
	}

	return &v1.DealWithWithdrawRes{
		Success: true,
		Message: message,
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
		balanceConsts.StatusPending: "待审核",
		balanceConsts.StatusSuccess: "成功",
		balanceConsts.StatusFail:    "失败",
		4:                           "处理中",
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
	if req.Type != 1 && req.Type != 2 {
		return nil, fmt.Errorf("操作类型无效")
	}
	if req.Money <= 0 {
		return nil, fmt.Errorf("操作金额必须大于0")
	}

	siteId := 1

	// 开启事务
	err := dao.BalanceUser.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
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

		if req.Type == 1 { // 加款
			balanceNew = balanceOld + req.Money
			tradeType = balanceConsts.TradeTypeManualAdd // 丢失补款
			remark = "手动加款"
		} else { // 扣款
			if balanceOld < req.Money {
				return fmt.Errorf("余额不足")
			}
			balanceNew = balanceOld - req.Money
			tradeType = balanceConsts.TradeTypeManualCut // 多出扣款
			remark = "手动扣款"
		}

		if req.Remark != "" {
			remark = req.Remark
		}

		// 更新用户余额
		_, err = dao.BalanceUser.Ctx(ctx).TX(tx).
			Where("id = ?", balance.Id).
			Update(g.Map{
				"balance":    balanceNew,
				"updated_at": "NOW()",
			})
		if err != nil {
			return fmt.Errorf("更新用户余额失败: %v", err)
		}

		// 生成流水号
		tradeNo := fmt.Sprintf("M%d%d", req.UserId, time.Now().Unix())

		// 记录账变
		_, err = dao.BalanceChange.Ctx(ctx).TX(tx).Insert(g.Map{
			"site_id":        siteId,
			"trade_type":     tradeType,
			"user_id":        req.UserId,
			"username":       "testuser", // TODO: 从用户服务获取
			"trade_no":       tradeNo,
			"balance_old":    balanceOld,
			"money":          req.Money,
			"balance_new":    balanceNew,
			"balance_frozen": balance.BalanceFrozen,
			"status":         balanceConsts.StatusSuccess,
			"remark":         remark,
			"created_at":     "NOW()",
			"updated_at":     "NOW()",
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
			"created_at": "NOW()",
			"updated_at": "NOW()",
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

// GetPaymentAccounts 获取支付接口列表
func (s *sBalance) GetPaymentAccounts(ctx context.Context, req *v1.GetPaymentAccountsReq) (*v1.GetPaymentAccountsRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetPaymentAccounts", trace.WithAttributes(
		attribute.String("method", "GetPaymentAccounts"),
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

	query := dao.PaymentAccount.Ctx(ctx).Where("site_id = ?", siteId)

	if req.PaymentId > 0 {
		query = query.Where("payment_id = ?", req.PaymentId)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}

	count, err := query.Count()
	if err != nil {
		return nil, fmt.Errorf("查询支付接口总数失败: %v", err)
	}

	var accounts []*entity.PaymentAccount
	offset := (page - 1) * size
	err = query.Order("sort ASC, created_at DESC").
		Limit(int(size)).
		Offset(int(offset)).
		Scan(&accounts)
	if err != nil {
		return nil, fmt.Errorf("查询支付接口列表失败: %v", err)
	}

	var list []*v1.PaymentAccountInfo
	for _, account := range accounts {
		info := &v1.PaymentAccountInfo{
			Id:          int32(account.Id),
			SiteId:      int32(account.SiteId),
			PaymentId:   int32(account.PaymentId),
			Gateway:     int32(account.Gateway),
			Name:        account.Name,
			Domain:      account.Domain,
			MerchantNo:  account.MerchantNo,
			Md5Key:      s.maskSensitiveData(account.Md5Key),
			EachMin:     account.EachMin,
			EachMax:     account.EachMax,
			DailyMax:    account.DailyMax,
			TodayCount:  int32(account.TodayCount),
			TodayAmount: account.TodayAmount,
			Status:      int32(account.Status),
			StatusName:  s.getPaymentAccountStatusName(account.Status),
			Sort:        int32(account.Sort),
			CreatedAt:   account.CreatedAt.String(),
			UpdatedAt:   account.UpdatedAt.String(),
			PublicKey:   s.maskSensitiveData(account.PublicKey),
			PrivateKey:  s.maskSensitiveData(account.PrivateKey),
			IsDecimal:   int32(account.IsDecimal),
			IsInt:       int32(account.IsInt),
			MoneyList:   account.MoneyList,
		}
		list = append(list, info)
	}

	return &v1.GetPaymentAccountsRes{
		List:  list,
		Count: int32(count),
	}, nil
}

// CreatePaymentAccount 创建支付接口
func (s *sBalance) CreatePaymentAccount(ctx context.Context, req *v1.CreatePaymentAccountReq) (*v1.CreatePaymentAccountRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.CreatePaymentAccount", trace.WithAttributes(
		attribute.String("method", "CreatePaymentAccount"),
		attribute.Int("payment_id", int(req.PaymentId)),
	))
	defer span.End()

	if req.PaymentId <= 0 {
		return nil, fmt.Errorf("支付ID不能为空")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("接口名称不能为空")
	}
	if req.MerchantNo == "" {
		return nil, fmt.Errorf("商户号不能为空")
	}

	siteId := 1

	// 检查账户名称是否重复
	count, err := dao.PaymentAccount.Ctx(ctx).Where("site_id = ? AND name = ?", siteId, req.Name).Count()
	if err != nil {
		return nil, fmt.Errorf("检查接口名称失败: %v", err)
	}
	if count > 0 {
		return &v1.CreatePaymentAccountRes{
			Success: false,
			Message: "接口名称已存在",
		}, nil
	}

	// 插入数据
	_, err = dao.PaymentAccount.Ctx(ctx).Insert(g.Map{
		"site_id":      siteId,
		"payment_id":   req.PaymentId,
		"gateway":      req.Gateway,
		"name":         req.Name,
		"domain":       req.Domain,
		"merchant_no":  req.MerchantNo,
		"md5_key":      req.Md5Key,
		"each_min":     req.EachMin,
		"each_max":     req.EachMax,
		"daily_max":    req.DailyMax,
		"today_count":  0,
		"today_amount": 0,
		"status":       req.Status,
		"sort":         req.Sort,
		"public_key":   req.PublicKey,
		"private_key":  req.PrivateKey,
		"is_decimal":   req.IsDecimal,
		"is_int":       req.IsInt,
		"money_list":   req.MoneyList,
		"created_at":   "NOW()",
		"updated_at":   "NOW()",
	})
	if err != nil {
		return &v1.CreatePaymentAccountRes{
			Success: false,
			Message: fmt.Sprintf("创建支付接口失败: %v", err),
		}, nil
	}

	return &v1.CreatePaymentAccountRes{
		Success: true,
		Message: "创建成功",
	}, nil
}

// GetPaymentAccountUpdate 获取支付接口编辑信息
func (s *sBalance) GetPaymentAccountUpdate(ctx context.Context, req *v1.GetPaymentAccountUpdateReq) (*v1.GetPaymentAccountUpdateRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.GetPaymentAccountUpdate", trace.WithAttributes(
		attribute.String("method", "GetPaymentAccountUpdate"),
		attribute.Int("id", int(req.Id)),
	))
	defer span.End()

	if req.Id <= 0 {
		return nil, fmt.Errorf("支付接口ID不能为空")
	}

	var account entity.PaymentAccount
	err := dao.PaymentAccount.Ctx(ctx).Where("id = ?", req.Id).Scan(&account)
	if err != nil {
		return nil, fmt.Errorf("查询支付接口失败: %v", err)
	}

	if account.Id == 0 {
		return nil, fmt.Errorf("支付接口不存在")
	}

	accountInfo := &v1.PaymentAccountInfo{
		Id:          int32(account.Id),
		SiteId:      int32(account.SiteId),
		PaymentId:   int32(account.PaymentId),
		Gateway:     int32(account.Gateway),
		Name:        account.Name,
		Domain:      account.Domain,
		MerchantNo:  account.MerchantNo,
		Md5Key:      account.Md5Key, // 编辑时显示完整密钥
		EachMin:     account.EachMin,
		EachMax:     account.EachMax,
		DailyMax:    account.DailyMax,
		TodayCount:  int32(account.TodayCount),
		TodayAmount: account.TodayAmount,
		Status:      int32(account.Status),
		StatusName:  s.getPaymentAccountStatusName(account.Status),
		Sort:        int32(account.Sort),
		CreatedAt:   account.CreatedAt.String(),
		UpdatedAt:   account.UpdatedAt.String(),
		PublicKey:   account.PublicKey,  // 编辑时显示完整公钥
		PrivateKey:  account.PrivateKey, // 编辑时显示完整私钥
		IsDecimal:   int32(account.IsDecimal),
		IsInt:       int32(account.IsInt),
		MoneyList:   account.MoneyList,
	}

	return &v1.GetPaymentAccountUpdateRes{
		Data: accountInfo,
	}, nil
}

// UpdatePaymentAccount 更新支付接口
func (s *sBalance) UpdatePaymentAccount(ctx context.Context, req *v1.UpdatePaymentAccountReq) (*v1.UpdatePaymentAccountRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.UpdatePaymentAccount", trace.WithAttributes(
		attribute.String("method", "UpdatePaymentAccount"),
		attribute.Int("id", int(req.Id)),
	))
	defer span.End()

	if req.Id <= 0 {
		return nil, fmt.Errorf("支付接口ID不能为空")
	}
	if req.PaymentId <= 0 {
		return nil, fmt.Errorf("支付ID不能为空")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("接口名称不能为空")
	}

	// 检查记录是否存在
	var account entity.PaymentAccount
	err := dao.PaymentAccount.Ctx(ctx).Where("id = ?", req.Id).Scan(&account)
	if err != nil {
		return nil, fmt.Errorf("查询支付接口失败: %v", err)
	}
	if account.Id == 0 {
		return &v1.UpdatePaymentAccountRes{
			Success: false,
			Message: "支付接口不存在",
		}, nil
	}

	// 检查接口名称是否重复（排除自己）
	count, err := dao.PaymentAccount.Ctx(ctx).Where("site_id = ? AND name = ? AND id != ?", account.SiteId, req.Name, req.Id).Count()
	if err != nil {
		return nil, fmt.Errorf("检查接口名称失败: %v", err)
	}
	if count > 0 {
		return &v1.UpdatePaymentAccountRes{
			Success: false,
			Message: "接口名称已存在",
		}, nil
	}

	// 更新数据
	_, err = dao.PaymentAccount.Ctx(ctx).Where("id = ?", req.Id).Update(g.Map{
		"payment_id":  req.PaymentId,
		"gateway":     req.Gateway,
		"name":        req.Name,
		"domain":      req.Domain,
		"merchant_no": req.MerchantNo,
		"md5_key":     req.Md5Key,
		"each_min":    req.EachMin,
		"each_max":    req.EachMax,
		"daily_max":   req.DailyMax,
		"status":      req.Status,
		"sort":        req.Sort,
		"public_key":  req.PublicKey,
		"private_key": req.PrivateKey,
		"is_decimal":  req.IsDecimal,
		"is_int":      req.IsInt,
		"money_list":  req.MoneyList,
		"updated_at":  "NOW()",
	})
	if err != nil {
		return &v1.UpdatePaymentAccountRes{
			Success: false,
			Message: fmt.Sprintf("更新支付接口失败: %v", err),
		}, nil
	}

	return &v1.UpdatePaymentAccountRes{
		Success: true,
		Message: "更新成功",
	}, nil
}

// DeletePaymentAccount 删除支付接口
func (s *sBalance) DeletePaymentAccount(ctx context.Context, req *v1.DeletePaymentAccountReq) (*v1.DeletePaymentAccountRes, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	ctx, span := tracing.StartSpan(ctx, "balance.DeletePaymentAccount", trace.WithAttributes(
		attribute.String("method", "DeletePaymentAccount"),
		attribute.Int("id", int(req.Id)),
	))
	defer span.End()

	if req.Id <= 0 {
		return nil, fmt.Errorf("支付接口ID不能为空")
	}

	// 检查记录是否存在
	var account entity.PaymentAccount
	err := dao.PaymentAccount.Ctx(ctx).Where("id = ?", req.Id).Scan(&account)
	if err != nil {
		return nil, fmt.Errorf("查询支付接口失败: %v", err)
	}
	if account.Id == 0 {
		return &v1.DeletePaymentAccountRes{
			Success: false,
			Message: "支付接口不存在",
		}, nil
	}

	// 检查是否有关联的支付记录
	count, err := dao.BalanceLogPayment.Ctx(ctx).Where("payment_account_id = ?", req.Id).Count()
	if err != nil {
		return nil, fmt.Errorf("检查关联记录失败: %v", err)
	}
	if count > 0 {
		return &v1.DeletePaymentAccountRes{
			Success: false,
			Message: "该支付接口下有支付记录，不能删除",
		}, nil
	}

	// 删除记录
	_, err = dao.PaymentAccount.Ctx(ctx).Where("id = ?", req.Id).Delete()
	if err != nil {
		return &v1.DeletePaymentAccountRes{
			Success: false,
			Message: fmt.Sprintf("删除支付接口失败: %v", err),
		}, nil
	}

	return &v1.DeletePaymentAccountRes{
		Success: true,
		Message: "删除成功",
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
