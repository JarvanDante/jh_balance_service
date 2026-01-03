// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogBonusPaymentDao is the data access object for the table balance_log_bonus_payment.
type BalanceLogBonusPaymentDao struct {
	table    string                        // table is the underlying table name of the DAO.
	group    string                        // group is the database configuration group name of the current DAO.
	columns  BalanceLogBonusPaymentColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler            // handlers for customized model modification.
}

// BalanceLogBonusPaymentColumns defines and stores column names for the table balance_log_bonus_payment.
type BalanceLogBonusPaymentColumns struct {
	Id                 string //
	SiteId             string // 站点ID
	UserId             string // 会员ID
	TradeNo            string // 流水号
	PaymentLogId       string // 在线充值日志ID
	ActivityRechargeId string // 充值活动ID
	Money              string // 金额
	CreatedAt          string //
	UpdatedAt          string //
}

// balanceLogBonusPaymentColumns holds the columns for the table balance_log_bonus_payment.
var balanceLogBonusPaymentColumns = BalanceLogBonusPaymentColumns{
	Id:                 "id",
	SiteId:             "site_id",
	UserId:             "user_id",
	TradeNo:            "trade_no",
	PaymentLogId:       "payment_log_id",
	ActivityRechargeId: "activity_recharge_id",
	Money:              "money",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewBalanceLogBonusPaymentDao creates and returns a new DAO object for table data access.
func NewBalanceLogBonusPaymentDao(handlers ...gdb.ModelHandler) *BalanceLogBonusPaymentDao {
	return &BalanceLogBonusPaymentDao{
		group:    "default",
		table:    "balance_log_bonus_payment",
		columns:  balanceLogBonusPaymentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogBonusPaymentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogBonusPaymentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogBonusPaymentDao) Columns() BalanceLogBonusPaymentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogBonusPaymentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogBonusPaymentDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogBonusPaymentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
