// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusTransfer is the golang structure of table balance_log_bonus_transfer for DAO operations like Where/Data.
type BalanceLogBonusTransfer struct {
	g.Meta             `orm:"table:balance_log_bonus_transfer, do:true"`
	Id                 any         //
	SiteId             any         // 站点ID
	UserId             any         // 会员ID
	TradeNo            any         // 流水号
	TransferLogId      any         // 转账汇款日志ID
	ActivityRechargeId any         // 充值活动ID
	Money              any         // 金额
	CreatedAt          *gtime.Time //
	UpdatedAt          *gtime.Time //
}
