// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowConsumeRetry is the golang structure for table balance_flow_consume_retry.
type BalanceFlowConsumeRetry struct {
	Id          uint64      `json:"id"          orm:"id"           description:""`
	SiteId      int         `json:"siteId"      orm:"site_id"      description:""`
	UserId      int64       `json:"userId"      orm:"user_id"      description:""`
	OrderNo     string      `json:"orderNo"     orm:"order_no"     description:""`
	ValidAmount float64     `json:"validAmount" orm:"valid_amount" description:""`
	ErrorMsg    string      `json:"errorMsg"    orm:"error_msg"    description:""`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
}
