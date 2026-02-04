package cmd

import (
	"context"
	"jh_balance_service/internal/mq"
	"jh_balance_service/internal/registry"
	"jh_balance_service/internal/tracing"
	"os"
	"os/signal"
	"syscall"

	// 导入logic包以触发init函数
	_ "jh_balance_service/internal/logic/frontend/balance"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	FlowConsumer = &gcmd.Command{
		Name:  "flow-consumer",
		Usage: "flow-consumer",
		Brief: "启动流水消费消费者，监听注单结算消息并更新流水要求",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "Flow Consumer 启动中...")

			// 初始化Jaeger追踪
			cleanup, err := tracing.InitJaeger()
			if err != nil {
				g.Log().Fatalf(ctx, "Failed to initialize Jaeger: %v", err)
			}
			defer cleanup()

			// 初始化Consul客户端
			consulClient, err := registry.InitConsul()
			if err != nil {
				g.Log().Fatalf(ctx, "Failed to initialize Consul: %v", err)
			}

			// 初始化数据库配置管理器
			dbManager := registry.GetDBConfigManager()
			if err := dbManager.Init(consulClient); err != nil {
				g.Log().Fatalf(ctx, "Failed to init balance db config manager: %v", err)
			}
			defer dbManager.Stop()

			// 初始化 RabbitMQ
			if err := mq.InitRabbitMQ(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to initialize RabbitMQ: %v", err)
			}
			defer mq.CloseRabbitMQ()

			// 启动消费者
			if err := mq.StartFlowConsumeConsumer(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to start flow consumer: %v", err)
			}

			// 启动流水幂等表定期清理（保留1个月）
			StartFlowConsumedCleanup(ctx, dbManager)

			g.Log().Info(ctx, "Flow Consumer 启动成功，等待消息...")

			// 等待退出信号
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			g.Log().Info(ctx, "Flow Consumer 停止")
			return nil
		},
	}
)
