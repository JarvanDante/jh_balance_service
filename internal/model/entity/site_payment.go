// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SitePayment is the golang structure for table site_payment.
type SitePayment struct {
	Id         uint        `json:"id"         orm:"id"          description:""`
	SiteId     int         `json:"siteId"     orm:"site_id"     description:"站点ID"`
	PaymentId  int         `json:"paymentId"  orm:"payment_id"  description:"总站里支付ID"`
	Name       string      `json:"name"       orm:"name"        description:""`
	Code       string      `json:"code"       orm:"code"        description:""`
	Gateway    int         `json:"gateway"    orm:"gateway"     description:"支付网关1在线支付2银行"`
	ApiUrl     string      `json:"apiUrl"     orm:"api_url"     description:"支付域名"`
	MerchantNo string      `json:"merchantNo" orm:"merchant_no" description:"商户号"`
	Md5Key     string      `json:"md5Key"     orm:"md5_key"     description:"MD5密钥"`
	EachMin    float64     `json:"eachMin"    orm:"each_min"    description:"单笔最低。默认10"`
	EachMax    float64     `json:"eachMax"    orm:"each_max"    description:"单笔最高。如果为0，表示没有限制。"`
	Sort       int         `json:"sort"       orm:"sort"        description:"排序。值越小排名越靠前"`
	MoneyList  string      `json:"moneyList"  orm:"moneyList"   description:"可选的金额数组，is_int =1 的时候必填"`
	IsInput    int         `json:"isInput"    orm:"is_input"    description:"是否输入金额0不支持1支持"`
	Remark     string      `json:"remark"     orm:"remark"      description:"描述"`
	Logo       string      `json:"logo"       orm:"logo"        description:"logo地址"`
	Status     int         `json:"status"     orm:"status"      description:""`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:""`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:""`
}
