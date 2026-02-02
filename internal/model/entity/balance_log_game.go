// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogGame is the golang structure for table balance_log_game.
type BalanceLogGame struct {
	Id                uint        `json:"id"                orm:"id"                   description:""`
	SiteId            int         `json:"siteId"            orm:"site_id"              description:"站点ID"`
	UserId            int         `json:"userId"            orm:"user_id"              description:"会员ID"`
	GameId            int         `json:"gameId"            orm:"game_id"              description:"游戏ID"`
	InOrOut           int         `json:"inOrOut"           orm:"in_or_out"            description:"转入还是转出；0=转出；1=转入"`
	TradeNo           string      `json:"tradeNo"           orm:"trade_no"             description:""`
	Money             float64     `json:"money"             orm:"money"                description:"操作金额"`
	Status            int         `json:"status"            orm:"status"               description:"状态0=初始化，1=待处理；2=处理中；3=成功；4=失败"`
	Remark            string      `json:"remark"            orm:"remark"               description:"备注"`
	ThirdPartyTradeNo string      `json:"thirdPartyTradeNo" orm:"third_party_trade_no" description:"第三方订单号"`
	Key               string      `json:"key"               orm:"key"                  description:"第三方游戏key"`
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"           description:""`
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"           description:""`
}
