// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogRebate is the golang structure for table balance_log_rebate.
type BalanceLogRebate struct {
	Id             uint64      `json:"id"             orm:"id"               description:""`
	SiteId         int         `json:"siteId"         orm:"site_id"          description:"站点ID"`
	UserId         int         `json:"userId"         orm:"user_id"          description:"用户ID"`
	Username       string      `json:"username"       orm:"username"         description:"会员用户名"`
	TradeNo        string      `json:"tradeNo"        orm:"trade_no"         description:"流水号"`
	LastBetId      uint64      `json:"lastBetId"      orm:"last_bet_id"      description:"注单记录游标ID"`
	ValidBetAmount float64     `json:"validBetAmount" orm:"valid_bet_amount" description:"有效投注总金额"`
	Money          float64     `json:"money"          orm:"money"            description:"返水金额"`
	Fee            float64     `json:"fee"            orm:"fee"              description:"手续费"`
	Status         int         `json:"status"         orm:"status"           description:"状态0=初始化，1=待处理；2=处理中；3=成功；4=失败"`
	AdminId        int         `json:"adminId"        orm:"admin_id"         description:"后台管理员ID"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"       description:""`
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"       description:""`
}
