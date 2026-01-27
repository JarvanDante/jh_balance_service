// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MqOutbox is the golang structure of table mq_outbox for DAO operations like Where/Data.
type MqOutbox struct {
	g.Meta      `orm:"table:mq_outbox, do:true"`
	Id          any         //
	BizKey      any         // 业务幂等key，例如 trade_no
	Topic       any         // exchange 或 topic
	RoutingKey  any         // routing key
	Payload     any         //
	Status      any         // 0=待发送 1=已发送 2=发送失败
	RetryCount  any         //
	NextRetryAt *gtime.Time //
	LastError   any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
