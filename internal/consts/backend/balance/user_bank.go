package balance

// 银行卡类型常量
const (
	BANK_GCASH = 1 // GCash
	BANK_MAYA  = 2 // Maya
)

// 银行卡类型名称映射
var BankTypeNames = map[int]string{
	BANK_GCASH: "GCash",
	BANK_MAYA:  "Maya",
}

// GetLogTypeName 获取银行卡类型名称
func GetBankTypeName(typeValue int) string {
	if name, ok := BankTypeNames[typeValue]; ok {
		return name
	}
	return "未知"
}
