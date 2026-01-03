// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceUser is the golang structure of table balance_user for DAO operations like Where/Data.
type BalanceUser struct {
	g.Meta        `orm:"table:balance_user, do:true"`
	Id            any         //
	SiteId        any         // 站点ID
	UserId        any         // 会员ID
	Balance       any         // 可用金额
	BalanceFrozen any         // 冻结金额
	BalanceDebt   any         // 负债金额。防止账户余额出现负数的情况。例如：注单结算错误，需要扣掉会员账户金额。如果会员账户金额足够，则直接扣除会员账户金额；如果会员账户金额不足，则不扣除会员账户金额，直接存入负债金额。在下次会员有资金变动的时候，先来计算这笔金额
	Points        any         // 会员积分
	BalanceStatus any         // 资金状态。1=正常；0=锁定
	CreatedAt     *gtime.Time //
	UpdatedAt     *gtime.Time //
}
