// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogManual is the golang structure of table balance_log_manual for DAO operations like Where/Data.
type BalanceLogManual struct {
	g.Meta     `orm:"table:balance_log_manual, do:true"`
	Id         any         //
	SiteId     any         // 站点ID
	TradeType  any         // 交易类型。1=加款-充值；2=加款-红利；3=加款-返水；4=补款（不入充值记录），5=扣款 - 提现，6=扣款（不入提现记录）7=第三方加款（不入账变）,8=第三方扣款（不入账变）;9=第三方账户 -> 中心账户;10=中心账户 -> 第三方账户
	UserId     any         // 用户ID
	Username   any         // 会员用户名
	GameId     any         // 游戏ID
	TradeNo    any         // 流水号
	Money      any         // 金额
	Fee        any         // 手续费
	WaterTimes any         // 流水倍数
	Status     any         // 状态. 1=待审核；2=成功；3=失败
	AdminId    any         // 后台管理员ID
	Remark     any         // 备注
	CreatedAt  *gtime.Time //
	UpdatedAt  *gtime.Time //
}
