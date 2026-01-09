package balance

// 账变类型常量（手动操作相关）
const (
	TYPE_ADD_RECHARGE      = 1  // 加款 - 充值
	TYPE_ADD_BONUS         = 2  // 加款 - 红利
	TYPE_ADD_REBATE        = 3  // 加款 - 返水
	TYPE_ADD_DIRECT        = 4  // 丢失补款（不入充值记录）
	TYPE_CUT_WITHDRAW      = 5  // 扣款 - 提现
	TYPE_CUT_DIRECT        = 6  // 多出扣款（不入提现记录）
	TYPE_GAME_ADD_DIRECT   = 7  // 第三方加款（不入账变）
	TYPE_GAME_CUT_DIRECT   = 8  // 第三方扣款（不入账变）
	TYPE_GAME_TRANSFER_IN  = 9  // 第三方账户 -> 中心账户
	TYPE_GAME_TRANSFER_OUT = 10 // 中心账户 -> 第三方账户
)

// 账变类型名称映射
var TypeNames = map[int]string{
	TYPE_ADD_RECHARGE:      "加款-充值",
	TYPE_ADD_BONUS:         "加款-红利",
	TYPE_ADD_REBATE:        "加款-返水",
	TYPE_ADD_DIRECT:        "丢失补款(不入充值记录)",
	TYPE_CUT_WITHDRAW:      "扣款-提现",
	TYPE_CUT_DIRECT:        "多出扣款(不入提现记录)",
	TYPE_GAME_ADD_DIRECT:   "第三方加款",
	TYPE_GAME_CUT_DIRECT:   "第三方扣款",
	TYPE_GAME_TRANSFER_IN:  "第三方账户 -> 中心账户",
	TYPE_GAME_TRANSFER_OUT: "中心账户 -> 第三方账户",
}

// 加款类型列表
var AddTypes = []int{
	TYPE_ADD_RECHARGE,
	TYPE_ADD_BONUS,
	TYPE_ADD_REBATE,
	TYPE_ADD_DIRECT,
	TYPE_GAME_ADD_DIRECT,
	TYPE_GAME_TRANSFER_IN,
}

// 扣款类型列表
var CutTypes = []int{
	TYPE_CUT_WITHDRAW,
	TYPE_CUT_DIRECT,
	TYPE_GAME_CUT_DIRECT,
	TYPE_GAME_TRANSFER_OUT,
}

// GetTypeName 获取账变类型名称
func GetTypeName(typeCode int) string {
	if name, ok := TypeNames[typeCode]; ok {
		return name
	}
	return "未知"
}

// IsAddType 判断是否为加款类型
func IsAddType(typeCode int) bool {
	for _, t := range AddTypes {
		if t == typeCode {
			return true
		}
	}
	return false
}

// IsCutType 判断是否为扣款类型
func IsCutType(typeCode int) bool {
	for _, t := range CutTypes {
		if t == typeCode {
			return true
		}
	}
	return false
}

// GetBalanceChangeTradeType 根据手动操作类型获取账变表对应的交易类型
// 参考 b_service 中的 BalanceLogManual::_getBalanceChangeTradeType 方法
func GetBalanceChangeTradeType(tradeType int) int {
	switch tradeType {
	case TYPE_ADD_RECHARGE: // 加款 - 充值
		return TradeTypeRecharge
	case TYPE_ADD_BONUS: // 加款 - 红利
		return TradeTypeBonus
	case TYPE_ADD_REBATE: // 加款 - 返水
		return TradeTypeRebate
	case TYPE_ADD_DIRECT: // 丢失补款（不入充值记录）
		return TradeTypeManualAdd
	case TYPE_CUT_WITHDRAW: // 扣款 - 提现
		return TradeTypeWithdrawSuccess
	case TYPE_CUT_DIRECT: // 多出扣款（不入提现记录）
		return TradeTypeManualCut
	case TYPE_GAME_TRANSFER_IN: // 第三方账户 -> 中心账户
		return TradeTypeTransferIn
	case TYPE_GAME_TRANSFER_OUT: // 中心账户 -> 第三方账户
		return TradeTypeTransferOut
	default:
		return 0
	}
}
