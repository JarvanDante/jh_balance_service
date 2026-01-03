// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusAgent is the golang structure for table balance_log_bonus_agent.
type BalanceLogBonusAgent struct {
	Id           int         `json:"id"           orm:"id"             description:""`
	SiteId       int         `json:"siteId"       orm:"site_id"        description:""`
	UserId       int         `json:"userId"       orm:"user_id"        description:""`
	AgentId      int         `json:"agentId"      orm:"agent_id"       description:""`
	AgentLevelId int         `json:"agentLevelId" orm:"agent_level_id" description:""`
	TradeNo      string      `json:"tradeNo"      orm:"trade_no"       description:"订单号"`
	Money        float64     `json:"money"        orm:"money"          description:""`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"     description:""`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"     description:""`
}
