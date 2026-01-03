// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogPayment is the golang structure of table balance_log_payment for DAO operations like Where/Data.
type BalanceLogPayment struct {
	g.Meta             `orm:"table:balance_log_payment, do:true"`
	Id                 any         //
	SiteId             any         // 站点ID
	UserId             any         // 用户ID
	Username           any         // 会员用户名
	ActivityRechargeId any         // 充值活动ID
	Gateway            any         // 网关类型
	PaymentId          any         // 支付ID
	PaymentAccountId   any         // 支付接口账号ID
	BankValue          any         // 银行代码
	TradeNo            any         // 流水号
	Money              any         // 充值金额
	Fee                any         // 手续费
	Status             any         // 状态。1=未付；2=成功；3=失败
	AdminId            any         // 后台管理员ID。默认为0
	Remark             any         // 备注
	CreatedAt          *gtime.Time //
	UpdatedAt          *gtime.Time //
}
