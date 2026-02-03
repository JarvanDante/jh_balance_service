package grpc_client

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gamev1 "jh_balance_service/api/grpc/game/frontend/v1"
	"jh_balance_service/internal/registry"
)

// GetGameClient 获取 Game 服务的 gRPC 客户端
func GetGameClient(ctx context.Context) (gamev1.GameClient, *grpc.ClientConn, error) {
	// 从 Consul 获取 game 服务地址
	addr, err := registry.GetServiceAddr("jh_game_service")
	if err != nil {
		return nil, nil, fmt.Errorf("从Consul获取game服务地址失败: %v", err)
	}

	g.Log().Infof(ctx, "game服务地址: %s", addr)

	// 创建 gRPC 连接
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("无法连接到game服务: %v", err)
	}

	return gamev1.NewGameClient(conn), conn, nil
}
