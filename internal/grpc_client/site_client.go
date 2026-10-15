package grpc_client

import (
	"context"
	"fmt"
	"strconv"

	sitev1 "jh_balance_service/api/grpc/site/backend/v1"
	"jh_balance_service/internal/registry"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// WithSiteContext 把站点 ID 写入 outgoing metadata，供 app_service 按租户切库。
func WithSiteContext(ctx context.Context, siteId int) context.Context {
	if siteId <= 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "site_id", strconv.Itoa(siteId))
}

// GetSiteClient 获取 Site 服务的 gRPC 客户端（由 app_service 提供）
func GetSiteClient(ctx context.Context) (sitev1.SiteClient, *grpc.ClientConn, error) {
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

	return sitev1.NewSiteClient(conn), conn, nil
}
