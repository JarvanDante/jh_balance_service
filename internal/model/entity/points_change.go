// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PointsChange is the golang structure for table points_change.
type PointsChange struct {
	Id           uint        `json:"id"           orm:"id"            description:""`
	SiteId       int         `json:"siteId"       orm:"site_id"       description:"站点ID"`
	ChangeType   int         `json:"changeType"   orm:"change_type"   description:"变化类型。1=增加；0=减少"`
	TradeType    int         `json:"tradeType"    orm:"trade_type"    description:"类型。1=签到;2=充值;3=投注"`
	UserId       int         `json:"userId"       orm:"user_id"       description:"会员ID"`
	Username     string      `json:"username"     orm:"username"      description:"会员用户名"`
	PointsOld    int         `json:"pointsOld"    orm:"points_old"    description:"原积分"`
	PointsChange int         `json:"pointsChange" orm:"points_change" description:"积分变动"`
	PointsNew    int         `json:"pointsNew"    orm:"points_new"    description:"现积分"`
	Status       int         `json:"status"       orm:"status"        description:"状态。1=待处理；2=成功；3=失败"`
	AdminId      int         `json:"adminId"      orm:"admin_id"      description:"后台员工ID"`
	Remark       string      `json:"remark"       orm:"remark"        description:""`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:""`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:""`
	TradeNo      string      `json:"tradeNo"      orm:"trade_no"      description:""`
}
