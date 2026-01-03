// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PointsChange is the golang structure of table points_change for DAO operations like Where/Data.
type PointsChange struct {
	g.Meta       `orm:"table:points_change, do:true"`
	Id           any         //
	SiteId       any         // 站点ID
	ChangeType   any         // 变化类型。1=增加；0=减少
	TradeType    any         // 类型。1=签到;2=充值;3=投注
	UserId       any         // 会员ID
	Username     any         // 会员用户名
	PointsOld    any         // 原积分
	PointsChange any         // 积分变动
	PointsNew    any         // 现积分
	Status       any         // 状态。1=待处理；2=成功；3=失败
	AdminId      any         // 后台员工ID
	Remark       any         //
	CreatedAt    *gtime.Time //
	UpdatedAt    *gtime.Time //
	TradeNo      any         //
}
