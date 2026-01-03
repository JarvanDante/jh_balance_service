package cmd

import (
	"context"
	"jh_balance_service/internal/controller/backend/balance"
	"jh_balance_service/internal/registry"
	"jh_balance_service/internal/tracing"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"google.golang.org/grpc"
)

var (
	// Main is the main command.
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start grpc server of balance service",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化Jaeger追踪
			cleanup, err := tracing.InitJaeger()
			if err != nil {
				g.Log().Fatalf(ctx, "Failed to initialize Jaeger: %v", err)
			}
			defer cleanup()

			// 初始化服务注册
			err = registry.InitConsul()
			if err != nil {
				g.Log().Fatalf(ctx, "Failed to initialize Consul: %v", err)
			}

			c := grpcx.Server.NewConfig()
			c.Options = append(c.Options, []grpc.ServerOption{
				grpcx.Server.ChainUnary(
					grpcx.Server.UnaryValidate,
				)}...,
			)
			s := grpcx.Server.New(c)

			// 注册服务
			balance.Register(s)

			s.Run()
			return nil
		},
	}
)
