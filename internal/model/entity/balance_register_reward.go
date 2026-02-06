// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceRegisterReward is the golang structure for table balance_register_reward.
type BalanceRegisterReward struct {
	Id          uint64      `json:"id"          orm:"id"           description:""`
	SiteId      int         `json:"siteId"      orm:"site_id"      description:""`
	UserId      int         `json:"userId"      orm:"user_id"      description:""`
	BonusAmount float64     `json:"bonusAmount" orm:"bonus_amount" description:""`
	TradeNo     string      `json:"tradeNo"     orm:"trade_no"     description:"发放流水号/订单号"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`
}
