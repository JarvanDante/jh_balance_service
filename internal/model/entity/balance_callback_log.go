// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceCallbackLog is the golang structure for table balance_callback_log.
type BalanceCallbackLog struct {
	Id            uint64      `json:"id"            orm:"id"             description:"主键ID"`
	SiteId        int         `json:"siteId"        orm:"site_id"        description:"站点ID"`
	CallbackType  int         `json:"callbackType"  orm:"callback_type"  description:"回调类型 1=充值 2=提现"`
	OrderSn       string      `json:"orderSn"       orm:"order_sn"       description:"订单号"`
	TradeStatus   string      `json:"tradeStatus"   orm:"trade_status"   description:"交易状态"`
	ActualAmount  float64     `json:"actualAmount"  orm:"actual_amount"  description:"实际金额"`
	FailureReason string      `json:"failureReason" orm:"failure_reason" description:"失败原因"`
	FailureStage  string      `json:"failureStage"  orm:"failure_stage"  description:"失败阶段"`
	RequestData   string      `json:"requestData"   orm:"request_data"   description:"请求数据(JSON)"`
	ErrorMessage  string      `json:"errorMessage"  orm:"error_message"  description:"错误详情"`
	RetryCount    int         `json:"retryCount"    orm:"retry_count"    description:"重试次数"`
	IsResolved    int         `json:"isResolved"    orm:"is_resolved"    description:"是否已解决 0=未解决 1=已解决"`
	ResolvedAt    *gtime.Time `json:"resolvedAt"    orm:"resolved_at"    description:"解决时间"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:"更新时间"`
}
