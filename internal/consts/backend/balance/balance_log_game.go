package balance

// 操作类型常量
const (
	TRANSFER_OUT = 0 // 转出：从中心账户转出到游戏账户
	TRANSFER_IN  = 1 // 转入：从游戏账户转入到中心账户
)

// 操作类型名称映射
var LogGameTypeNames = map[int]string{
	TRANSFER_OUT: "从中心账户转出到游戏账户",
	TRANSFER_IN:  "从游戏账户转入到中心账户",
}

// GetLogGameTypeName 获取操作类型名称
func GetLogGameTypeName(typeValue int) string {
	if name, ok := LogGameTypeNames[typeValue]; ok {
		return name
	}
	return "未知"
}
