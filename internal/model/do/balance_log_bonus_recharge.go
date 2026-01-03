// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusRecharge is the golang structure of table balance_log_bonus_recharge for DAO operations like Where/Data.
type BalanceLogBonusRecharge struct {
	g.Meta        `orm:"table:balance_log_bonus_recharge, do:true"`
	Id            any         //
	SiteId        any         // 站点ID
	UserId        any         // 会员ID
	TradeType     any         // 类型。1=在线支付；2=转账汇款
	TradeNo       any         // 流水号
	PaymentLogId  any         // 在线充值日志ID
	TransferLogId any         // 转账汇款日志ID
	ActivityId    any         // 优惠活动ID
	Money         any         // 金额
	CreatedAt     *gtime.Time //
	UpdatedAt     *gtime.Time //
}
