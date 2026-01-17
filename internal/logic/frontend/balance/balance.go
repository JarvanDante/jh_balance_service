package balance

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "jh_balance_service/api/frontend/balance/v1"
	"jh_balance_service/internal/dao"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/service/frontend"
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
