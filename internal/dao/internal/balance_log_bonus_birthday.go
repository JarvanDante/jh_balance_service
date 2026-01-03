// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogBonusBirthdayDao is the data access object for the table balance_log_bonus_birthday.
type BalanceLogBonusBirthdayDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  BalanceLogBonusBirthdayColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// BalanceLogBonusBirthdayColumns defines and stores column names for the table balance_log_bonus_birthday.
type BalanceLogBonusBirthdayColumns struct {
	Id        string //
	SiteId    string // 站点ID
	UserId    string // 会员ID
	TradeNo   string // 流水号
	Year      string // 申请的年份。每年只能申请一次
	Money     string // 金额
	CreatedAt string //
	UpdatedAt string //
}

// balanceLogBonusBirthdayColumns holds the columns for the table balance_log_bonus_birthday.
var balanceLogBonusBirthdayColumns = BalanceLogBonusBirthdayColumns{
	Id:        "id",
	SiteId:    "site_id",
	UserId:    "user_id",
	TradeNo:   "trade_no",
	Year:      "year",
	Money:     "money",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewBalanceLogBonusBirthdayDao creates and returns a new DAO object for table data access.
func NewBalanceLogBonusBirthdayDao(handlers ...gdb.ModelHandler) *BalanceLogBonusBirthdayDao {
	return &BalanceLogBonusBirthdayDao{
		group:    "default",
		table:    "balance_log_bonus_birthday",
		columns:  balanceLogBonusBirthdayColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogBonusBirthdayDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogBonusBirthdayDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogBonusBirthdayDao) Columns() BalanceLogBonusBirthdayColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogBonusBirthdayDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogBonusBirthdayDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogBonusBirthdayDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
