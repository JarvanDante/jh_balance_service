// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogGame is the golang structure of table balance_log_game for DAO operations like Where/Data.
type BalanceLogGame struct {
	g.Meta            `orm:"table:balance_log_game, do:true"`
	Id                any         //
	SiteId            any         // 站点ID
	UserId            any         // 会员ID
	GameId            any         // 游戏ID
	InOrOut           any         // 转入还是转出；0=转出；1=转入
	TradeNo           any         //
	Money             any         // 操作金额
	Status            any         // 状态. 1=待确认，2=成功，3=失败
	Remark            any         // 备注
	ThirdPartyTradeNo any         // 第三方订单号
	CreatedAt         *gtime.Time //
	UpdatedAt         *gtime.Time //
}
