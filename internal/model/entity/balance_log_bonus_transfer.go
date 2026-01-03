// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusTransfer is the golang structure for table balance_log_bonus_transfer.
type BalanceLogBonusTransfer struct {
	Id                 uint64      `json:"id"                 orm:"id"                   description:""`
	SiteId             int         `json:"siteId"             orm:"site_id"              description:"站点ID"`
	UserId             int         `json:"userId"             orm:"user_id"              description:"会员ID"`
	TradeNo            string      `json:"tradeNo"            orm:"trade_no"             description:"流水号"`
	TransferLogId      int64       `json:"transferLogId"      orm:"transfer_log_id"      description:"转账汇款日志ID"`
	ActivityRechargeId int         `json:"activityRechargeId" orm:"activity_recharge_id" description:"充值活动ID"`
	Money              float64     `json:"money"              orm:"money"                description:"金额"`
	CreatedAt          *gtime.Time `json:"createdAt"          orm:"created_at"           description:""`
	UpdatedAt          *gtime.Time `json:"updatedAt"          orm:"updated_at"           description:""`
}
