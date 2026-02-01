// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BalanceCallbackLogDao is the data access object for the table balance_callback_log.
type BalanceCallbackLogDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  BalanceCallbackLogColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// BalanceCallbackLogColumns defines and stores column names for the table balance_callback_log.
type BalanceCallbackLogColumns struct {
	Id            string // 主键ID
	SiteId        string // 站点ID
	CallbackType  string // 回调类型 1=充值 2=提现
	OrderSn       string // 订单号
	TradeStatus   string // 交易状态
	ActualAmount  string // 实际金额
	FailureReason string // 失败原因
	FailureStage  string // 失败阶段
	RequestData   string // 请求数据(JSON)
	ErrorMessage  string // 错误详情
	RetryCount    string // 重试次数
	IsResolved    string // 是否已解决 0=未解决 1=已解决
	ResolvedAt    string // 解决时间
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
}

// balanceCallbackLogColumns holds the columns for the table balance_callback_log.
var balanceCallbackLogColumns = BalanceCallbackLogColumns{
	Id:            "id",
	SiteId:        "site_id",
	CallbackType:  "callback_type",
	OrderSn:       "order_sn",
	TradeStatus:   "trade_status",
	ActualAmount:  "actual_amount",
	FailureReason: "failure_reason",
	FailureStage:  "failure_stage",
	RequestData:   "request_data",
	ErrorMessage:  "error_message",
	RetryCount:    "retry_count",
	IsResolved:    "is_resolved",
	ResolvedAt:    "resolved_at",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewBalanceCallbackLogDao creates and returns a new DAO object for table data access.
func NewBalanceCallbackLogDao(handlers ...gdb.ModelHandler) *BalanceCallbackLogDao {
	return &BalanceCallbackLogDao{
		group:    "default",
		table:    "balance_callback_log",
		columns:  balanceCallbackLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BalanceCallbackLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BalanceCallbackLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BalanceCallbackLogDao) Columns() BalanceCallbackLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BalanceCallbackLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BalanceCallbackLogDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BalanceCallbackLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
