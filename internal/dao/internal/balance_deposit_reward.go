// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceDepositRewardDao is the data access object for the table balance_deposit_reward.
type BalanceDepositRewardDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  BalanceDepositRewardColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// BalanceDepositRewardColumns defines and stores column names for the table balance_deposit_reward.
type BalanceDepositRewardColumns struct {
	Id          string //
	SiteId      string //
	UserId      string //
	Date        string // 充值日期年月日
	BonusAmount string // 发送金额
	TradeNo     string // 发放流水号/订单号
	CreatedAt   string //
	UpdatedAt   string //
}

// balanceDepositRewardColumns holds the columns for the table balance_deposit_reward.
var balanceDepositRewardColumns = BalanceDepositRewardColumns{
	Id:          "id",
	SiteId:      "site_id",
	UserId:      "user_id",
	Date:        "date",
	BonusAmount: "bonus_amount",
	TradeNo:     "trade_no",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewBalanceDepositRewardDao creates and returns a new DAO object for table data access.
func NewBalanceDepositRewardDao(handlers ...gdb.ModelHandler) *BalanceDepositRewardDao {
	return &BalanceDepositRewardDao{
		group:    "default",
		table:    "balance_deposit_reward",
		columns:  balanceDepositRewardColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceDepositRewardDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceDepositRewardDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceDepositRewardDao) Columns() BalanceDepositRewardColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceDepositRewardDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceDepositRewardDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceDepositRewardDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
