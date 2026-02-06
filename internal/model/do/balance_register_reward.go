// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceRegisterReward is the golang structure of table balance_register_reward for DAO operations like Where/Data.
type BalanceRegisterReward struct {
	g.Meta      `orm:"table:balance_register_reward, do:true"`
	Id          any         //
	SiteId      any         //
	UserId      any         //
	BonusAmount any         //
	TradeNo     any         // 发放流水号/订单号
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
