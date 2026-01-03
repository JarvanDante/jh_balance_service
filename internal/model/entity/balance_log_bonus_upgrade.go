// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusUpgrade is the golang structure for table balance_log_bonus_upgrade.
type BalanceLogBonusUpgrade struct {
	Id        uint64      `json:"id"        orm:"id"         description:""`
	SiteId    int         `json:"siteId"    orm:"site_id"    description:"站点ID"`
	UserId    int         `json:"userId"    orm:"user_id"    description:"会员ID"`
	TradeNo   string      `json:"tradeNo"   orm:"trade_no"   description:"流水号"`
	GradeId   int         `json:"gradeId"   orm:"grade_id"   description:"会员等级ID"`
	Money     float64     `json:"money"     orm:"money"      description:"金额"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
}
