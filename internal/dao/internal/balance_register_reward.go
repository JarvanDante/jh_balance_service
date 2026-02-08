// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceRegisterRewardDao is the data access object for the table balance_register_reward.
type BalanceRegisterRewardDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  BalanceRegisterRewardColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// BalanceRegisterRewardColumns defines and stores column names for the table balance_register_reward.
type BalanceRegisterRewardColumns struct {
	Id          string //
	SiteId      string //
	UserId      string //
	Date        string // 注册日期年月日
	RegisterIp  string // 注册IP
	DeviceId    string // 设备ID
	BonusAmount string //
	TradeNo     string // 发放流水号/订单号
	CreatedAt   string //
	UpdatedAt   string //
}

// balanceRegisterRewardColumns holds the columns for the table balance_register_reward.
var balanceRegisterRewardColumns = BalanceRegisterRewardColumns{
	Id:          "id",
	SiteId:      "site_id",
	UserId:      "user_id",
	Date:        "date",
	RegisterIp:  "register_ip",
	DeviceId:    "device_id",
	BonusAmount: "bonus_amount",
	TradeNo:     "trade_no",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewBalanceRegisterRewardDao creates and returns a new DAO object for table data access.
func NewBalanceRegisterRewardDao(handlers ...gdb.ModelHandler) *BalanceRegisterRewardDao {
	return &BalanceRegisterRewardDao{
		group:    "default",
		table:    "balance_register_reward",
		columns:  balanceRegisterRewardColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceRegisterRewardDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceRegisterRewardDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceRegisterRewardDao) Columns() BalanceRegisterRewardColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceRegisterRewardDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceRegisterRewardDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceRegisterRewardDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
