// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceDepositReward is the golang structure of table balance_deposit_reward for DAO operations like Where/Data.
type BalanceDepositReward struct {
	g.Meta      `orm:"table:balance_deposit_reward, do:true"`
	Id          any         //
	SiteId      any         //
	UserId      any         //
	Date        *gtime.Time // 充值日期年月日
	BonusAmount any         // 发送金额
	TradeNo     any         // 发放流水号/订单号
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
