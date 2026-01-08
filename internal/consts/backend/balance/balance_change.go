package balance

// 账变类型
const (
	ChangeTypeIncome = 1 // 入款
	ChangeTypePay    = 2 // 出款
)

// 交易类型
const (
	TradeTypeTransferIn      = 1  // 转入
	TradeTypeRecharge        = 2  // 充值
	TradeTypeBonus           = 3  // 红利
	TradeTypeRebate          = 4  // 返水
	TradeTypeTransferOut     = 5  // 转出
	TradeTypeWithdrawSuccess = 6  // 提现成功
	TradeTypeWithdrawRefuse  = 7  // 提现返回
	TradeTypeWithdrawFrozen  = 8  // 提现冻结
	TradeTypeManualAdd       = 9  // 丢失补款
	TradeTypeManualCut       = 10 // 多出扣款
)

// 状态
const (
	StatusPending = 1 // 待处理
	StatusSuccess = 2 // 成功
	StatusFail    = 3 // 失败
)

// 交易类型名称映射
var TradeTypeNames = map[int]string{
	TradeTypeTransferIn:      "转入",
	TradeTypeRecharge:        "充值",
	TradeTypeBonus:           "红利",
	TradeTypeRebate:          "返水",
	TradeTypeTransferOut:     "转出",
	TradeTypeWithdrawSuccess: "提现成功",
	TradeTypeWithdrawRefuse:  "提现返回",
	TradeTypeWithdrawFrozen:  "提现冻结",
	TradeTypeManualAdd:       "丢失补款",
	TradeTypeManualCut:       "多出扣款",
}

// 账变类型名称映射
var ChangeTypeNames = map[int]string{
	ChangeTypeIncome: "入款",
	ChangeTypePay:    "出款",
}

// 状态名称映射
var StatusNames = map[int]string{
	StatusPending: "待处理",
	StatusSuccess: "成功",
	StatusFail:    "失败",
}

// 入款交易类型列表
var IncomeTradeTypes = []int{
	TradeTypeTransferIn,
	TradeTypeRecharge,
	TradeTypeBonus,
	TradeTypeRebate,
	TradeTypeWithdrawRefuse,
	TradeTypeManualAdd,
}

// 出款交易类型列表
var PayTradeTypes = []int{
	TradeTypeTransferOut,
	TradeTypeWithdrawSuccess,
	TradeTypeWithdrawFrozen,
	TradeTypeManualCut,
}

// GetTradeTypeName 获取交易类型名称
func GetTradeTypeName(tradeType int) string {
	if name, ok := TradeTypeNames[tradeType]; ok {
		return name
	}
	return "未知"
}

// GetChangeTypeName 获取账变类型名称
func GetChangeTypeName(changeType int) string {
	if name, ok := ChangeTypeNames[changeType]; ok {
		return name
	}
	return "未知"
}

// GetStatusName 获取状态名称
func GetStatusName(status int) string {
	if name, ok := StatusNames[status]; ok {
		return name
	}
	return "未知"
}

// GetChangeTypeByTradeType 根据交易类型获取账变类型
func GetChangeTypeByTradeType(tradeType int) int {
	for _, t := range IncomeTradeTypes {
		if t == tradeType {
			return ChangeTypeIncome
		}
	}
	for _, t := range PayTradeTypes {
		if t == tradeType {
			return ChangeTypePay
		}
	}
	return 0
}
