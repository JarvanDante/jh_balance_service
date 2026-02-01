// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceCallbackLog is the golang structure of table balance_callback_log for DAO operations like Where/Data.
type BalanceCallbackLog struct {
	g.Meta        `orm:"table:balance_callback_log, do:true"`
	Id            any         // 主键ID
	SiteId        any         // 站点ID
	CallbackType  any         // 回调类型 1=充值 2=提现
	OrderSn       any         // 订单号
	TradeStatus   any         // 交易状态
	ActualAmount  any         // 实际金额
	FailureReason any         // 失败原因
	FailureStage  any         // 失败阶段
	RequestData   any         // 请求数据(JSON)
	ErrorMessage  any         // 错误详情
	RetryCount    any         // 重试次数
	IsResolved    any         // 是否已解决 0=未解决 1=已解决
	ResolvedAt    *gtime.Time // 解决时间
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
}
