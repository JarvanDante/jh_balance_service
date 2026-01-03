// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PaymentAccount is the golang structure for table payment_account.
type PaymentAccount struct {
	Id          uint        `json:"id"          orm:"id"           description:""`
	SiteId      int         `json:"siteId"      orm:"site_id"      description:""`
	PaymentId   int         `json:"paymentId"   orm:"payment_id"   description:"第三方支付ID"`
	Gateway     int         `json:"gateway"     orm:"gateway"      description:"支付网关"`
	Name        string      `json:"name"        orm:"name"         description:"接口名称"`
	Domain      string      `json:"domain"      orm:"domain"       description:"支付域名"`
	MerchantNo  string      `json:"merchantNo"  orm:"merchant_no"  description:"商户号"`
	Md5Key      string      `json:"md5Key"      orm:"md5_key"      description:"MD5密钥"`
	EachMin     float64     `json:"eachMin"     orm:"each_min"     description:"单笔最低。默认10"`
	EachMax     float64     `json:"eachMax"     orm:"each_max"     description:"单笔最高。如果为0，表示没有限制。"`
	DailyMax    float64     `json:"dailyMax"    orm:"daily_max"    description:"单日停用上限。如果为0，表示没有限制。"`
	TodayCount  int         `json:"todayCount"  orm:"today_count"  description:"今日入款次数"`
	TodayAmount float64     `json:"todayAmount" orm:"today_amount" description:"今日总转账"`
	Status      int         `json:"status"      orm:"status"       description:"状态。1=启用；0=禁用"`
	Sort        int         `json:"sort"        orm:"sort"         description:"排序。值越小排名越靠前"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`
	PublicKey   string      `json:"publicKey"   orm:"public_key"   description:"公钥"`
	PrivateKey  string      `json:"privateKey"  orm:"private_key"  description:"私钥"`
	IsDecimal   int         `json:"isDecimal"   orm:"is_decimal"   description:"是否携带小数，0为否，1为真"`
	IsInt       int         `json:"isInt"       orm:"is_int"       description:"是否为规定整数数组，默认0，不需要 ，1需要"`
	MoneyList   string      `json:"moneyList"   orm:"moneyList"    description:"可选的金额数组，is_int =1 的时候必填"`
}
