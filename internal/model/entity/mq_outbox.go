// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MqOutbox is the golang structure for table mq_outbox.
type MqOutbox struct {
	Id          uint64      `json:"id"          orm:"id"            description:""`
	BizKey      string      `json:"bizKey"      orm:"biz_key"       description:"业务幂等key，例如 trade_no"`
	Topic       string      `json:"topic"       orm:"topic"         description:"exchange 或 topic"`
	RoutingKey  string      `json:"routingKey"  orm:"routing_key"   description:"routing key"`
	Payload     string      `json:"payload"     orm:"payload"       description:""`
	Status      int         `json:"status"      orm:"status"        description:"0=待发送 1=已发送 2=发送失败"`
	RetryCount  int         `json:"retryCount"  orm:"retry_count"   description:""`
	NextRetryAt *gtime.Time `json:"nextRetryAt" orm:"next_retry_at" description:""`
	LastError   string      `json:"lastError"   orm:"last_error"    description:""`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"    description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"    description:""`
}
