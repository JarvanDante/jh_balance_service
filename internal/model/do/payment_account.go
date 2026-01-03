// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PaymentAccount is the golang structure of table payment_account for DAO operations like Where/Data.
type PaymentAccount struct {
	g.Meta      `orm:"table:payment_account, do:true"`
	Id          any         //
	SiteId      any         //
	PaymentId   any         // 第三方支付ID
	Gateway     any         // 支付网关
	Name        any         // 接口名称
	Domain      any         // 支付域名
	MerchantNo  any         // 商户号
	Md5Key      any         // MD5密钥
	EachMin     any         // 单笔最低。默认10
	EachMax     any         // 单笔最高。如果为0，表示没有限制。
	DailyMax    any         // 单日停用上限。如果为0，表示没有限制。
	TodayCount  any         // 今日入款次数
	TodayAmount any         // 今日总转账
	Status      any         // 状态。1=启用；0=禁用
	Sort        any         // 排序。值越小排名越靠前
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
	PublicKey   any         // 公钥
	PrivateKey  any         // 私钥
	IsDecimal   any         // 是否携带小数，0为否，1为真
	IsInt       any         // 是否为规定整数数组，默认0，不需要 ，1需要
	MoneyList   any         // 可选的金额数组，is_int =1 的时候必填
}
