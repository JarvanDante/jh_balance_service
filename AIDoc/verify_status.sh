#!/bin/bash

# 状态常量验证脚本
# 用于快速验证状态常量修改是否正确

echo "=========================================="
echo "状态常量验证脚本"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 计数器
PASS=0
FAIL=0

# 验证函数
check_pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((PASS++))
}

check_fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    ((FAIL++))
}

check_warn() {
    echo -e "${YELLOW}⚠️  WARN${NC}: $1"
}

echo "1. 检查常量定义..."
echo "-------------------------------------------"

# 检查 StatusPending
if grep -q "StatusPending.*=.*1.*待处理" internal/consts/backend/balance/balance_change.go; then
    check_pass "StatusPending = 1 (待处理)"
else
    check_fail "StatusPending 定义不正确"
fi

# 检查 StatusProcessing
if grep -q "StatusProcessing.*=.*2.*处理中" internal/consts/backend/balance/balance_change.go; then
    check_pass "StatusProcessing = 2 (处理中)"
else
    check_fail "StatusProcessing 定义不正确"
fi

# 检查 StatusSuccess
if grep -q "StatusSuccess.*=.*3.*成功" internal/consts/backend/balance/balance_change.go; then
    check_pass "StatusSuccess = 3 (成功)"
else
    check_fail "StatusSuccess 定义不正确"
fi

# 检查 StatusFail
if grep -q "StatusFail.*=.*4.*失败" internal/consts/backend/balance/balance_change.go; then
    check_pass "StatusFail = 4 (失败)"
else
    check_fail "StatusFail 定义不正确"
fi

echo ""
echo "2. 检查硬编码状态值..."
echo "-------------------------------------------"

# 检查是否还有硬编码的 status == 2
if grep -r "Status\s*==\s*2\s*//" internal/logic/ 2>/dev/null | grep -v "StatusProcessing"; then
    check_fail "发现硬编码 status == 2"
else
    check_pass "没有硬编码 status == 2"
fi

# 检查是否还有硬编码的 status == 3
if grep -r "Status\s*==\s*3\s*//" internal/logic/ 2>/dev/null | grep -v "StatusSuccess"; then
    check_fail "发现硬编码 status == 3"
else
    check_pass "没有硬编码 status == 3"
fi

# 检查是否还有硬编码的 Where("status", 1)
HARD_CODED=$(grep -r 'Where("status", 1)' internal/logic/ 2>/dev/null | wc -l)
if [ "$HARD_CODED" -gt 0 ]; then
    check_warn "发现 $HARD_CODED 处 Where(\"status\", 1) (可能是渠道状态)"
else
    check_pass "Logic 层没有硬编码 Where(\"status\", 1)"
fi

echo ""
echo "3. 检查常量使用..."
echo "-------------------------------------------"

# 检查 StatusSuccess 使用
SUCCESS_COUNT=$(grep -r "StatusSuccess" internal/ 2>/dev/null | wc -l)
if [ "$SUCCESS_COUNT" -gt 0 ]; then
    check_pass "StatusSuccess 被使用 $SUCCESS_COUNT 次"
else
    check_warn "StatusSuccess 没有被使用"
fi

# 检查 StatusFail 使用
FAIL_COUNT=$(grep -r "StatusFail" internal/ 2>/dev/null | wc -l)
if [ "$FAIL_COUNT" -gt 0 ]; then
    check_pass "StatusFail 被使用 $FAIL_COUNT 次"
else
    check_warn "StatusFail 没有被使用"
fi

# 检查 StatusProcessing 使用
PROCESSING_COUNT=$(grep -r "StatusProcessing" internal/ 2>/dev/null | wc -l)
if [ "$PROCESSING_COUNT" -gt 0 ]; then
    check_pass "StatusProcessing 被使用 $PROCESSING_COUNT 次"
else
    check_warn "StatusProcessing 没有被使用"
fi

echo ""
echo "4. 检查 DAO 层..."
echo "-------------------------------------------"

# 检查 site_payment.go
if grep -q "ChannelStatusEnabled" internal/dao/site_payment.go; then
    check_pass "site_payment.go 使用 ChannelStatusEnabled"
else
    check_fail "site_payment.go 没有使用 ChannelStatusEnabled"
fi

# 检查 site_withdraw.go
if grep -q "ChannelStatusEnabled" internal/dao/site_withdraw.go; then
    check_pass "site_withdraw.go 使用 ChannelStatusEnabled"
else
    check_fail "site_withdraw.go 没有使用 ChannelStatusEnabled"
fi

# 检查 balance_log_payment.go
if grep -q "PaymentStatusUnpaid" internal/dao/balance_log_payment.go; then
    check_pass "balance_log_payment.go 使用 PaymentStatusUnpaid"
else
    check_fail "balance_log_payment.go 没有使用 PaymentStatusUnpaid"
fi

echo ""
echo "5. 编译验证..."
echo "-------------------------------------------"

# 尝试编译
if GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jh_balance_service_test . 2>/dev/null; then
    check_pass "编译成功"
    rm -f jh_balance_service_test
else
    check_fail "编译失败"
fi

echo ""
echo "=========================================="
echo "验证结果汇总"
echo "=========================================="
echo -e "${GREEN}通过: $PASS${NC}"
echo -e "${RED}失败: $FAIL${NC}"
echo ""

if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}🎉 所有验证通过！可以放心使用。${NC}"
    exit 0
else
    echo -e "${RED}⚠️  发现 $FAIL 个问题，请检查！${NC}"
    exit 1
fi
