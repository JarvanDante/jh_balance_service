// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogDao is the data access object for the table balance_log.
type BalanceLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  BalanceLogColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// BalanceLogColumns defines and stores column names for the table balance_log.
type BalanceLogColumns struct {
	Id        string //
	SiteId    string // 站点ID
	UserId    string // 会员ID
	TradeType string // 交易类型
	Gateway   string // 支付相关：支付类型
	BankType  string // 转账相关：银行类型
	GameId    string // 游戏相关：游戏ID
	TradeNo   string // 交易流水号
	Money     string // 金额
	Fee       string // 手续费
	Status    string // 状态。0=待处理，1=处理中；2=成功；3=失败
	AdminId   string // 后台员工ID
	Remark    string // 备注
	CreatedAt string //
	UpdatedAt string //
}

// balanceLogColumns holds the columns for the table balance_log.
var balanceLogColumns = BalanceLogColumns{
	Id:        "id",
	SiteId:    "site_id",
	UserId:    "user_id",
	TradeType: "trade_type",
	Gateway:   "gateway",
	BankType:  "bank_type",
	GameId:    "game_id",
	TradeNo:   "trade_no",
	Money:     "money",
	Fee:       "fee",
	Status:    "status",
	AdminId:   "admin_id",
	Remark:    "remark",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewBalanceLogDao creates and returns a new DAO object for table data access.
func NewBalanceLogDao(handlers ...gdb.ModelHandler) *BalanceLogDao {
	return &BalanceLogDao{
		group:    "default",
		table:    "balance_log",
		columns:  balanceLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogDao) Columns() BalanceLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
