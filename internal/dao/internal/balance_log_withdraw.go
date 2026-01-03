// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogWithdrawDao is the data access object for the table balance_log_withdraw.
type BalanceLogWithdrawDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  BalanceLogWithdrawColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// BalanceLogWithdrawColumns defines and stores column names for the table balance_log_withdraw.
type BalanceLogWithdrawColumns struct {
	Id          string //
	SiteId      string // 站点ID
	UserId      string // 会员ID
	UserLevelId string // 会员层级ID
	Username    string // 会员用户名
	TradeNo     string // 流水号
	Money       string // 充值金额
	Fee         string // 手续费
	BankName    string // 提现银行
	CardAccount string // 银行户名
	CardNo      string // 卡号
	DepositBank string // 开户行
	Status      string // 状态. 1=待审核；2=成功；3=失败
	AdminId     string // 后台用户ID
	Remark      string // 备注
	CreatedAt   string //
	UpdatedAt   string //
}

// balanceLogWithdrawColumns holds the columns for the table balance_log_withdraw.
var balanceLogWithdrawColumns = BalanceLogWithdrawColumns{
	Id:          "id",
	SiteId:      "site_id",
	UserId:      "user_id",
	UserLevelId: "user_level_id",
	Username:    "username",
	TradeNo:     "trade_no",
	Money:       "money",
	Fee:         "fee",
	BankName:    "bank_name",
	CardAccount: "card_account",
	CardNo:      "card_no",
	DepositBank: "deposit_bank",
	Status:      "status",
	AdminId:     "admin_id",
	Remark:      "remark",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewBalanceLogWithdrawDao creates and returns a new DAO object for table data access.
func NewBalanceLogWithdrawDao(handlers ...gdb.ModelHandler) *BalanceLogWithdrawDao {
	return &BalanceLogWithdrawDao{
		group:    "default",
		table:    "balance_log_withdraw",
		columns:  balanceLogWithdrawColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogWithdrawDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogWithdrawDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogWithdrawDao) Columns() BalanceLogWithdrawColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogWithdrawDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogWithdrawDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogWithdrawDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
