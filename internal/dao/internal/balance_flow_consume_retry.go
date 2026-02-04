// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceFlowConsumeRetryDao is the data access object for the table balance_flow_consume_retry.
type BalanceFlowConsumeRetryDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  BalanceFlowConsumeRetryColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// BalanceFlowConsumeRetryColumns defines and stores column names for the table balance_flow_consume_retry.
type BalanceFlowConsumeRetryColumns struct {
	Id          string //
	SiteId      string //
	UserId      string //
	OrderNo     string //
	ValidAmount string //
	ErrorMsg    string //
	CreatedAt   string //
}

// balanceFlowConsumeRetryColumns holds the columns for the table balance_flow_consume_retry.
var balanceFlowConsumeRetryColumns = BalanceFlowConsumeRetryColumns{
	Id:          "id",
	SiteId:      "site_id",
	UserId:      "user_id",
	OrderNo:     "order_no",
	ValidAmount: "valid_amount",
	ErrorMsg:    "error_msg",
	CreatedAt:   "created_at",
}

// NewBalanceFlowConsumeRetryDao creates and returns a new DAO object for table data access.
func NewBalanceFlowConsumeRetryDao(handlers ...gdb.ModelHandler) *BalanceFlowConsumeRetryDao {
	return &BalanceFlowConsumeRetryDao{
		group:    "default",
		table:    "balance_flow_consume_retry",
		columns:  balanceFlowConsumeRetryColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceFlowConsumeRetryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceFlowConsumeRetryDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceFlowConsumeRetryDao) Columns() BalanceFlowConsumeRetryColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceFlowConsumeRetryDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceFlowConsumeRetryDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceFlowConsumeRetryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
