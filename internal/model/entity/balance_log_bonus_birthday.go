// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusBirthday is the golang structure for table balance_log_bonus_birthday.
type BalanceLogBonusBirthday struct {
	Id        uint64      `json:"id"        orm:"id"         description:""`
	SiteId    int         `json:"siteId"    orm:"site_id"    description:"站点ID"`
	UserId    int         `json:"userId"    orm:"user_id"    description:"会员ID"`
	TradeNo   string      `json:"tradeNo"   orm:"trade_no"   description:"流水号"`
	Year      int         `json:"year"      orm:"year"       description:"申请的年份。每年只能申请一次"`
	Money     float64     `json:"money"     orm:"money"      description:"金额"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
}
