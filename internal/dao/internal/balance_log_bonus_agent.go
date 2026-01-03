// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogBonusAgentDao is the data access object for the table balance_log_bonus_agent.
type BalanceLogBonusAgentDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  BalanceLogBonusAgentColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// BalanceLogBonusAgentColumns defines and stores column names for the table balance_log_bonus_agent.
type BalanceLogBonusAgentColumns struct {
	Id           string //
	SiteId       string //
	UserId       string //
	AgentId      string //
	AgentLevelId string //
	TradeNo      string // 订单号
	Money        string //
	CreatedAt    string //
	UpdatedAt    string //
}

// balanceLogBonusAgentColumns holds the columns for the table balance_log_bonus_agent.
var balanceLogBonusAgentColumns = BalanceLogBonusAgentColumns{
	Id:           "id",
	SiteId:       "site_id",
	UserId:       "user_id",
	AgentId:      "agent_id",
	AgentLevelId: "agent_level_id",
	TradeNo:      "trade_no",
	Money:        "money",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewBalanceLogBonusAgentDao creates and returns a new DAO object for table data access.
func NewBalanceLogBonusAgentDao(handlers ...gdb.ModelHandler) *BalanceLogBonusAgentDao {
	return &BalanceLogBonusAgentDao{
		group:    "default",
		table:    "balance_log_bonus_agent",
		columns:  balanceLogBonusAgentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogBonusAgentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogBonusAgentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogBonusAgentDao) Columns() BalanceLogBonusAgentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogBonusAgentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogBonusAgentDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogBonusAgentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
