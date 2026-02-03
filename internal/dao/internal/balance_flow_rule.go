// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceFlowRuleDao is the data access object for the table balance_flow_rule.
type BalanceFlowRuleDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  BalanceFlowRuleColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// BalanceFlowRuleColumns defines and stores column names for the table balance_flow_rule.
type BalanceFlowRuleColumns struct {
	Id           string // 主键ID
	SiteId       string // 站点ID
	RuleName     string // 规则名称
	RuleType     string // 规则类型: 1=充值流水, 2=红利流水, 3=活动流水, 4=人工加款流水
	FlowMultiple string // 默认流水倍数
	ExpireDays   string // 有效天数（0表示永久）
	Status       string // 状态: 0=禁用, 1=启用
	Sort         string // 排序
	Remark       string // 备注说明
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
}

// balanceFlowRuleColumns holds the columns for the table balance_flow_rule.
var balanceFlowRuleColumns = BalanceFlowRuleColumns{
	Id:           "id",
	SiteId:       "site_id",
	RuleName:     "rule_name",
	RuleType:     "rule_type",
	FlowMultiple: "flow_multiple",
	ExpireDays:   "expire_days",
	Status:       "status",
	Sort:         "sort",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewBalanceFlowRuleDao creates and returns a new DAO object for table data access.
func NewBalanceFlowRuleDao(handlers ...gdb.ModelHandler) *BalanceFlowRuleDao {
	return &BalanceFlowRuleDao{
		group:    "default",
		table:    "balance_flow_rule",
		columns:  balanceFlowRuleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceFlowRuleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceFlowRuleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceFlowRuleDao) Columns() BalanceFlowRuleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceFlowRuleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceFlowRuleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceFlowRuleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
