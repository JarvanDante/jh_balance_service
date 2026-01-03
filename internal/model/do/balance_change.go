// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceChange is the golang structure of table balance_change for DAO operations like Where/Data.
type BalanceChange struct {
	g.Meta        `orm:"table:balance_change, do:true"`
	Id            any         //
	SiteId        any         // 站点ID
	TradeType     any         // 交易类型。1=转入；2=充值；3=红利；4=返水；5=转出；6=提现成功；7=提现返回；8=提现冻结；9=丢失补款；10=多出扣款
	UserId        any         // 用户ID
	Username      any         // 用户名
	TradeNo       any         // 流水号
	BalanceOld    any         // 旧的可用余额
	Money         any         // 变动金额
	BalanceNew    any         // 新的可用余额
	BalanceFrozen any         // 新的冻结余额
	Status        any         // 账变状态。1=待审核；2=成功；3=失败
	Remark        any         //
	CreatedAt     *gtime.Time //
	UpdatedAt     *gtime.Time //
}
