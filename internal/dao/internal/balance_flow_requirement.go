// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceFlowRequirementDao is the data access object for the table balance_flow_requirement.
type BalanceFlowRequirementDao struct {
	table    string                        // table is the underlying table name of the DAO.
	group    string                        // group is the database configuration group name of the current DAO.
	columns  BalanceFlowRequirementColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler            // handlers for customized model modification.
}

// BalanceFlowRequirementColumns defines and stores column names for the table balance_flow_requirement.
type BalanceFlowRequirementColumns struct {
	Id            string // 主键ID
	SiteId        string // 站点ID
	UserId        string // 用户ID
	Username      string // 用户名
	SourceType    string // 来源类型: 1=充值, 2=红利, 3=返水, 4=活动, 5=人工加款
	SourceId      string // 来源ID（充值订单ID、红利ID等）
	TradeNo       string // 关联交易号
	Amount        string // 触发金额（充值金额、红利金额等）
	BonusAmount   string // 赠送金额（如果有）
	FlowMultiple  string // 流水倍数（如：1倍、3倍、5倍）
	RequiredFlow  string // 需要完成的流水金额 = (amount + bonus_amount) * flow_multiple
	CompletedFlow string // 已完成的流水金额
	RemainingFlow string // 剩余流水金额 = required_flow - completed_flow
	Status        string // 状态: 1=进行中, 2=已完成, 3=已过期, 4=已取消
	Progress      string // 完成进度百分比（0-100）
	StartTime     string // 开始时间
	ExpireTime    string // 过期时间（NULL表示永久有效）
	CompleteTime  string // 完成时间
	Remark        string // 备注说明
	AdminId       string // 操作管理员ID（人工操作时）
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
}

// balanceFlowRequirementColumns holds the columns for the table balance_flow_requirement.
var balanceFlowRequirementColumns = BalanceFlowRequirementColumns{
	Id:            "id",
	SiteId:        "site_id",
	UserId:        "user_id",
	Username:      "username",
	SourceType:    "source_type",
	SourceId:      "source_id",
	TradeNo:       "trade_no",
	Amount:        "amount",
	BonusAmount:   "bonus_amount",
	FlowMultiple:  "flow_multiple",
	RequiredFlow:  "required_flow",
	CompletedFlow: "completed_flow",
	RemainingFlow: "remaining_flow",
	Status:        "status",
	Progress:      "progress",
	StartTime:     "start_time",
	ExpireTime:    "expire_time",
	CompleteTime:  "complete_time",
	Remark:        "remark",
	AdminId:       "admin_id",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewBalanceFlowRequirementDao creates and returns a new DAO object for table data access.
func NewBalanceFlowRequirementDao(handlers ...gdb.ModelHandler) *BalanceFlowRequirementDao {
	return &BalanceFlowRequirementDao{
		group:    "default",
		table:    "balance_flow_requirement",
		columns:  balanceFlowRequirementColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceFlowRequirementDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceFlowRequirementDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceFlowRequirementDao) Columns() BalanceFlowRequirementColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceFlowRequirementDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceFlowRequirementDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceFlowRequirementDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
