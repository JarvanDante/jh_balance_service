// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogBonusRechargeDao is the data access object for the table balance_log_bonus_recharge.
type BalanceLogBonusRechargeDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  BalanceLogBonusRechargeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// BalanceLogBonusRechargeColumns defines and stores column names for the table balance_log_bonus_recharge.
type BalanceLogBonusRechargeColumns struct {
	Id            string //
	SiteId        string // 站点ID
	UserId        string // 会员ID
	TradeType     string // 类型。1=在线支付；2=转账汇款
	TradeNo       string // 流水号
	PaymentLogId  string // 在线充值日志ID
	TransferLogId string // 转账汇款日志ID
	ActivityId    string // 优惠活动ID
	Money         string // 金额
	CreatedAt     string //
	UpdatedAt     string //
}

// balanceLogBonusRechargeColumns holds the columns for the table balance_log_bonus_recharge.
var balanceLogBonusRechargeColumns = BalanceLogBonusRechargeColumns{
	Id:            "id",
	SiteId:        "site_id",
	UserId:        "user_id",
	TradeType:     "trade_type",
	TradeNo:       "trade_no",
	PaymentLogId:  "payment_log_id",
	TransferLogId: "transfer_log_id",
	ActivityId:    "activity_id",
	Money:         "money",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewBalanceLogBonusRechargeDao creates and returns a new DAO object for table data access.
func NewBalanceLogBonusRechargeDao(handlers ...gdb.ModelHandler) *BalanceLogBonusRechargeDao {
	return &BalanceLogBonusRechargeDao{
		group:    "default",
		table:    "balance_log_bonus_recharge",
		columns:  balanceLogBonusRechargeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogBonusRechargeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogBonusRechargeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogBonusRechargeDao) Columns() BalanceLogBonusRechargeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogBonusRechargeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogBonusRechargeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogBonusRechargeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
