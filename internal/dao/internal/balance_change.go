// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceChangeDao is the data access object for the table balance_change.
type BalanceChangeDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  BalanceChangeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// BalanceChangeColumns defines and stores column names for the table balance_change.
type BalanceChangeColumns struct {
	Id            string //
	SiteId        string // 站点ID
	TradeType     string // 交易类型。1=转入；2=充值；3=红利；4=返水；5=转出；6=提现成功；7=提现返回；8=提现冻结；9=丢失补款；10=多出扣款
	UserId        string // 用户ID
	Username      string // 用户名
	TradeNo       string // 流水号
	BalanceOld    string // 旧的可用余额
	Money         string // 变动金额
	BalanceNew    string // 新的可用余额
	BalanceFrozen string // 新的冻结余额
	Status        string // 账变状态0=初始化，1=待处理；2=处理中；3=成功；4=失败
	Remark        string //
	CreatedAt     string //
	UpdatedAt     string //
}

// balanceChangeColumns holds the columns for the table balance_change.
var balanceChangeColumns = BalanceChangeColumns{
	Id:            "id",
	SiteId:        "site_id",
	TradeType:     "trade_type",
	UserId:        "user_id",
	Username:      "username",
	TradeNo:       "trade_no",
	BalanceOld:    "balance_old",
	Money:         "money",
	BalanceNew:    "balance_new",
	BalanceFrozen: "balance_frozen",
	Status:        "status",
	Remark:        "remark",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewBalanceChangeDao creates and returns a new DAO object for table data access.
func NewBalanceChangeDao(handlers ...gdb.ModelHandler) *BalanceChangeDao {
	return &BalanceChangeDao{
		group:    "default",
		table:    "balance_change",
		columns:  balanceChangeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceChangeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceChangeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceChangeDao) Columns() BalanceChangeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceChangeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceChangeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceChangeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
