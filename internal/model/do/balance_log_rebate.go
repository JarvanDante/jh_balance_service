// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogRebate is the golang structure of table balance_log_rebate for DAO operations like Where/Data.
type BalanceLogRebate struct {
	g.Meta         `orm:"table:balance_log_rebate, do:true"`
	Id             any         //
	SiteId         any         // 站点ID
	UserId         any         // 用户ID
	Username       any         // 会员用户名
	TradeNo        any         // 流水号
	RebateDate     *gtime.Time // 返水日期
	ValidBetAmount any         // 有效投注总金额
	Money          any         // 返水金额
	Fee            any         // 手续费
	Status         any         // 状态。0=未确认；1=已确认
	AdminId        any         // 后台管理员ID
	CreatedAt      *gtime.Time //
	UpdatedAt      *gtime.Time //
}
