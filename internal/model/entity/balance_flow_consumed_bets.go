// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowConsumedBets is the golang structure for table balance_flow_consumed_bets.
type BalanceFlowConsumedBets struct {
	Id         uint64      `json:"id"         orm:"id"          description:""`
	SiteId     int         `json:"siteId"     orm:"site_id"     description:""`
	UserId     int64       `json:"userId"     orm:"user_id"     description:""`
	OrderNo    string      `json:"orderNo"    orm:"order_no"    description:""`
	ConsumedAt *gtime.Time `json:"consumedAt" orm:"consumed_at" description:""`
}
