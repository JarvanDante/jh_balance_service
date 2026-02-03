// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowRule is the golang structure for table balance_flow_rule.
type BalanceFlowRule struct {
	Id           int         `json:"id"           orm:"id"            description:"主键ID"`
	SiteId       int         `json:"siteId"       orm:"site_id"       description:"站点ID"`
	RuleName     string      `json:"ruleName"     orm:"rule_name"     description:"规则名称"`
	RuleType     int         `json:"ruleType"     orm:"rule_type"     description:"规则类型"`
	FlowMultiple float64     `json:"flowMultiple" orm:"flow_multiple" description:"默认流水倍数"`
	ExpireDays   int         `json:"expireDays"   orm:"expire_days"   description:"有效天数"`
	Status       int         `json:"status"       orm:"status"        description:"状态"`
	Sort         int         `json:"sort"         orm:"sort"          description:"排序"`
	Remark       string      `json:"remark"       orm:"remark"        description:"备注说明"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`
}
