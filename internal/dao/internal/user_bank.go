// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserBankDao is the data access object for the table user_bank.
type UserBankDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserBankColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserBankColumns defines and stores column names for the table user_bank.
type UserBankColumns struct {
	Id          string //
	SiteId      string // 站点ID
	UserId      string // 会员ID
	Type        string // 银行类型1=GCash,2=Maya
	BankName    string // 银行名称
	CardAccount string // 银行账户名
	CardNo      string // 银行卡号
	IsDefault   string // 默认地址。1=默认，0=非默认
	CreatedAt   string //
	UpdatedAt   string //
}

// userBankColumns holds the columns for the table user_bank.
var userBankColumns = UserBankColumns{
	Id:          "id",
	SiteId:      "site_id",
	UserId:      "user_id",
	Type:        "type",
	BankName:    "bank_name",
	CardAccount: "card_account",
	CardNo:      "card_no",
	IsDefault:   "is_default",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewUserBankDao creates and returns a new DAO object for table data access.
func NewUserBankDao(handlers ...gdb.ModelHandler) *UserBankDao {
	return &UserBankDao{
		group:    "default",
		table:    "user_bank",
		columns:  userBankColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserBankDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserBankDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserBankDao) Columns() UserBankColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserBankDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserBankDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserBankDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
