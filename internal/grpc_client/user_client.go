package grpc_client

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userv1 "jh_balance_service/api/grpc/user/v1"
	"jh_balance_service/internal/registry"
)

// GetUserClient 获取 User 服务的 gRPC 客户端
func GetUserClient(ctx context.Context) (userv1.UserClient, *grpc.ClientConn, error) {
	// 从 Consul 获取 app 服务地址
	addr, err := registry.GetServiceAddr("jh_app_service")
	if err != nil {
		return nil, nil, fmt.Errorf("从Consul获取app服务地址失败: %v", err)
	}

	g.Log().Infof(ctx, "app服务地址: %s", addr)

	// 创建 gRPC 连接
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("无法连接到app服务: %v", err)
	}

	return userv1.NewUserClient(conn), conn, nil
}
