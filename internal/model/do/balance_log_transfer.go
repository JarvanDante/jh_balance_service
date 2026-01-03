// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogTransfer is the golang structure of table balance_log_transfer for DAO operations like Where/Data.
type BalanceLogTransfer struct {
	g.Meta             `orm:"table:balance_log_transfer, do:true"`
	Id                 any         //
	SiteId             any         // 站点ID
	UserId             any         // 用户ID
	Username           any         // 会员用户名
	ActivityRechargeId any         // 充值活动ID
	BankType           any         // 转账汇款类型
	AccountId          any         // 转账汇款账号ID
	TradeNo            any         // 流水号
	ThirdTradeNo       any         // 第三方订单号
	Money              any         // 充值金额
	Fee                any         // 手续费
	Status             any         // 状态. 1=待审核；2=成功；3=失败
	TransferAccount    any         // 汇款账户名
	Code               any         // 随机码。其实就是会员用户名
	AdminId            any         // 操作人id
	Remark             any         // 备注
	CreatedAt          *gtime.Time //
	UpdatedAt          *gtime.Time //
}
