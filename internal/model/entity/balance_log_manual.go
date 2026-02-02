// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceLogManual is the golang structure for table balance_log_manual.
type BalanceLogManual struct {
	Id         uint64      `json:"id"         orm:"id"          description:""`
	SiteId     int         `json:"siteId"     orm:"site_id"     description:"站点ID"`
	TradeType  int         `json:"tradeType"  orm:"trade_type"  description:"交易类型。1=加款-充值；2=加款-红利；3=加款-返水；4=补款（不入充值记录），5=扣款 - 提现，6=扣款（不入提现记录）7=第三方加款（不入账变）,8=第三方扣款（不入账变）;9=第三方账户 -> 中心账户;10=中心账户 -> 第三方账户"`
	UserId     int         `json:"userId"     orm:"user_id"     description:"用户ID"`
	Username   string      `json:"username"   orm:"username"    description:"会员用户名"`
	GameId     int         `json:"gameId"     orm:"game_id"     description:"游戏ID"`
	TradeNo    string      `json:"tradeNo"    orm:"trade_no"    description:"流水号"`
	Money      float64     `json:"money"      orm:"money"       description:"金额"`
	Fee        float64     `json:"fee"        orm:"fee"         description:"手续费"`
	WaterTimes int         `json:"waterTimes" orm:"water_times" description:"流水倍数"`
	Status     int         `json:"status"     orm:"status"      description:"状态0=初始化，1=待处理；2=处理中；3=成功；4=失败"`
	AdminId    int         `json:"adminId"    orm:"admin_id"    description:"后台管理员ID"`
	Remark     string      `json:"remark"     orm:"remark"      description:"备注"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:""`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:""`
}
