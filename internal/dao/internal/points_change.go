// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PointsChangeDao is the data access object for the table points_change.
type PointsChangeDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  PointsChangeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// PointsChangeColumns defines and stores column names for the table points_change.
type PointsChangeColumns struct {
	Id           string //
	SiteId       string // 站点ID
	ChangeType   string // 变化类型。1=增加；0=减少
	TradeType    string // 类型。1=签到;2=充值;3=投注
	UserId       string // 会员ID
	Username     string // 会员用户名
	PointsOld    string // 原积分
	PointsChange string // 积分变动
	PointsNew    string // 现积分
	Status       string // 状态。1=待处理；2=成功；3=失败
	AdminId      string // 后台员工ID
	Remark       string //
	CreatedAt    string //
	UpdatedAt    string //
	TradeNo      string //
}

// pointsChangeColumns holds the columns for the table points_change.
var pointsChangeColumns = PointsChangeColumns{
	Id:           "id",
	SiteId:       "site_id",
	ChangeType:   "change_type",
	TradeType:    "trade_type",
	UserId:       "user_id",
	Username:     "username",
	PointsOld:    "points_old",
	PointsChange: "points_change",
	PointsNew:    "points_new",
	Status:       "status",
	AdminId:      "admin_id",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	TradeNo:      "trade_no",
}

// NewPointsChangeDao creates and returns a new DAO object for table data access.
func NewPointsChangeDao(handlers ...gdb.ModelHandler) *PointsChangeDao {
	return &PointsChangeDao{
		group:    "default",
		table:    "points_change",
		columns:  pointsChangeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PointsChangeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PointsChangeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PointsChangeDao) Columns() PointsChangeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PointsChangeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PointsChangeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PointsChangeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
