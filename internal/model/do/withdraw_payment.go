// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// WithdrawPayment is the golang structure of table withdraw_payment for DAO operations like Where/Data.
type WithdrawPayment struct {
	g.Meta     `orm:"table:withdraw_payment, do:true"`
	Id         any         //
	SiteId     any         // 站点ID
	PaymentId  any         // 总站里提现ID
	Name       any         //
	Code       any         //
	Gateway    any         // 提现网关1在线支付2银行
	ApiUrl     any         // 提现请求地址
	MerchantNo any         // 商户号
	Md5Key     any         // MD5密钥
	EachMin    any         // 单笔最低。默认10
	EachMax    any         // 单笔最高。如果为0，表示没有限制。
	Sort       any         // 排序。值越小排名越靠前
	MoneyList  any         // 可选的金额数组，is_int =1 的时候必填
	IsInput    any         // 是否输入金额0不支持1支持
	Remark     any         // 描述
	Logo       any         // logo地址
	Status     any         //
	CreatedAt  *gtime.Time //
	UpdatedAt  *gtime.Time //
}
