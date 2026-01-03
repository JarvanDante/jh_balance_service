// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceRebateHistoryDao is the data access object for the table balance_rebate_history.
type BalanceRebateHistoryDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  BalanceRebateHistoryColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// BalanceRebateHistoryColumns defines and stores column names for the table balance_rebate_history.
type BalanceRebateHistoryColumns struct {
	Id             string //
	SiteId         string // 站点ID
	RebateDate     string // 返水日期
	UserCount      string // 返水人数
	ValidBetAmount string // 有效投注金额
	Money          string // 返水金额
	CreatedAt      string //
	UpdatedAt      string //
}

// balanceRebateHistoryColumns holds the columns for the table balance_rebate_history.
var balanceRebateHistoryColumns = BalanceRebateHistoryColumns{
	Id:             "id",
	SiteId:         "site_id",
	RebateDate:     "rebate_date",
	UserCount:      "user_count",
	ValidBetAmount: "valid_bet_amount",
	Money:          "money",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
}

// NewBalanceRebateHistoryDao creates and returns a new DAO object for table data access.
func NewBalanceRebateHistoryDao(handlers ...gdb.ModelHandler) *BalanceRebateHistoryDao {
	return &BalanceRebateHistoryDao{
		group:    "default",
		table:    "balance_rebate_history",
		columns:  balanceRebateHistoryColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceRebateHistoryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceRebateHistoryDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceRebateHistoryDao) Columns() BalanceRebateHistoryColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceRebateHistoryDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceRebateHistoryDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceRebateHistoryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
