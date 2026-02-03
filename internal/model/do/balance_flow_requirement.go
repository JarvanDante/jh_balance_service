// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowRequirement is the golang structure of table balance_flow_requirement for DAO operations like Where/Data.
type BalanceFlowRequirement struct {
	g.Meta        `orm:"table:balance_flow_requirement, do:true"`
	Id            any         // 主键ID
	SiteId        any         // 站点ID
	UserId        any         // 用户ID
	Username      any         // 用户名
	SourceType    any         // 来源类型: 1=充值, 2=红利, 3=返水, 4=活动, 5=人工加款
	SourceId      any         // 来源ID（充值订单ID、红利ID等）
	TradeNo       any         // 关联交易号
	Amount        any         // 触发金额（充值金额、红利金额等）
	BonusAmount   any         // 赠送金额（如果有）
	FlowMultiple  any         // 流水倍数（如：1倍、3倍、5倍）
	RequiredFlow  any         // 需要完成的流水金额 = (amount + bonus_amount) * flow_multiple
	CompletedFlow any         // 已完成的流水金额
	RemainingFlow any         // 剩余流水金额 = required_flow - completed_flow
	Status        any         // 状态: 1=进行中, 2=已完成, 3=已过期, 4=已取消
	Progress      any         // 完成进度百分比（0-100）
	StartTime     *gtime.Time // 开始时间
	ExpireTime    *gtime.Time // 过期时间（NULL表示永久有效）
	CompleteTime  *gtime.Time // 完成时间
	Remark        any         // 备注说明
	AdminId       any         // 操作管理员ID（人工操作时）
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
}
