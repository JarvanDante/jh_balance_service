// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserReportTotal is the golang structure for table user_report_total.
type UserReportTotal struct {
	Id                uint64      `json:"id"                orm:"id"                   description:"主键"`
	SiteId            int         `json:"siteId"            orm:"site_id"              description:"站点ID"`
	UserId            uint64      `json:"userId"            orm:"user_id"              description:"用户ID"`
	DepositAmount     float64     `json:"depositAmount"     orm:"deposit_amount"       description:"累计充值总额"`
	BetAmount         float64     `json:"betAmount"         orm:"bet_amount"           description:"累计有效投注总额"`
	LastDepositId     uint64      `json:"lastDepositId"     orm:"last_deposit_id"      description:"充值记录游标ID"`
	LastBetId         uint64      `json:"lastBetId"         orm:"last_bet_id"          description:"注单记录游标ID"`
	LastDepositTime   *gtime.Time `json:"lastDepositTime"   orm:"last_deposit_time"    description:"最近累计的充值时间"`
	LastBetSettleTime *gtime.Time `json:"lastBetSettleTime" orm:"last_bet_settle_time" description:"最近累计的注单结算时间"`
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"           description:"创建时间"`
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"           description:"更新时间"`
}
