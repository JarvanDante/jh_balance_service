package mg

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"jh_balance_service/internal/logic/frontend/payment"
)

// MGPlatform MG888支付平台实现
type MGPlatform struct {
	gateway      string // 创建订单网关
	gatewayQuery string // 查询订单网关
	gatewayBank  string // 银行流水号匹配网关
}

// NewMG 创建MG平台实例
func NewMG() payment.PaymentPlatform {
	return &MGPlatform{
		gateway:      "http://54.179.49.179:8960/api/merchant/order",
		gatewayQuery: "http://54.179.49.179:8960/api/merchant/order",
		gatewayBank:  "http://54.179.49.179:8960/api/merchant/order/serial",
	}
}

// 注册到工厂
func init() {
	payment.GetFactory().Register("MG", func() payment.PaymentPlatform {
		return NewMG()
	})
}

// CreateOrder 创建充值订单
func (p *MGPlatform) CreateOrder(ctx context.Context, params *payment.OrderParams) (*payment.OrderResult, error) {
	g.Log().Infof(ctx, "MG CreateOrder 开始: tradeNo=%s, siteId=%d, userId=%d, money=%.2f",
		params.TradeNo, params.SiteId, params.UserId, params.Money)

	// 构建请求参数
	reqParams := make(map[string]string)
	reqParams["merchant_no"] = params.MerchantNo
	reqParams["notify_url"] = p.getNotifyUrl()
	reqParams["order_sn"] = params.TradeNo
	reqParams["order_type"] = p.convertOrderType(params.OrderType)
	reqParams["pay_type"] = p.getPayType(params.Gateway)
	reqParams["amount"] = fmt.Sprintf("%.2f", params.Money)
	reqParams["user_id"] = fmt.Sprintf("%d", params.UserId)
	reqParams["user_card_name"] = params.UserCardName

	// 如果是提现，添加银行卡号
	if params.OrderType == "withdraw" {
		reqParams["user_card_no"] = params.UserCardNo
	}

	// 生成签名
	sign := p.generateSign(reqParams, params.Md5Key)
	reqParams["sign"] = sign

	g.Log().Infof(ctx, "MG CreateOrder 请求参数: %+v", reqParams)

	// 发送HTTP请求
	resp, err := p.postRequest(ctx, p.gateway, reqParams)
	if err != nil {
		g.Log().Errorf(ctx, "MG CreateOrder 请求失败: %v", err)
		return nil, fmt.Errorf("请求支付网关失败: %v", err)
	}

	g.Log().Infof(ctx, "MG CreateOrder 响应: %s", resp)

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		g.Log().Errorf(ctx, "MG CreateOrder 解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查响应状态
	code, _ := result["code"].(float64)
	if code != 200 {
		message, _ := result["message"].(string)
		return nil, fmt.Errorf("创建订单失败: %s", message)
	}

	// 构建返回结果
	payType := reqParams["pay_type"]
	orderResult := &payment.OrderResult{
		Code:          int(code),
		OrderSn:       params.TradeNo,
		SysTradeNo:    p.getString(result, "sys_trade_no"),
		ChannelCardNo: p.getString(result, "channel_card_no"),
		ChannelName:   strings.ToUpper(payType),
		ImageUrl:      p.getPayTypeImageUrl(payType),
		Result:        "ok",
	}

	g.Log().Infof(ctx, "MG CreateOrder 成功: %+v", orderResult)
	return orderResult, nil
}

// QueryOrder 查询订单状态
func (p *MGPlatform) QueryOrder(ctx context.Context, params *payment.QueryParams) (*payment.QueryResult, error) {
	g.Log().Infof(ctx, "MG QueryOrder 开始: orderSn=%s", params.OrderSn)

	// 构建查询参数
	queryParams := make(map[string]string)
	queryParams["merchant_no"] = params.MerchantNo
	queryParams["order_sn"] = params.OrderSn

	// 生成签名
	sign := p.generateSign(queryParams, params.Md5Key)
	queryParams["sign"] = sign

	// 构建查询URL
	queryUrl := fmt.Sprintf("%s?%s", p.gatewayQuery, p.buildQueryString(queryParams))

	g.Log().Infof(ctx, "MG QueryOrder URL: %s", queryUrl)

	// 发送GET请求
	resp, err := p.getRequest(ctx, queryUrl)
	if err != nil {
		g.Log().Errorf(ctx, "MG QueryOrder 请求失败: %v", err)
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}

	g.Log().Infof(ctx, "MG QueryOrder 响应: %s", resp)

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		g.Log().Errorf(ctx, "MG QueryOrder 解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	code, _ := result["code"].(float64)
	queryResult := &payment.QueryResult{
		Code:    int(code),
		Message: p.getString(result, "message"),
	}

	if code == 200 {
		queryResult.TradeStatus = "success"
	} else {
		queryResult.TradeStatus = "failed"
	}

	return queryResult, nil
}

// HandleCallback 处理异步回调
func (p *MGPlatform) HandleCallback(ctx context.Context, data map[string]interface{}) (*payment.CallbackResult, error) {
	g.Log().Infof(ctx, "MG HandleCallback 收到回调: %+v", data)

	// 提取回调数据
	tradeStatus := p.getStringFromMap(data, "trade_status")
	orderSn := p.getStringFromMap(data, "order_sn")
	actualAmount := p.getFloatFromMap(data, "actual_amount")

	result := &payment.CallbackResult{
		Success:      tradeStatus == "success",
		TradeNo:      orderSn,
		TradeStatus:  tradeStatus,
		ActualAmount: actualAmount,
	}

	if tradeStatus != "success" {
		result.Message = fmt.Sprintf("支付失败，状态: %s", tradeStatus)
		return result, nil
	}

	result.Message = "支付成功"
	return result, nil
}

// VerifyCallback 是供外部（如资金回调处理逻辑）直接调用的回调验签入口，
// 内部复用与下单一致的 MD5 签名算法。
func VerifyCallback(ctx context.Context, params map[string]string, sign, secretKey string) bool {
	p := &MGPlatform{}
	return p.VerifyCallbackSign(ctx, params, sign, secretKey)
}

// VerifyCallbackSign 校验回调签名。
//
// 使用与下单/查单一致的签名算法（generateSign：过滤空值 -> key 升序 -> key=value& 拼接
// -> 追加 &key=<secretKey> -> MD5 大写）对回调参数重新计算签名，并与回调携带的 sign 比对。
//
// secretKey 为对应站点支付渠道的 MD5 密钥（site_payment.md5_key），由调用方从订单归属的
// 支付配置中读取传入。
//
// params 为参与签名的回调原始参数（不含 sign 本身）。
//
// 注意（TODO）：MG 平台回调「参与签名的字段集合」需以对方接入文档为准确认。当前实现按
// generateSign 的通用规则对「除 sign 外的全部回调参数」做签名校验；若上游网关在转发时丢弃了
// 部分回调字段（当前 gRPC DepositCallbackReq 仅透传 trade_status/order_sn/actual_amount/sign），
// 则应在网关侧补齐原始回调参数后再校验，或改为按文档指定的字段子集校验。
// 在字段集合未最终确认前，此函数遵循 fail-closed：签名不匹配一律拒绝。
func (p *MGPlatform) VerifyCallbackSign(ctx context.Context, params map[string]string, sign, secretKey string) bool {
	if secretKey == "" {
		g.Log().Errorf(ctx, "MG VerifyCallbackSign 密钥为空，拒绝回调")
		return false
	}
	if sign == "" {
		g.Log().Errorf(ctx, "MG VerifyCallbackSign 回调未携带签名，拒绝回调")
		return false
	}

	// 拷贝参数并剔除 sign 字段（sign 不参与签名计算）
	toSign := make(map[string]string, len(params))
	for k, v := range params {
		if k == "sign" {
			continue
		}
		toSign[k] = v
	}

	expected := p.generateSign(toSign, secretKey)
	if !strings.EqualFold(expected, sign) {
		g.Log().Errorf(ctx, "MG VerifyCallbackSign 签名不匹配: expected=%s, got=%s, params=%+v", expected, sign, toSign)
		return false
	}
	return true
}

// MatchOrder 支付匹配(提交银行流水号)
func (p *MGPlatform) MatchOrder(ctx context.Context, params *payment.MatchParams) error {
	g.Log().Infof(ctx, "MG MatchOrder 开始: sysTradeNo=%s, bankSerial=%s", params.SysTradeNo, params.BankSerial)

	// 构建请求参数
	reqParams := make(map[string]string)
	reqParams["merchant_no"] = params.MerchantNo
	reqParams["sys_trade_no"] = params.SysTradeNo
	reqParams["bank_serial"] = params.BankSerial

	// 生成签名
	sign := p.generateSign(reqParams, params.Md5Key)
	reqParams["sign"] = sign

	// 发送HTTP请求
	resp, err := p.postRequest(ctx, p.gatewayBank, reqParams)
	if err != nil {
		g.Log().Errorf(ctx, "MG MatchOrder 请求失败: %v", err)
		return fmt.Errorf("支付匹配失败: %v", err)
	}

	g.Log().Infof(ctx, "MG MatchOrder 响应: %s", resp)

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	code, _ := result["code"].(float64)
	if code != 200 {
		return fmt.Errorf("支付匹配失败")
	}

	return nil
}

// generateSign 生成MD5签名
func (p *MGPlatform) generateSign(params map[string]string, secretKey string) string {
	// 1. 过滤空值
	filtered := make(map[string]string)
	for k, v := range params {
		if v != "" {
			filtered[k] = v
		}
	}

	// 2. 按key排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 3. 拼接成 key=value&key2=value2
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}
	queryString := strings.Join(parts, "&")

	// 4. 加上密钥
	stringToSign := queryString + "&key=" + secretKey

	// 5. MD5签名并转大写
	hash := md5.Sum([]byte(stringToSign))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// convertOrderType 转换订单类型
func (p *MGPlatform) convertOrderType(orderType string) string {
	switch orderType {
	case "deposit":
		return "recharge"
	case "withdraw":
		return "withdrawal"
	default:
		return orderType
	}
}

// getPayType 获取支付类型
func (p *MGPlatform) getPayType(gateway int) string {
	// 1=银行卡, 2=微信, 3=支付宝
	switch gateway {
	case 1:
		return "kbz"
	case 2:
		return "wave"
	default:
		return "kbz"
	}
}

// getPayTypeImageUrl 获取支付引导图
func (p *MGPlatform) getPayTypeImageUrl(payType string) string {
	switch payType {
	case "kbz":
		return "https://img.okszckoo.com/upload/images/202502/c18cd50f-1025-48dd-af55-c5d49a819cc5.jpg"
	case "wave":
		return "https://img.okszckoo.com/upload/images/202502/95054348-4363-41d6-87b5-0a48afbb39c3.jpg"
	default:
		return ""
	}
}

// getNotifyUrl 获取异步通知地址
func (p *MGPlatform) getNotifyUrl() string {
	// TODO: 从配置文件读取
	return "https://jjj2.com/payment/asyn/mg"
}

// postRequest 发送POST请求
func (p *MGPlatform) postRequest(ctx context.Context, url string, params map[string]string) (string, error) {
	// 构建表单数据
	formData := make(urlpkg.Values)
	for k, v := range params {
		formData.Set(k, v)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// getRequest 发送GET请求
func (p *MGPlatform) getRequest(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// buildQueryString 构建查询字符串
func (p *MGPlatform) buildQueryString(params map[string]string) string {
	values := make(urlpkg.Values)
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// getString 从map中获取字符串
func (p *MGPlatform) getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getStringFromMap 从map中获取字符串
func (p *MGPlatform) getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getFloatFromMap 从map中获取浮点数
func (p *MGPlatform) getFloatFromMap(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case string:
			var f float64
			fmt.Sscanf(val, "%f", &f)
			return f
		}
	}
	return 0
}
