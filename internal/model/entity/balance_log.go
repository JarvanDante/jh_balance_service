// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLog is the golang structure for table balance_log.
type BalanceLog struct {
	Id          uint        `json:"id"          orm:"id"           description:""`
	SiteId      int         `json:"siteId"      orm:"site_id"      description:"站点ID"`
	UserId      int         `json:"userId"      orm:"user_id"      description:"会员ID"`
	TradeType   int         `json:"tradeType"   orm:"trade_type"   description:"交易类型"`
	Gateway     int         `json:"gateway"     orm:"gateway"      description:"支付相关：支付类型"`
	BankType    int         `json:"bankType"    orm:"bank_type"    description:"转账相关：银行类型"`
	GameId      int         `json:"gameId"      orm:"game_id"      description:"游戏相关：游戏ID"`
	TradeNo     string      `json:"tradeNo"     orm:"trade_no"     description:"交易流水号"`
	Money       float64     `json:"money"       orm:"money"        description:"金额"`
	ActualMoney float64     `json:"actualMoney" orm:"actual_money" description:"实际充值金额"`
	Fee         float64     `json:"fee"         orm:"fee"          description:"手续费"`
	Status      int         `json:"status"      orm:"status"       description:"状态。0=待处理，1=处理中；2=成功；3=失败"`
	AdminId     int         `json:"adminId"     orm:"admin_id"     description:"后台员工ID"`
	Remark      string      `json:"remark"      orm:"remark"       description:"备注"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`
}
