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
	BonusAmount any         //
	TradeNo     any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
