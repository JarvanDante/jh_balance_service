// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowConsumedBets is the golang structure of table balance_flow_consumed_bets for DAO operations like Where/Data.
type BalanceFlowConsumedBets struct {
	g.Meta     `orm:"table:balance_flow_consumed_bets, do:true"`
	Id         any         //
	SiteId     any         //
	UserId     any         //
	OrderNo    any         //
	ConsumedAt *gtime.Time //
}
