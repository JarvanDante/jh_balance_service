// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogTransfer is the golang structure for table balance_log_transfer.
type BalanceLogTransfer struct {
	Id                 uint64      `json:"id"                 orm:"id"                   description:""`
	SiteId             int         `json:"siteId"             orm:"site_id"              description:"站点ID"`
	UserId             int         `json:"userId"             orm:"user_id"              description:"用户ID"`
	Username           string      `json:"username"           orm:"username"             description:"会员用户名"`
	ActivityRechargeId int         `json:"activityRechargeId" orm:"activity_recharge_id" description:"充值活动ID"`
	BankType           int         `json:"bankType"           orm:"bank_type"            description:"转账汇款类型"`
	AccountId          int         `json:"accountId"          orm:"account_id"           description:"转账汇款账号ID"`
	TradeNo            string      `json:"tradeNo"            orm:"trade_no"             description:"流水号"`
	ThirdTradeNo       string      `json:"thirdTradeNo"       orm:"third_trade_no"       description:"第三方订单号"`
	Money              float64     `json:"money"              orm:"money"                description:"充值金额"`
	Fee                float64     `json:"fee"                orm:"fee"                  description:"手续费"`
	Status             int         `json:"status"             orm:"status"               description:"状态. 1=待审核；2=成功；3=失败"`
	TransferAccount    string      `json:"transferAccount"    orm:"transfer_account"     description:"汇款账户名"`
	Code               string      `json:"code"               orm:"code"                 description:"随机码。其实就是会员用户名"`
	AdminId            int         `json:"adminId"            orm:"admin_id"             description:"操作人id"`
	Remark             string      `json:"remark"             orm:"remark"               description:"备注"`
	CreatedAt          *gtime.Time `json:"createdAt"          orm:"created_at"           description:""`
	UpdatedAt          *gtime.Time `json:"updatedAt"          orm:"updated_at"           description:""`
}
