// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogWithdraw is the golang structure for table balance_log_withdraw.
type BalanceLogWithdraw struct {
	Id          uint64      `json:"id"          orm:"id"            description:""`
	SiteId      int         `json:"siteId"      orm:"site_id"       description:"站点ID"`
	UserId      int         `json:"userId"      orm:"user_id"       description:"会员ID"`
	UserLevelId int         `json:"userLevelId" orm:"user_level_id" description:"会员层级ID"`
	Username    string      `json:"username"    orm:"username"      description:"会员用户名"`
	TradeNo     string      `json:"tradeNo"     orm:"trade_no"      description:"流水号"`
	Money       float64     `json:"money"       orm:"money"         description:"充值金额"`
	Fee         float64     `json:"fee"         orm:"fee"           description:"手续费"`
	BankName    string      `json:"bankName"    orm:"bank_name"     description:"提现银行"`
	CardAccount string      `json:"cardAccount" orm:"card_account"  description:"银行户名"`
	CardNo      string      `json:"cardNo"      orm:"card_no"       description:"卡号"`
	DepositBank string      `json:"depositBank" orm:"deposit_bank"  description:"开户行"`
	Status      int         `json:"status"      orm:"status"        description:"状态0=初始化，1=待处理；2=处理中；3=成功；4=失败"`
	AdminId     int         `json:"adminId"     orm:"admin_id"      description:"后台用户ID"`
	Remark      string      `json:"remark"      orm:"remark"        description:"备注"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"    description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"    description:""`
}
