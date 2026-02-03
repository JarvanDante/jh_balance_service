package frontend

import (
	"context"
	v1 "jh_balance_service/api/grpc/game/frontend/v1"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct {
	v1.UnimplementedGameServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterGameServer(s.Server, &Controller{})
}

func (*Controller) RedirectToGame(ctx context.Context, req *v1.RedirectToGameReq) (res *v1.RedirectToGameRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetGameCategories(ctx context.Context, req *v1.GetGameCategoriesReq) (res *v1.GetGameCategoriesRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetGameList(ctx context.Context, req *v1.GetGameListReq) (res *v1.GetGameListRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetGameBalance(ctx context.Context, req *v1.GetGameBalanceReq) (res *v1.GetGameBalanceRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) RecallGameBalance(ctx context.Context, req *v1.RecallGameBalanceReq) (res *v1.RecallGameBalanceRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
