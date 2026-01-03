// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceRebateHistory is the golang structure for table balance_rebate_history.
type BalanceRebateHistory struct {
	Id             uint        `json:"id"             orm:"id"               description:""`
	SiteId         uint        `json:"siteId"         orm:"site_id"          description:"站点ID"`
	RebateDate     *gtime.Time `json:"rebateDate"     orm:"rebate_date"      description:"返水日期"`
	UserCount      uint        `json:"userCount"      orm:"user_count"       description:"返水人数"`
	ValidBetAmount float64     `json:"validBetAmount" orm:"valid_bet_amount" description:"有效投注金额"`
	Money          float64     `json:"money"          orm:"money"            description:"返水金额"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"       description:""`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"       description:""`
}
