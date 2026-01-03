// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// BalanceSite is the golang structure of table balance_site for DAO operations like Where/Data.
type BalanceSite struct {
	g.Meta           `orm:"table:balance_site, do:true"`
	Id               any         //
	SiteId           any         // 站点ID
	BalanceDefault   any         // 站点默认额度。默认50W
	BalanceAvailable any         // 站点可用额度
	CreatedAt        *gtime.Time //
	UpdatedAt        *gtime.Time //
}
