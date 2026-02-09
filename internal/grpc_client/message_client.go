package grpc_client

import (
	"context"
	"fmt"

	messagev1 "jh_balance_service/api/grpc/message/backend/v1"
	"jh_balance_service/internal/registry"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GetMessageClient 获取 Message 服务的 gRPC 客户端
func GetMessageClient(ctx context.Context) (messagev1.MessageClient, *grpc.ClientConn, error) {
	addr, err := registry.GetServiceAddr("jh_app_service")
	if err != nil {
		return nil, nil, fmt.Errorf("从Consul获取app服务地址失败: %v", err)
	}

	g.Log().Infof(ctx, "app服务地址: %s", addr)

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("无法连接到app服务: %v", err)
	}

	return messagev1.NewMessageClient(conn), conn, nil
}

// SendUserMessage 调用 app_service 发送用户消息
func SendUserMessage(ctx context.Context, req *messagev1.SendUserMessageReq) (*messagev1.SendUserMessageRes, error) {
	client, conn, err := GetMessageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return client.SendUserMessage(ctx, req)
}
