// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserReportTotal is the golang structure of table user_report_total for DAO operations like Where/Data.
type UserReportTotal struct {
	g.Meta            `orm:"table:user_report_total, do:true"`
	Id                any         // 主键
	SiteId            any         // 站点ID
	UserId            any         // 用户ID
	DepositAmount     any         // 累计充值总额
	BetAmount         any         // 累计有效投注总额
	LastDepositId     any         // 充值记录游标ID
	LastBetId         any         // 注单记录游标ID
	LastDepositTime   *gtime.Time // 最近累计的充值时间
	LastBetSettleTime *gtime.Time // 最近累计的注单结算时间
	CreatedAt         *gtime.Time // 创建时间
	UpdatedAt         *gtime.Time // 更新时间
}
