// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceInviteReward is the golang structure of table balance_invite_reward for DAO operations like Where/Data.
type BalanceInviteReward struct {
	g.Meta      `orm:"table:balance_invite_reward, do:true"`
	Id          any         //
	SiteId      any         //
	InviterId   any         //
	InviteeId   any         //
	Date        *gtime.Time // 邀请日期年月日
	BonusAmount any         // 发送金额
	TradeNo     any         // 最后触发订单号
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
