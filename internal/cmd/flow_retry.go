package cmd

import (
	"context"
	"fmt"
	"jh_balance_service/internal/mq"
	"jh_balance_service/internal/registry"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	FlowRetry = &gcmd.Command{
		Name:  "flow-retry",
		Usage: "flow-retry",
		Brief: "重试流水消费失败记录",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "Flow Retry 启动中...")

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

			siteCodes := dbManager.GetAllSiteCodes()
			total := 0
			for _, siteCode := range siteCodes {
				count, err := retrySite(ctx, siteCode, 200)
				if err != nil {
					g.Log().Errorf(ctx, "重试失败: site=%s, err=%v", siteCode, err)
					continue
				}
				if count > 0 {
					g.Log().Infof(ctx, "重试完成: site=%s, count=%d", siteCode, count)
				}
				total += count
			}

			g.Log().Infof(ctx, "Flow Retry 完成，总重试=%d", total)
			return nil
		},
	}
)

type retryRow struct {
	Id          uint64  `json:"id"`
	SiteId      int     `json:"site_id"`
	UserId      int64   `json:"user_id"`
	OrderNo     string  `json:"order_no"`
	ValidAmount float64 `json:"valid_amount"`
}

func retrySite(ctx context.Context, siteCode string, limit int) (int, error) {
	groupName := "site_" + siteCode
	if gdb.GetConfig(groupName) == nil {
		return 0, fmt.Errorf("动态数据库组不存在: %s", groupName)
	}

	siteCtx := context.WithValue(ctx, "dynamic_db_group", groupName)

	// 表不存在时跳过（部分站点可能未建表）
	val, err := g.DB(groupName).GetValue(siteCtx, "SHOW TABLES LIKE 'balance_flow_consume_retry'")
	if err != nil {
		return 0, err
	}
	if val == nil {
		g.Log().Warningf(ctx, "重试表不存在，跳过: site=%s", siteCode)
		return 0, nil
	}

	var rows []retryRow
	if err := g.DB(groupName).Model("balance_flow_consume_retry").
		Order("id ASC").
		Limit(limit).
		Scan(&rows); err != nil {
		return 0, err
	}

	success := 0
	for _, row := range rows {
		err := g.DB(groupName).Transaction(siteCtx, func(ctx context.Context, tx gdb.TX) error {
			return mq.ApplyFlowRequirementTx(ctx, tx, row.SiteId, row.UserId, row.ValidAmount)
		})
		if err != nil {
			_, _ = g.DB(groupName).Model("balance_flow_consume_retry").
				Where("id", row.Id).
				Data(g.Map{"error_msg": fmt.Sprintf("%s | %s", err.Error(), time.Now().Format("2006-01-02 15:04:05"))}).
				Update()
			continue
		}

		_, _ = g.DB(groupName).Model("balance_flow_consume_retry").
			Where("id", row.Id).
			Delete()
		success++
	}

	return success, nil
}
