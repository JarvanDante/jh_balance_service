// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusRecharge is the golang structure for table balance_log_bonus_recharge.
type BalanceLogBonusRecharge struct {
	Id            uint64      `json:"id"            orm:"id"              description:""`
	SiteId        int         `json:"siteId"        orm:"site_id"         description:"站点ID"`
	UserId        int         `json:"userId"        orm:"user_id"         description:"会员ID"`
	TradeType     uint        `json:"tradeType"     orm:"trade_type"      description:"类型。1=在线支付；2=转账汇款"`
	TradeNo       string      `json:"tradeNo"       orm:"trade_no"        description:"流水号"`
	PaymentLogId  int64       `json:"paymentLogId"  orm:"payment_log_id"  description:"在线充值日志ID"`
	TransferLogId int64       `json:"transferLogId" orm:"transfer_log_id" description:"转账汇款日志ID"`
	ActivityId    int         `json:"activityId"    orm:"activity_id"     description:"优惠活动ID"`
	Money         float64     `json:"money"         orm:"money"           description:"金额"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"      description:""`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"      description:""`
}
