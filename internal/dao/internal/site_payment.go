// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SitePaymentDao is the data access object for the table site_payment.
type SitePaymentDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SitePaymentColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SitePaymentColumns defines and stores column names for the table site_payment.
type SitePaymentColumns struct {
	Id         string //
	SiteId     string // 站点ID
	PaymentId  string // 总站里支付ID
	Name       string //
	Code       string //
	Gateway    string // 支付网关1在线支付2银行
	ApiUrl     string // 支付域名
	MerchantNo string // 商户号
	Md5Key     string // MD5密钥
	EachMin    string // 单笔最低。默认10
	EachMax    string // 单笔最高。如果为0，表示没有限制。
	Sort       string // 排序。值越小排名越靠前
	MoneyList  string // 可选的金额数组，is_int =1 的时候必填
	IsInput    string // 是否输入金额0不支持1支持
	Remark     string // 描述
	Logo       string // logo地址
	Status     string //
	CreatedAt  string //
	UpdatedAt  string //
}

// sitePaymentColumns holds the columns for the table site_payment.
var sitePaymentColumns = SitePaymentColumns{
	Id:         "id",
	SiteId:     "site_id",
	PaymentId:  "payment_id",
	Name:       "name",
	Code:       "code",
	Gateway:    "gateway",
	ApiUrl:     "api_url",
	MerchantNo: "merchant_no",
	Md5Key:     "md5_key",
	EachMin:    "each_min",
	EachMax:    "each_max",
	Sort:       "sort",
	MoneyList:  "moneyList",
	IsInput:    "is_input",
	Remark:     "remark",
	Logo:       "logo",
	Status:     "status",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewSitePaymentDao creates and returns a new DAO object for table data access.
func NewSitePaymentDao(handlers ...gdb.ModelHandler) *SitePaymentDao {
	return &SitePaymentDao{
		group:    "default",
		table:    "site_payment",
		columns:  sitePaymentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SitePaymentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SitePaymentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SitePaymentDao) Columns() SitePaymentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SitePaymentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SitePaymentDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *SitePaymentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
