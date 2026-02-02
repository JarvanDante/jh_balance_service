// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogManualDao is the data access object for the table balance_log_manual.
type BalanceLogManualDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  BalanceLogManualColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// BalanceLogManualColumns defines and stores column names for the table balance_log_manual.
type BalanceLogManualColumns struct {
	Id         string //
	SiteId     string // 站点ID
	TradeType  string // 交易类型。1=加款-充值；2=加款-红利；3=加款-返水；4=补款（不入充值记录），5=扣款 - 提现，6=扣款（不入提现记录）7=第三方加款（不入账变）,8=第三方扣款（不入账变）;9=第三方账户 -> 中心账户;10=中心账户 -> 第三方账户
	UserId     string // 用户ID
	Username   string // 会员用户名
	GameId     string // 游戏ID
	TradeNo    string // 流水号
	Money      string // 金额
	Fee        string // 手续费
	WaterTimes string // 流水倍数
	Status     string // 状态0=初始化，1=待处理；2=处理中；3=成功；4=失败
	AdminId    string // 后台管理员ID
	Remark     string // 备注
	CreatedAt  string //
	UpdatedAt  string //
}

// balanceLogManualColumns holds the columns for the table balance_log_manual.
var balanceLogManualColumns = BalanceLogManualColumns{
	Id:         "id",
	SiteId:     "site_id",
	TradeType:  "trade_type",
	UserId:     "user_id",
	Username:   "username",
	GameId:     "game_id",
	TradeNo:    "trade_no",
	Money:      "money",
	Fee:        "fee",
	WaterTimes: "water_times",
	Status:     "status",
	AdminId:    "admin_id",
	Remark:     "remark",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewBalanceLogManualDao creates and returns a new DAO object for table data access.
func NewBalanceLogManualDao(handlers ...gdb.ModelHandler) *BalanceLogManualDao {
	return &BalanceLogManualDao{
		group:    "default",
		table:    "balance_log_manual",
		columns:  balanceLogManualColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogManualDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogManualDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogManualDao) Columns() BalanceLogManualColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogManualDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogManualDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogManualDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
