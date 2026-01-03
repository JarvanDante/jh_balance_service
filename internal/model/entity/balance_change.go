// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceChange is the golang structure for table balance_change.
type BalanceChange struct {
	Id            uint        `json:"id"            orm:"id"             description:""`
	SiteId        int         `json:"siteId"        orm:"site_id"        description:"站点ID"`
	TradeType     uint        `json:"tradeType"     orm:"trade_type"     description:"交易类型。1=转入；2=充值；3=红利；4=返水；5=转出；6=提现成功；7=提现返回；8=提现冻结；9=丢失补款；10=多出扣款"`
	UserId        int         `json:"userId"        orm:"user_id"        description:"用户ID"`
	Username      string      `json:"username"      orm:"username"       description:"用户名"`
	TradeNo       string      `json:"tradeNo"       orm:"trade_no"       description:"流水号"`
	BalanceOld    float64     `json:"balanceOld"    orm:"balance_old"    description:"旧的可用余额"`
	Money         float64     `json:"money"         orm:"money"          description:"变动金额"`
	BalanceNew    float64     `json:"balanceNew"    orm:"balance_new"    description:"新的可用余额"`
	BalanceFrozen float64     `json:"balanceFrozen" orm:"balance_frozen" description:"新的冻结余额"`
	Status        int         `json:"status"        orm:"status"         description:"账变状态。1=待审核；2=成功；3=失败"`
	Remark        string      `json:"remark"        orm:"remark"         description:""`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:""`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:""`
}
