package util

import (
	"context"
	"crypto/rand"
	"fmt"
	balanceConsts "jh_balance_service/internal/consts/backend/balance"
	"jh_balance_service/internal/middleware"
	"math/big"
	randNew "math/rand"
	"strings"
	"time"
)

var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// 创建独立随机源（推荐方式）
var r = randNew.New(randNew.NewSource(time.Now().UnixNano()))

//GenerateTradeNo
/**
 * @desc：生成交易流水号（全站唯一）因为第三方游戏API的原因，网站订单号生成字母数字组合；且数字长度大于等于17个
 * @param tradeType
 * @return string
 * @author : Carson
 */
func GenerateTradeNo(ctx context.Context, tradeType int) string {
	prefix := getTradePrefix(tradeType)

	siteCode, _ := middleware.GetSiteCodeFromContext(ctx)
	// 14位时间戳
	timeStr := time.Now().Format("20060102150405")

	// 纳秒后6位（避免同秒冲突）
	nanoPart := time.Now().UnixNano() % 1000000

	// 4位随机字母
	randomLetters := make([]rune, 4)
	for i := 0; i < 4; i++ {
		randomLetters[i] = letters[r.Intn(len(letters))]
	}

	orderNo := fmt.Sprintf("%s%s%06d%s%s",
		prefix,
		timeStr,
		nanoPart,
		string(randomLetters),
		siteCode,
	)

	return strings.ToUpper(orderNo)
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
