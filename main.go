package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	_ "jh_balance_service/internal/logic"
	_ "jh_balance_service/internal/logic/backend/balance"

	"github.com/gogf/gf/v2/os/gctx"

	"jh_balance_service/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.New())
}
