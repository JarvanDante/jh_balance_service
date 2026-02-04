// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowConsumeRetry is the golang structure of table balance_flow_consume_retry for DAO operations like Where/Data.
type BalanceFlowConsumeRetry struct {
	g.Meta      `orm:"table:balance_flow_consume_retry, do:true"`
	Id          any         //
	SiteId      any         //
	UserId      any         //
	OrderNo     any         //
	ValidAmount any         //
	ErrorMsg    any         //
	CreatedAt   *gtime.Time //
}
