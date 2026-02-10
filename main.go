package main

import (
	"context"
	"fmt"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	_ "jh_balance_service/internal/logic"
	_ "jh_balance_service/internal/logic/backend/balance"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"

	"jh_balance_service/internal/cmd"
)

func main() {
	// 在最早的时刻设置时区
	ctx := gctx.New()
	timezone := g.Cfg().MustGet(ctx, "timezone").String()
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Printf("加载时区失败: %v，使用默认时区\n", err)
		loc, _ = time.LoadLocation("Asia/Shanghai")
	}
	time.Local = loc
	fmt.Printf("时区设置: %s\n", timezone)

	// 创建根命令
	command := &gcmd.Command{
		Name:  "jh_balance_service",
		Brief: "JH Balance Service",
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			// 默认执行main命令（启动服务）
			return cmd.Main.Func(ctx, parser)
		},
	}

	// 添加子命令
	command.AddCommand(&cmd.Main, cmd.MqConsumer, cmd.FlowConsumer, cmd.FlowRetry, cmd.VipUpgrade, cmd.OutboxRelay)

	// 运行命令
	command.Run(ctx)
}
