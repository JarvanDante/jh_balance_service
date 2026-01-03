// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusAgent is the golang structure of table balance_log_bonus_agent for DAO operations like Where/Data.
type BalanceLogBonusAgent struct {
	g.Meta       `orm:"table:balance_log_bonus_agent, do:true"`
	Id           any         //
	SiteId       any         //
	UserId       any         //
	AgentId      any         //
	AgentLevelId any         //
	TradeNo      any         // 订单号
	Money        any         //
	CreatedAt    *gtime.Time //
	UpdatedAt    *gtime.Time //
}
