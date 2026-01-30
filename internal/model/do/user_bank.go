// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserBank is the golang structure of table user_bank for DAO operations like Where/Data.
type UserBank struct {
	g.Meta      `orm:"table:user_bank, do:true"`
	Id          any         //
	SiteId      any         // 站点ID
	UserId      any         // 会员ID
	Type        any         // 银行类型1=GCash,2=Maya
	CardAccount any         // 银行账户名
	CardNo      any         // 银行卡号
	IsDefault   any         // 默认地址。1=默认，0=非默认
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
