// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusPayment is the golang structure of table balance_log_bonus_payment for DAO operations like Where/Data.
type BalanceLogBonusPayment struct {
	g.Meta             `orm:"table:balance_log_bonus_payment, do:true"`
	Id                 any         //
	SiteId             any         // 站点ID
	UserId             any         // 会员ID
	TradeNo            any         // 流水号
	PaymentLogId       any         // 在线充值日志ID
	ActivityRechargeId any         // 充值活动ID
	Money              any         // 金额
	CreatedAt          *gtime.Time //
	UpdatedAt          *gtime.Time //
}
