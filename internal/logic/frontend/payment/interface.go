package payment

import (
	"context"
)

// PaymentPlatform 支付平台接口
// 所有支付平台都需要实现这个接口
type PaymentPlatform interface {
	// CreateOrder 创建充值订单
	CreateOrder(ctx context.Context, params *OrderParams) (*OrderResult, error)

	// QueryOrder 查询订单状态
	QueryOrder(ctx context.Context, params *QueryParams) (*QueryResult, error)

	// HandleCallback 处理异步回调
	HandleCallback(ctx context.Context, data map[string]interface{}) (*CallbackResult, error)

	// MatchOrder 支付匹配(提交银行流水号)
	MatchOrder(ctx context.Context, params *MatchParams) error
}

// OrderParams 创建订单参数
type OrderParams struct {
	SiteId       int     // 站点ID
	ChannelId    int     // 渠道ID
	UserId       int     // 用户ID
	Username     string  // 用户名
	PaymentId    int     // 支付方式ID
	AccountId    int     // 支付账号ID
	MerchantNo   string  // 商户号
	Md5Key       string  // MD5密钥
	PrivateKey   string  // 私钥
	PublicKey    string  // 公钥
	Domain       string  // 域名
	TradeNo      string  // 交易流水号
	Money        float64 // 金额
	OrderType    string  // 订单类型: deposit=充值, withdraw=提现
	BankCode     string  // 银行代码(可选)
	Mobile       int     // 是否手机端: 1=是, 0=否
	UserCardNo   string  // 用户银行卡号(提现必填)
	UserCardName string  // 用户银行卡姓名
	Gateway      int     // 支付网关类型
	Nonce        string  // 防重放随机数
}

// OrderResult 创建订单结果
type OrderResult struct {
	Code          int    `json:"code"`            // 状态码
	OrderSn       string `json:"order_sn"`        // 系统订单号
	SysTradeNo    string `json:"sys_trade_no"`    // 平台交易流水号
	ChannelCardNo string `json:"channel_card_no"` // 收款卡号
	ChannelName   string `json:"channel_name"`    // 渠道名称
	ImageUrl      string `json:"image_url"`       // 支付引导图
	Result        string `json:"result"`          // 返回信息
}

// QueryParams 查询订单参数
type QueryParams struct {
	MerchantNo string // 商户号
	OrderSn    string // 订单号
	Md5Key     string // MD5密钥
}

// QueryResult 查询订单结果
type QueryResult struct {
	Code        int     // 状态码
	TradeStatus string  // 交易状态: wait=处理中, success=成功, failed=失败, cancel=取消
	Amount      float64 // 金额
	Message     string  // 消息
}

// CallbackResult 回调处理结果
type CallbackResult struct {
	Success      bool    // 是否成功
	TradeNo      string  // 订单号
	TradeStatus  string  // 交易状态
	ActualAmount float64 // 实际支付金额
	Message      string  // 消息
}

// MatchParams 支付匹配参数
type MatchParams struct {
	MerchantNo string // 商户号
	SysTradeNo string // 平台流水号
	BankSerial string // 银行流水号
	Md5Key     string // MD5密钥
}
