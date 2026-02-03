// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceFlowRequirement is the golang structure for table balance_flow_requirement.
type BalanceFlowRequirement struct {
	Id            uint64      `json:"id"            orm:"id"             description:"主键ID"`
	SiteId        int         `json:"siteId"        orm:"site_id"        description:"站点ID"`
	UserId        int64       `json:"userId"        orm:"user_id"        description:"用户ID"`
	Username      string      `json:"username"      orm:"username"       description:"用户名"`
	SourceType    int         `json:"sourceType"    orm:"source_type"    description:"来源类型: 1=充值, 2=红利, 3=返水, 4=活动, 5=人工加款"`
	SourceId      int64       `json:"sourceId"      orm:"source_id"      description:"来源ID（充值订单ID、红利ID等）"`
	TradeNo       string      `json:"tradeNo"       orm:"trade_no"       description:"关联交易号"`
	Amount        float64     `json:"amount"        orm:"amount"         description:"触发金额（充值金额、红利金额等）"`
	BonusAmount   float64     `json:"bonusAmount"   orm:"bonus_amount"   description:"赠送金额（如果有）"`
	FlowMultiple  float64     `json:"flowMultiple"  orm:"flow_multiple"  description:"流水倍数（如：1倍、3倍、5倍）"`
	RequiredFlow  float64     `json:"requiredFlow"  orm:"required_flow"  description:"需要完成的流水金额 = (amount + bonus_amount) * flow_multiple"`
	CompletedFlow float64     `json:"completedFlow" orm:"completed_flow" description:"已完成的流水金额"`
	RemainingFlow float64     `json:"remainingFlow" orm:"remaining_flow" description:"剩余流水金额 = required_flow - completed_flow"`
	Status        int         `json:"status"        orm:"status"         description:"状态: 1=进行中, 2=已完成, 3=已过期, 4=已取消"`
	Progress      float64     `json:"progress"      orm:"progress"       description:"完成进度百分比（0-100）"`
	StartTime     *gtime.Time `json:"startTime"     orm:"start_time"     description:"开始时间"`
	ExpireTime    *gtime.Time `json:"expireTime"    orm:"expire_time"    description:"过期时间（NULL表示永久有效）"`
	CompleteTime  *gtime.Time `json:"completeTime"  orm:"complete_time"  description:"完成时间"`
	Remark        string      `json:"remark"        orm:"remark"         description:"备注说明"`
	AdminId       int         `json:"adminId"       orm:"admin_id"       description:"操作管理员ID（人工操作时）"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:"更新时间"`
}
