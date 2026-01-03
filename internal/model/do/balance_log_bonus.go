// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogBonus is the golang structure of table balance_log_bonus for DAO operations like Where/Data.
type BalanceLogBonus struct {
	g.Meta             `orm:"table:balance_log_bonus, do:true"`
	Id                 any         //
	SiteId             any         // 站点ID
	TradeType          any         // 交易类型。1=在线充值红利；2=转账汇款红利；3=生日红利；4=签到红利；
	UserId             any         // 会员ID
	Username           any         // 会员用户名
	TradeNo            any         // 流水号
	ActivityRechargeId any         // 充值活动ID
	Money              any         // 金额
	Fee                any         // 手续费
	Status             any         // 状态。1=待处理；2=成功；3=失败
	AdminId            any         // 后台管理员ID。默认为0
	Remark             any         // 备注
	CreatedAt          *gtime.Time //
	UpdatedAt          *gtime.Time //
}
