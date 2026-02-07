// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogRebateDao is the data access object for the table balance_log_rebate.
type BalanceLogRebateDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  BalanceLogRebateColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// BalanceLogRebateColumns defines and stores column names for the table balance_log_rebate.
type BalanceLogRebateColumns struct {
	Id             string //
	SiteId         string // 站点ID
	UserId         string // 用户ID
	Username       string // 会员用户名
	TradeNo        string // 流水号
	LastBetId      string // 注单记录游标ID
	ValidBetAmount string // 有效投注总金额
	Money          string // 返水金额
	Fee            string // 手续费
	Status         string // 状态0=初始化，1=待处理；2=处理中；3=成功；4=失败
	AdminId        string // 后台管理员ID
	CreatedAt      string //
	UpdatedAt      string //
}

// balanceLogRebateColumns holds the columns for the table balance_log_rebate.
var balanceLogRebateColumns = BalanceLogRebateColumns{
	Id:             "id",
	SiteId:         "site_id",
	UserId:         "user_id",
	Username:       "username",
	TradeNo:        "trade_no",
	LastBetId:      "last_bet_id",
	ValidBetAmount: "valid_bet_amount",
	Money:          "money",
	Fee:            "fee",
	Status:         "status",
	AdminId:        "admin_id",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
}

// NewBalanceLogRebateDao creates and returns a new DAO object for table data access.
func NewBalanceLogRebateDao(handlers ...gdb.ModelHandler) *BalanceLogRebateDao {
	return &BalanceLogRebateDao{
		group:    "default",
		table:    "balance_log_rebate",
		columns:  balanceLogRebateColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogRebateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogRebateDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogRebateDao) Columns() BalanceLogRebateColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogRebateDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogRebateDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogRebateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
