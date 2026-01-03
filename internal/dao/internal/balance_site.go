// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceSiteDao is the data access object for the table balance_site.
type BalanceSiteDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BalanceSiteColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BalanceSiteColumns defines and stores column names for the table balance_site.
type BalanceSiteColumns struct {
	Id               string //
	SiteId           string // 站点ID
	BalanceDefault   string // 站点默认额度。默认50W
	BalanceAvailable string // 站点可用额度
	CreatedAt        string //
	UpdatedAt        string //
}

// balanceSiteColumns holds the columns for the table balance_site.
var balanceSiteColumns = BalanceSiteColumns{
	Id:               "id",
	SiteId:           "site_id",
	BalanceDefault:   "balance_default",
	BalanceAvailable: "balance_available",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
}

// NewBalanceSiteDao creates and returns a new DAO object for table data access.
func NewBalanceSiteDao(handlers ...gdb.ModelHandler) *BalanceSiteDao {
	return &BalanceSiteDao{
		group:    "default",
		table:    "balance_site",
		columns:  balanceSiteColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceSiteDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceSiteDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceSiteDao) Columns() BalanceSiteColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceSiteDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceSiteDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceSiteDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
