// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceWheelReward is the golang structure for table balance_wheel_reward.
type BalanceWheelReward struct {
	Id           uint64      `json:"id"           orm:"id"            description:""`
	SiteId       int         `json:"siteId"       orm:"site_id"       description:"站点ID"`
	UserId       int         `json:"userId"       orm:"user_id"       description:"用户ID"`
	Date         *gtime.Time `json:"date"         orm:"date"          description:"抽奖日期(yyyy-mm-dd)"`
	SegmentNo    uint        `json:"segmentNo"    orm:"segment_no"    description:"命中扇区(1~12)"`
	PrizeName    string      `json:"prizeName"    orm:"prize_name"    description:"奖项名称"`
	RewardAmount float64     `json:"rewardAmount" orm:"reward_amount" description:"中奖金额"`
	TradeNo      string      `json:"tradeNo"      orm:"trade_no"      description:"发放流水号/订单号"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:""`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:""`
}
