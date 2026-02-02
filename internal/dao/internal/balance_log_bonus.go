// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceLogBonusDao is the data access object for the table balance_log_bonus.
type BalanceLogBonusDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  BalanceLogBonusColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// BalanceLogBonusColumns defines and stores column names for the table balance_log_bonus.
type BalanceLogBonusColumns struct {
	Id                 string //
	SiteId             string // 站点ID
	TradeType          string // 交易类型。1=在线充值红利；2=转账汇款红利；3=生日红利；4=签到红利；
	UserId             string // 会员ID
	Username           string // 会员用户名
	TradeNo            string // 流水号
	ActivityRechargeId string // 充值活动ID
	Money              string // 金额
	Fee                string // 手续费
	Status             string // 状态0=初始化，1=待处理；2=处理中；3=成功；4=失败
	AdminId            string // 后台管理员ID。默认为0
	Remark             string // 备注
	CreatedAt          string //
	UpdatedAt          string //
}

// balanceLogBonusColumns holds the columns for the table balance_log_bonus.
var balanceLogBonusColumns = BalanceLogBonusColumns{
	Id:                 "id",
	SiteId:             "site_id",
	TradeType:          "trade_type",
	UserId:             "user_id",
	Username:           "username",
	TradeNo:            "trade_no",
	ActivityRechargeId: "activity_recharge_id",
	Money:              "money",
	Fee:                "fee",
	Status:             "status",
	AdminId:            "admin_id",
	Remark:             "remark",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewBalanceLogBonusDao creates and returns a new DAO object for table data access.
func NewBalanceLogBonusDao(handlers ...gdb.ModelHandler) *BalanceLogBonusDao {
	return &BalanceLogBonusDao{
		group:    "default",
		table:    "balance_log_bonus",
		columns:  balanceLogBonusColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceLogBonusDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceLogBonusDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceLogBonusDao) Columns() BalanceLogBonusColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceLogBonusDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceLogBonusDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceLogBonusDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
