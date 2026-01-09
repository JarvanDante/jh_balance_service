// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceUserDao is the data access object for the table balance_user.
type BalanceUserDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BalanceUserColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BalanceUserColumns defines and stores column names for the table balance_user.
type BalanceUserColumns struct {
	Id            string //
	SiteId        string // 站点ID
	UserId        string // 会员ID
	Username      string // 会员用户名
	Balance       string // 可用金额
	BalanceFrozen string // 冻结金额
	BalanceDebt   string // 负债金额。防止账户余额出现负数的情况。例如：注单结算错误，需要扣掉会员账户金额。如果会员账户金额足够，则直接扣除会员账户金额；如果会员账户金额不足，则不扣除会员账户金额，直接存入负债金额。在下次会员有资金变动的时候，先来计算这笔金额
	Points        string // 会员积分
	BalanceStatus string // 资金状态。1=正常；0=锁定
	CreatedAt     string //
	UpdatedAt     string //
}

// balanceUserColumns holds the columns for the table balance_user.
var balanceUserColumns = BalanceUserColumns{
	Id:            "id",
	SiteId:        "site_id",
	UserId:        "user_id",
	Username:      "username",
	Balance:       "balance",
	BalanceFrozen: "balance_frozen",
	BalanceDebt:   "balance_debt",
	Points:        "points",
	BalanceStatus: "balance_status",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewBalanceUserDao creates and returns a new DAO object for table data access.
func NewBalanceUserDao(handlers ...gdb.ModelHandler) *BalanceUserDao {
	return &BalanceUserDao{
		group:    "default",
		table:    "balance_user",
		columns:  balanceUserColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceUserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceUserDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceUserDao) Columns() BalanceUserColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceUserDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceUserDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
