// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogTransferDao is the data access object for the table balance_log_transfer.
type BalanceLogTransferDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  BalanceLogTransferColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// BalanceLogTransferColumns defines and stores column names for the table balance_log_transfer.
type BalanceLogTransferColumns struct {
	Id                 string //
	SiteId             string // 站点ID
	UserId             string // 用户ID
	Username           string // 会员用户名
	ActivityRechargeId string // 充值活动ID
	BankType           string // 转账汇款类型
	AccountId          string // 转账汇款账号ID
	TradeNo            string // 流水号
	ThirdTradeNo       string // 第三方订单号
	Money              string // 充值金额
	Fee                string // 手续费
	Status             string // 状态0=初始化，1=待处理；2=处理中；3=成功；4=失败
	TransferAccount    string // 汇款账户名
	Code               string // 随机码。其实就是会员用户名
	AdminId            string // 操作人id
	Remark             string // 备注
	CreatedAt          string //
	UpdatedAt          string //
}

// balanceLogTransferColumns holds the columns for the table balance_log_transfer.
var balanceLogTransferColumns = BalanceLogTransferColumns{
	Id:                 "id",
	SiteId:             "site_id",
	UserId:             "user_id",
	Username:           "username",
	ActivityRechargeId: "activity_recharge_id",
	BankType:           "bank_type",
	AccountId:          "account_id",
	TradeNo:            "trade_no",
	ThirdTradeNo:       "third_trade_no",
	Money:              "money",
	Fee:                "fee",
	Status:             "status",
	TransferAccount:    "transfer_account",
	Code:               "code",
	AdminId:            "admin_id",
	Remark:             "remark",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewBalanceLogTransferDao creates and returns a new DAO object for table data access.
func NewBalanceLogTransferDao(handlers ...gdb.ModelHandler) *BalanceLogTransferDao {
	return &BalanceLogTransferDao{
		group:    "default",
		table:    "balance_log_transfer",
		columns:  balanceLogTransferColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogTransferDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogTransferDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogTransferDao) Columns() BalanceLogTransferColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogTransferDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogTransferDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogTransferDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
