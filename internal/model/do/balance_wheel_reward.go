// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceWheelReward is the golang structure of table balance_wheel_reward for DAO operations like Where/Data.
type BalanceWheelReward struct {
	g.Meta       `orm:"table:balance_wheel_reward, do:true"`
	Id           any         //
	SiteId       any         // 站点ID
	UserId       any         // 用户ID
	Date         *gtime.Time // 抽奖日期(yyyy-mm-dd)
	SegmentNo    any         // 命中扇区(1~12)
	PrizeName    any         // 奖项名称
	RewardAmount any         // 中奖金额
	TradeNo      any         // 发放流水号/订单号
	CreatedAt    *gtime.Time //
	UpdatedAt    *gtime.Time //
}
