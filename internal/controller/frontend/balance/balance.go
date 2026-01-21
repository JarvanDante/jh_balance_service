package balance

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"

	v1 "jh_balance_service/api/frontend/balance/v1"
	"jh_balance_service/internal/service/frontend"
)

type Controller struct {
	v1.UnimplementedBalanceServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterBalanceServer(s.Server, &Controller{})
}

// InitBalanceUser 初始化用户余额
func (c *Controller) InitBalanceUser(ctx context.Context, req *v1.InitBalanceUserReq) (res *v1.InitBalanceUserRes, err error) {
	g.Log().Infof(ctx, "Frontend Balance Controller收到InitBalanceUser请求")
	return frontend.Balance().InitBalanceUser(ctx, req)
}

// RefreshBalance 刷新用户余额
func (c *Controller) RefreshBalance(ctx context.Context, req *v1.RefreshBalanceReq) (res *v1.RefreshBalanceRes, err error) {
	return frontend.Balance().RefreshBalance(ctx, req)
}

// AdjustBalance 调整用户余额
func (c *Controller) AdjustBalance(ctx context.Context, req *v1.AdjustBalanceReq) (res *v1.AdjustBalanceRes, err error) {
	g.Log().Infof(ctx, "========== Balance Controller收到AdjustBalance请求 ==========")
	g.Log().Infof(ctx, "请求参数: site_id=%d, user_id=%d, type=%d",
		req.SiteId, req.UserId, req.Type)

	res, err = frontend.Balance().AdjustBalance(ctx, req)

	g.Log().Infof(ctx, "Balance Controller处理完成: success=%v, err=%v",
		res != nil && res.Success, err)

	return res, err
}

// UpdateTransferStatus 更新转账状态
func (c *Controller) UpdateTransferStatus(ctx context.Context, req *v1.UpdateTransferStatusReq) (res *v1.UpdateTransferStatusRes, err error) {
	return frontend.Balance().UpdateTransferStatus(ctx, req)
}

// CompensateBalance 补偿余额
func (c *Controller) CompensateBalance(ctx context.Context, req *v1.CompensateBalanceReq) (res *v1.CompensateBalanceRes, err error) {
	return frontend.Balance().CompensateBalance(ctx, req)
}

// GetLastBalanceLogGame 获取最后一次进入游戏
func (c *Controller) GetLastBalanceLogGame(ctx context.Context, req *v1.GetLastBalanceLogGameReq) (res *v1.GetLastBalanceLogGameRes, err error) {
	return frontend.Balance().GetLastBalanceLogGame(ctx, req)
}
