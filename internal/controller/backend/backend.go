package backend

import (
	"context"
	v1 "jh_balance_service/api/grpc/game/backend/v1"

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

func (*Controller) GetGames(ctx context.Context, req *v1.GetGamesReq) (res *v1.GetGamesRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) UpdateGames(ctx context.Context, req *v1.UpdateGamesReq) (res *v1.UpdateGamesRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetGameRecords(ctx context.Context, req *v1.GetGameRecordsReq) (res *v1.GetGameRecordsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) SupplementBetRecords(ctx context.Context, req *v1.SupplementBetRecordsReq) (res *v1.SupplementBetRecordsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetGameList(ctx context.Context, req *v1.GetGameListReq) (res *v1.GetGameListRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) UpdateGameListItem(ctx context.Context, req *v1.UpdateGameListItemReq) (res *v1.UpdateGameListItemRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetGamePlatforms(ctx context.Context, req *v1.GetGamePlatformsReq) (res *v1.GetGamePlatformsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetUserBetStats(ctx context.Context, req *v1.GetUserBetStatsReq) (res *v1.GetUserBetStatsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetUserBetSummary(ctx context.Context, req *v1.GetUserBetSummaryReq) (res *v1.GetUserBetSummaryRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

func (*Controller) GetSyncTasks(ctx context.Context, req *v1.GetSyncTasksReq) (res *v1.GetSyncTasksRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
