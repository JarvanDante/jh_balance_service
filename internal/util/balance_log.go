package util

import (
	"context"
	"crypto/rand"
	"fmt"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"math/big"
	"strings"
	"time"
)

//GenerateTradeNo
/**
 * @desc：生成交易流水号（全站唯一）因为第三方游戏API的原因，网站订单号生成字母数字组合；且数字长度大于等于17个
 * @param tradeType
 * @return string
 * @author : Carson
 */
func GenerateTradeNo(ctx context.Context, tradeType int) string {
	prefix := getTradePrefix(tradeType)

	// 时间戳：yyyyMMddHHmmss + 毫秒(3位) = 17位
	now := time.Now()
	timePart := now.Format("20060102150405") + fmt.Sprintf("%03d", now.Nanosecond()/1e6)

	// 随机数：1~999999，固定 6 位
	randomPart := randIntString(1, 999999, 6)

	tradeNo := prefix + timePart + randomPart
	return strings.ToUpper(tradeNo)
}

func getTradePrefix(tradeType int) string {
	switch tradeType {
	case balanceConsts.LOG_TYPE_PAYMENT:
		return "ZF"
	case balanceConsts.LOG_TYPE_TRANSFER:
		return "ZZ"
	case balanceConsts.LOG_TYPE_RECHARGE_MANUAL:
		return "CZ"
	case balanceConsts.LOG_TYPE_WITHDRAW, balanceConsts.LOG_TYPE_WITHDRAW_MANUAL:
		return "T"
	case balanceConsts.LOG_TYPE_MANUAL_ADD:
		return "JK"
	case balanceConsts.LOG_TYPE_MANUAL_CUT:
		return "KK"
	case balanceConsts.LOG_TYPE_BONUS:
		return "HL"
	case balanceConsts.LOG_TYPE_REBATE:
		return "FS"
	case balanceConsts.LOG_TYPE_TRANSFER_IN:
		return "ZR"
	case balanceConsts.LOG_TYPE_TRANSFER_OUT:
		return "ZC"
	case balanceConsts.LOG_TYPE_GAME_ADD, balanceConsts.LOG_TYPE_GAME_CUT:
		return "YX"
	case balanceConsts.LOG_TYPE_POINTS:
		return "JF"
	default:
		return "QT"
	}
}

// randIntString 生成 [min, max] 范围随机数，左侧补0到固定长度 width
// 使用 crypto/rand，避免 math/rand 在高并发下撞号风险
func randIntString(min, max int64, width int) string {
	if min > max {
		min, max = max, min
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		// 极少情况下 crypto/rand 失败，fallback 用时间兜底
		v := time.Now().UnixNano() % (max - min + 1)
		return fmt.Sprintf("%0*d", width, v+min)
	}
	v := nBig.Int64() + min
	return fmt.Sprintf("%0*d", width, v)
}
