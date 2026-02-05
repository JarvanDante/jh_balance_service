// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceInviteReward is the golang structure for table balance_invite_reward.
type BalanceInviteReward struct {
	Id          uint64      `json:"id"          orm:"id"           description:""`
	SiteId      int         `json:"siteId"      orm:"site_id"      description:""`
	InviterId   int         `json:"inviterId"   orm:"inviter_id"   description:""`
	InviteeId   int         `json:"inviteeId"   orm:"invitee_id"   description:""`
	BonusAmount float64     `json:"bonusAmount" orm:"bonus_amount" description:""`
	TradeNo     string      `json:"tradeNo"     orm:"trade_no"     description:""`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`
}
