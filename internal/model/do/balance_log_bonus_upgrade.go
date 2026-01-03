// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusUpgrade is the golang structure of table balance_log_bonus_upgrade for DAO operations like Where/Data.
type BalanceLogBonusUpgrade struct {
	g.Meta    `orm:"table:balance_log_bonus_upgrade, do:true"`
	Id        any         //
	SiteId    any         // 站点ID
	UserId    any         // 会员ID
	TradeNo   any         // 流水号
	GradeId   any         // 会员等级ID
	Money     any         // 金额
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
}
