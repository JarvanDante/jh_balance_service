// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogWithdraw is the golang structure of table balance_log_withdraw for DAO operations like Where/Data.
type BalanceLogWithdraw struct {
	g.Meta      `orm:"table:balance_log_withdraw, do:true"`
	Id          any         //
	SiteId      any         // 站点ID
	UserId      any         // 会员ID
	UserLevelId any         // 会员层级ID
	Username    any         // 会员用户名
	TradeNo     any         // 流水号
	Money       any         // 充值金额
	Fee         any         // 手续费
	BankName    any         // 提现银行
	CardAccount any         // 银行户名
	CardNo      any         // 卡号
	DepositBank any         // 开户行
	Status      any         // 状态. 1=待审核；2=成功；3=失败
	AdminId     any         // 后台用户ID
	Remark      any         // 备注
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
