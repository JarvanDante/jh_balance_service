// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceFlowConsumedBetsDao is the data access object for the table balance_flow_consumed_bets.
type BalanceFlowConsumedBetsDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  BalanceFlowConsumedBetsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// BalanceFlowConsumedBetsColumns defines and stores column names for the table balance_flow_consumed_bets.
type BalanceFlowConsumedBetsColumns struct {
	Id         string //
	SiteId     string //
	UserId     string //
	OrderNo    string //
	ConsumedAt string //
}

// balanceFlowConsumedBetsColumns holds the columns for the table balance_flow_consumed_bets.
var balanceFlowConsumedBetsColumns = BalanceFlowConsumedBetsColumns{
	Id:         "id",
	SiteId:     "site_id",
	UserId:     "user_id",
	OrderNo:    "order_no",
	ConsumedAt: "consumed_at",
}

// NewBalanceFlowConsumedBetsDao creates and returns a new DAO object for table data access.
func NewBalanceFlowConsumedBetsDao(handlers ...gdb.ModelHandler) *BalanceFlowConsumedBetsDao {
	return &BalanceFlowConsumedBetsDao{
		group:    "default",
		table:    "balance_flow_consumed_bets",
		columns:  balanceFlowConsumedBetsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceFlowConsumedBetsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceFlowConsumedBetsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceFlowConsumedBetsDao) Columns() BalanceFlowConsumedBetsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceFlowConsumedBetsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceFlowConsumedBetsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceFlowConsumedBetsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
