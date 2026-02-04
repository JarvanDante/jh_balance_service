package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"jh_balance_service/internal/registry"
)

// StartFlowConsumedCleanup 启动流水幂等表定期清理（保留1个月）
func StartFlowConsumedCleanup(ctx context.Context, dbManager *registry.DBConfigManager) {
	go func() {
		run := func() {
			if err := cleanupFlowConsumedBets(ctx, dbManager); err != nil {
				g.Log().Errorf(ctx, "流水幂等表清理失败: %v", err)
			}
		}

		// 启动后先跑一次
		run()

		// 每24小时清理一次
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}

func cleanupFlowConsumedBets(ctx context.Context, dbManager *registry.DBConfigManager) error {
	cutoff := time.Now().AddDate(0, -1, 0)
	siteCodes := dbManager.GetAllSiteCodes()

	const batchSize = 50000
	total := 0

	for _, siteCode := range siteCodes {
		groupName := "site_" + siteCode
		if gdb.GetConfig(groupName) == nil {
			continue
		}
		siteCtx := context.WithValue(ctx, "dynamic_db_group", groupName)

		// 表不存在则跳过
		val, err := g.DB(groupName).GetValue(siteCtx, "SHOW TABLES LIKE 'balance_flow_consumed_bets'")
		if err != nil {
			return err
		}
		if val == nil {
			continue
		}

		for {
			result, err := g.DB(groupName).Model("balance_flow_consumed_bets").
				Where("consumed_at < ?", cutoff).
				Limit(batchSize).
				Delete()
			if err != nil {
				return fmt.Errorf("site=%s 删除失败: %w", siteCode, err)
			}

			affected, _ := result.RowsAffected()
			if affected == 0 {
				break
			}
			total += int(affected)
			g.Log().Infof(ctx, "流水幂等表清理: site=%s, deleted=%d, total=%d", siteCode, affected, total)
			time.Sleep(200 * time.Millisecond)
		}
	}

	if total > 0 {
		g.Log().Infof(ctx, "流水幂等表清理完成: total=%d", total)
	}
	return nil
}
