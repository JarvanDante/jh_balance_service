// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogPaymentDao is the data access object for the table balance_log_payment.
type BalanceLogPaymentDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  BalanceLogPaymentColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// BalanceLogPaymentColumns defines and stores column names for the table balance_log_payment.
type BalanceLogPaymentColumns struct {
	Id                 string //
	SiteId             string // 站点ID
	UserId             string // 用户ID
	Username           string // 会员用户名
	ActivityRechargeId string // 充值活动ID
	Gateway            string // 网关类型
	PaymentId          string // 支付ID
	PaymentAccountId   string // 支付接口账号ID
	BankValue          string // 银行代码
	TradeNo            string // 流水号
	Money              string // 充值金额
	ActualMoney        string // 实际充值金额
	Fee                string // 手续费
	Status             string // 状态0=初始化，1=待处理；2=处理中；3=成功；4=失败
	AdminId            string // 后台管理员ID。默认为0
	Remark             string // 备注
	CreatedAt          string //
	UpdatedAt          string //
}

// balanceLogPaymentColumns holds the columns for the table balance_log_payment.
var balanceLogPaymentColumns = BalanceLogPaymentColumns{
	Id:                 "id",
	SiteId:             "site_id",
	UserId:             "user_id",
	Username:           "username",
	ActivityRechargeId: "activity_recharge_id",
	Gateway:            "gateway",
	PaymentId:          "payment_id",
	PaymentAccountId:   "payment_account_id",
	BankValue:          "bank_value",
	TradeNo:            "trade_no",
	Money:              "money",
	ActualMoney:        "actual_money",
	Fee:                "fee",
	Status:             "status",
	AdminId:            "admin_id",
	Remark:             "remark",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewBalanceLogPaymentDao creates and returns a new DAO object for table data access.
func NewBalanceLogPaymentDao(handlers ...gdb.ModelHandler) *BalanceLogPaymentDao {
	return &BalanceLogPaymentDao{
		group:    "default",
		table:    "balance_log_payment",
		columns:  balanceLogPaymentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogPaymentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogPaymentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogPaymentDao) Columns() BalanceLogPaymentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogPaymentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogPaymentDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogPaymentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
