// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceWheelRewardDao is the data access object for the table balance_wheel_reward.
type BalanceWheelRewardDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  BalanceWheelRewardColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// BalanceWheelRewardColumns defines and stores column names for the table balance_wheel_reward.
type BalanceWheelRewardColumns struct {
	Id           string //
	SiteId       string // 站点ID
	UserId       string // 用户ID
	Date         string // 抽奖日期(yyyy-mm-dd)
	SegmentNo    string // 命中扇区(1~12)
	PrizeName    string // 奖项名称
	RewardAmount string // 中奖金额
	TradeNo      string // 发放流水号/订单号
	CreatedAt    string //
	UpdatedAt    string //
}

// balanceWheelRewardColumns holds the columns for the table balance_wheel_reward.
var balanceWheelRewardColumns = BalanceWheelRewardColumns{
	Id:           "id",
	SiteId:       "site_id",
	UserId:       "user_id",
	Date:         "date",
	SegmentNo:    "segment_no",
	PrizeName:    "prize_name",
	RewardAmount: "reward_amount",
	TradeNo:      "trade_no",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewBalanceWheelRewardDao creates and returns a new DAO object for table data access.
func NewBalanceWheelRewardDao(handlers ...gdb.ModelHandler) *BalanceWheelRewardDao {
	return &BalanceWheelRewardDao{
		group:    "default",
		table:    "balance_wheel_reward",
		columns:  balanceWheelRewardColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceWheelRewardDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceWheelRewardDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceWheelRewardDao) Columns() BalanceWheelRewardColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceWheelRewardDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceWheelRewardDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceWheelRewardDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
