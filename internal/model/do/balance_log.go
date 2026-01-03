// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLog is the golang structure of table balance_log for DAO operations like Where/Data.
type BalanceLog struct {
	g.Meta    `orm:"table:balance_log, do:true"`
	Id        any         //
	SiteId    any         // 站点ID
	UserId    any         // 会员ID
	TradeType any         // 交易类型
	Gateway   any         // 支付相关：支付类型
	BankType  any         // 转账相关：银行类型
	GameId    any         // 游戏相关：游戏ID
	TradeNo   any         // 交易流水号
	Money     any         // 金额
	Fee       any         // 手续费
	Status    any         // 状态。0=待处理，1=处理中；2=成功；3=失败
	AdminId   any         // 后台员工ID
	Remark    any         // 备注
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
}
