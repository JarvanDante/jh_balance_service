// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonusBirthday is the golang structure of table balance_log_bonus_birthday for DAO operations like Where/Data.
type BalanceLogBonusBirthday struct {
	g.Meta    `orm:"table:balance_log_bonus_birthday, do:true"`
	Id        any         //
	SiteId    any         // 站点ID
	UserId    any         // 会员ID
	TradeNo   any         // 流水号
	Year      any         // 申请的年份。每年只能申请一次
	Money     any         // 金额
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
}
