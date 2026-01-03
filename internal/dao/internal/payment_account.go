// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PaymentAccountDao is the data access object for the table payment_account.
type PaymentAccountDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  PaymentAccountColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// PaymentAccountColumns defines and stores column names for the table payment_account.
type PaymentAccountColumns struct {
	Id          string //
	SiteId      string //
	PaymentId   string // 第三方支付ID
	Gateway     string // 支付网关
	Name        string // 接口名称
	Domain      string // 支付域名
	MerchantNo  string // 商户号
	Md5Key      string // MD5密钥
	EachMin     string // 单笔最低。默认10
	EachMax     string // 单笔最高。如果为0，表示没有限制。
	DailyMax    string // 单日停用上限。如果为0，表示没有限制。
	TodayCount  string // 今日入款次数
	TodayAmount string // 今日总转账
	Status      string // 状态。1=启用；0=禁用
	Sort        string // 排序。值越小排名越靠前
	CreatedAt   string //
	UpdatedAt   string //
	PublicKey   string // 公钥
	PrivateKey  string // 私钥
	IsDecimal   string // 是否携带小数，0为否，1为真
	IsInt       string // 是否为规定整数数组，默认0，不需要 ，1需要
	MoneyList   string // 可选的金额数组，is_int =1 的时候必填
}

// paymentAccountColumns holds the columns for the table payment_account.
var paymentAccountColumns = PaymentAccountColumns{
	Id:          "id",
	SiteId:      "site_id",
	PaymentId:   "payment_id",
	Gateway:     "gateway",
	Name:        "name",
	Domain:      "domain",
	MerchantNo:  "merchant_no",
	Md5Key:      "md5_key",
	EachMin:     "each_min",
	EachMax:     "each_max",
	DailyMax:    "daily_max",
	TodayCount:  "today_count",
	TodayAmount: "today_amount",
	Status:      "status",
	Sort:        "sort",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	PublicKey:   "public_key",
	PrivateKey:  "private_key",
	IsDecimal:   "is_decimal",
	IsInt:       "is_int",
	MoneyList:   "moneyList",
}

// NewPaymentAccountDao creates and returns a new DAO object for table data access.
func NewPaymentAccountDao(handlers ...gdb.ModelHandler) *PaymentAccountDao {
	return &PaymentAccountDao{
		group:    "default",
		table:    "payment_account",
		columns:  paymentAccountColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PaymentAccountDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PaymentAccountDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PaymentAccountDao) Columns() PaymentAccountColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PaymentAccountDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PaymentAccountDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PaymentAccountDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
