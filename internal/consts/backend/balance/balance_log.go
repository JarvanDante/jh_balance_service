package balance

// 操作类型常量
const (
	LOG_TYPE_PAYMENT         = 1  // 在线充值
	LOG_TYPE_TRANSFER        = 2  // 转账汇款
	LOG_TYPE_RECHARGE_MANUAL = 3  // 手动充值
	LOG_TYPE_WITHDRAW        = 4  // 会员提现
	LOG_TYPE_WITHDRAW_MANUAL = 5  // 手动提现
	LOG_TYPE_MANUAL_ADD      = 6  // 手动加款
	LOG_TYPE_MANUAL_CUT      = 7  // 手动扣款
	LOG_TYPE_BONUS           = 8  // 红利
	LOG_TYPE_REBATE          = 9  // 返水
	LOG_TYPE_TRANSFER_IN     = 10 // 转入
	LOG_TYPE_TRANSFER_OUT    = 11 // 转出
	LOG_TYPE_GAME_ADD        = 12 // 游戏直接加款
	LOG_TYPE_GAME_CUT        = 13 // 游戏直接扣款
	LOG_TYPE_POINTS          = 14 // 积分
)

// 操作类型名称映射
var LogTypeNames = map[int]string{
	LOG_TYPE_PAYMENT:         "在线充值",
	LOG_TYPE_TRANSFER:        "转账汇款",
	LOG_TYPE_RECHARGE_MANUAL: "手动充值",
	LOG_TYPE_WITHDRAW:        "会员提现",
	LOG_TYPE_WITHDRAW_MANUAL: "手动提现",
	LOG_TYPE_MANUAL_ADD:      "手动加款",
	LOG_TYPE_MANUAL_CUT:      "手动扣款",
	LOG_TYPE_BONUS:           "红利",
	LOG_TYPE_REBATE:          "返水",
	LOG_TYPE_TRANSFER_IN:     "转入",
	LOG_TYPE_TRANSFER_OUT:    "转出",
	LOG_TYPE_GAME_ADD:        "游戏直接加款",
	LOG_TYPE_GAME_CUT:        "游戏直接扣款",
	LOG_TYPE_POINTS:          "积分",
}

// GetLogTypeName 获取操作类型名称
func GetLogTypeName(typeValue int) string {
	if name, ok := LogTypeNames[typeValue]; ok {
		return name
	}
	return "未知"
}
