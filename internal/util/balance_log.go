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

var balanceRemarkTranslations = map[string]string{
	"提现审核中":     "Withdrawal under review",
	"（已确认）成功":   "(Confirmed) Success",
	"（已确认）金额补偿": "(Confirmed) Balance compensation",
	"（已确认）转出失败": "(Confirmed) Transfer out failed",
	"异步确认在线支付":  "Payment confirmed successfully",
	"充值流水要求":    "Deposit turnover requirement",
	"邀请红利":      "Invitation bonus",
	"邀请充值红利":    "Invitation deposit bonus",
	"注册送彩金":     "Registration bonus",
	"充值送彩金":     "Deposit bonus",
	"洗码返水":      "Wash code rebate",
	"成功":        "Success",
	"失败":        "Failed",
	"在线充值":      "Online deposit",
	"转账汇款":      "Transfer",
	"手动充值":      "Manual deposit",
	"会员提现":      "Member withdrawal",
	"手动提现":      "Manual withdrawal",
	"手动加款":      "Manual credit",
	"手动扣款":      "Manual debit",
	"红利":        "Bonus",
	"返水":        "Rebate",
	"转入":        "Transfer in",
	"转出":        "Transfer out",
	"游戏直接加款":    "Game direct credit",
	"游戏直接扣款":    "Game direct debit",
	"积分":        "Points",
	"未知":        "Unknown",
	"提现确认成功":    "Withdrawal confirmed successfully",
	"提现已拒绝":     "Withdrawal rejected",
}

// TranslateBalanceRemark 将余额日志备注中的中文翻译为英文（若已为英文则原样返回）
func TranslateBalanceRemark(remark string) string {
	if remark == "" {
		return remark
	}
	if !containsChinese(remark) {
		return remark
	}

	trimmed := strings.TrimSpace(remark)
	if v, ok := balanceRemarkTranslations[trimmed]; ok {
		return v
	}

	if strings.HasPrefix(trimmed, "（未确认）中心账户 -> ") {
		return "(Unconfirmed) Main account -> " + strings.TrimPrefix(trimmed, "（未确认）中心账户 -> ")
	}
	if strings.HasPrefix(trimmed, "中心账户 -> ") {
		return "Main account -> " + strings.TrimPrefix(trimmed, "中心账户 -> ")
	}
	if strings.HasPrefix(trimmed, "（未确认）") && strings.HasSuffix(trimmed, " -> 中心账户") {
		rest := strings.TrimPrefix(trimmed, "（未确认）")
		return "(Unconfirmed) " + strings.TrimSuffix(rest, " -> 中心账户") + " -> Main account"
	}
	if strings.HasSuffix(trimmed, " -> 中心账户") {
		return strings.TrimSuffix(trimmed, " -> 中心账户") + " -> Main account"
	}

	if strings.HasPrefix(trimmed, "大转盘中奖[") && strings.HasSuffix(trimmed, "]") {
		inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "大转盘中奖["), "]")
		return "Lucky spin winner [" + inner + "]"
	}
	if strings.HasPrefix(trimmed, "每日签到送彩金(") && strings.HasSuffix(trimmed, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "每日签到送彩金("), ")")
		return "Daily sign-in bonus (" + inner + ")"
	}
	if strings.HasPrefix(trimmed, "洗码返水(") && strings.HasSuffix(trimmed, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "洗码返水("), ")")
		return "Wash code rebate (" + strings.TrimSpace(inner) + ")"
	}
	if strings.HasPrefix(trimmed, "升级") && strings.HasSuffix(trimmed, "礼金") {
		inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "升级"), "礼金")
		inner = strings.TrimSpace(inner)
		if inner != "" {
			return "Upgrade " + inner + " bonus"
		}
	}

	if strings.HasPrefix(trimmed, "（已确认）") {
		rest := strings.TrimPrefix(trimmed, "（已确认）")
		translated := TranslateBalanceRemark(rest)
		if translated == rest {
			return "(Confirmed) " + rest
		}
		return "(Confirmed) " + translated
	}
	if strings.HasPrefix(trimmed, "（未确认）") {
		rest := strings.TrimPrefix(trimmed, "（未确认）")
		translated := TranslateBalanceRemark(rest)
		if translated == rest {
			return "(Unconfirmed) " + rest
		}
		return "(Unconfirmed) " + translated
	}

	// 兜底：去除中文，保证不返回任何中文
	cleaned := removeChinese(remark)
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "Remark unavailable"
	}
	return cleaned
}

func containsChinese(s string) bool {
	for _, r := range s {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}

func removeChinese(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '\u4e00' && r <= '\u9fff' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
