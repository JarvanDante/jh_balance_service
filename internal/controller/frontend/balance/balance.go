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

//// AdjustBalance 调整用户余额
//func (c *Controller) AdjustBalance(ctx context.Context, req *v1.AdjustBalanceReq) (res *v1.AdjustBalanceRes, err error) {
//	return frontend.Balance().AdjustBalance(ctx, req)
//}
