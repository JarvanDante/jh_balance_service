// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowRule is the golang structure of table balance_flow_rule for DAO operations like Where/Data.
type BalanceFlowRule struct {
	g.Meta       `orm:"table:balance_flow_rule, do:true"`
	Id           any         // 主键ID
	SiteId       any         // 站点ID
	RuleName     any         // 规则名称
	RuleType     any         // 规则类型: 1=充值流水, 2=红利流水, 3=活动流水, 4=人工加款流水
	FlowMultiple any         // 默认流水倍数
	ExpireDays   any         // 有效天数（0表示永久）
	Status       any         // 状态: 0=禁用, 1=启用
	Sort         any         // 排序
	Remark       any         // 备注说明
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
}
