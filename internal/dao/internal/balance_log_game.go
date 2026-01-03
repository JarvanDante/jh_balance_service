// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogGameDao is the data access object for the table balance_log_game.
type BalanceLogGameDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  BalanceLogGameColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// BalanceLogGameColumns defines and stores column names for the table balance_log_game.
type BalanceLogGameColumns struct {
	Id                string //
	SiteId            string // 站点ID
	UserId            string // 会员ID
	GameId            string // 游戏ID
	InOrOut           string // 转入还是转出；0=转出；1=转入
	TradeNo           string //
	Money             string // 操作金额
	Status            string // 状态. 1=待确认，2=成功，3=失败
	Remark            string // 备注
	ThirdPartyTradeNo string // 第三方订单号
	CreatedAt         string //
	UpdatedAt         string //
}

// balanceLogGameColumns holds the columns for the table balance_log_game.
var balanceLogGameColumns = BalanceLogGameColumns{
	Id:                "id",
	SiteId:            "site_id",
	UserId:            "user_id",
	GameId:            "game_id",
	InOrOut:           "in_or_out",
	TradeNo:           "trade_no",
	Money:             "money",
	Status:            "status",
	Remark:            "remark",
	ThirdPartyTradeNo: "third_party_trade_no",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewBalanceLogGameDao creates and returns a new DAO object for table data access.
func NewBalanceLogGameDao(handlers ...gdb.ModelHandler) *BalanceLogGameDao {
	return &BalanceLogGameDao{
		group:    "default",
		table:    "balance_log_game",
		columns:  balanceLogGameColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogGameDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogGameDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogGameDao) Columns() BalanceLogGameColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogGameDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogGameDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogGameDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
