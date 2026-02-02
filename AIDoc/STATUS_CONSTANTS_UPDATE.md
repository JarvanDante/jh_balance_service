# 状态常量更新完成

## 概述

已成功更新状态常量定义，添加了 `StatusProcessing = 2` (处理中) 状态。

## 更新内容

### 1. 状态常量定义

**文件**: `internal/consts/backend/balance/balance_change.go`

**新的状态定义**:

```go
const (
    StatusPending    = 1 // 待处理
    StatusProcessing = 2 // 处理中  ← 新增
    StatusSuccess    = 3 // 成功
    StatusFail       = 4 // 失败
)
```

**状态名称映射**:

```go
var StatusNames = map[int]string{
    StatusPending:    "待处理",
    StatusProcessing: "处理中",  ← 新增
    StatusSuccess:    "成功",
    StatusFail:       "失败",
}
```

### 2. 修复硬编码状态值

**文件**: `internal/logic/backend/balance/balance.go`

**修改前**:

```go
func (s *sBalance) getWithdrawStatusName(status int) string {
    names := map[int]string{
        balanceConsts.StatusPending: "待审核",
        balanceConsts.StatusSuccess: "成功",
        balanceConsts.StatusFail:    "失败",
        4:                           "处理中",  // ❌ 硬编码
    }
    // ...
}
```

**修改后**:

```go
func (s *sBalance) getWithdrawStatusName(status int) string {
    names := map[int]string{
        balanceConsts.StatusPending:    "待审核",
        balanceConsts.StatusProcessing: "处理中",  // ✅ 使用常量
        balanceConsts.StatusSuccess:    "成功",
        balanceConsts.StatusFail:       "失败",
    }
    // ...
}
```

## 状态使用说明

### 订单/交易状态流转

#### 充值订单

```
0 (未支付) → 1 (待处理) → 3 (成功) / 4 (失败)
```

#### 提现订单

```
1 (待处理) → 2 (处理中) → 3 (成功) / 4 (失败)
                      ↘ 4 (失败)
```

#### 游戏转账

```
1 (待确认) → 3 (成功) / 4 (失败)
```

### 状态常量使用位置

#### 1. 待处理 (StatusPending = 1)

- 提现申请创建时
- 游戏转账创建时
- 充值订单创建时（部分场景）

**使用文件**:

- `internal/dao/balance_log_withdraw.go` - 创建提现订单
- `internal/dao/balance_log.go` - 创建提现日志、充值日志
- `internal/dao/balance_change.go` - 创建账变记录
- `internal/logic/frontend/balance/balance.go` - 游戏转账

#### 2. 处理中 (StatusProcessing = 2)

- 提现订单审核通过，等待支付平台处理
- 后台管理员标记为处理中

**使用文件**:

- `internal/logic/backend/balance/balance.go` - 提现审核

#### 3. 成功 (StatusSuccess = 3)

- 充值成功
- 提现成功
- 游戏转账成功
- 手动加款/扣款成功

**使用文件**:

- `internal/logic/frontend/balance/balance.go` - 转账成功、充值回调
- `internal/logic/backend/balance/balance.go` - 提现审核通过、手动操作

#### 4. 失败 (StatusFail = 4)

- 充值失败
- 提现拒绝
- 游戏转账失败

**使用文件**:

- `internal/logic/frontend/balance/balance.go` - 转账失败
- `internal/logic/backend/balance/balance.go` - 提现拒绝

## 代码中的状态使用统计

### 使用 balanceConsts.StatusPending 的地方

- ✅ `internal/dao/balance_log_withdraw.go` (2处)
- ✅ `internal/dao/balance_log.go` (2处)
- ✅ `internal/dao/balance_change.go` (1处)
- ✅ `internal/logic/frontend/balance/balance.go` (8处)
- ✅ `internal/logic/backend/balance/balance.go` (2处)

### 使用 balanceConsts.StatusProcessing 的地方

- ✅ `internal/consts/backend/balance/balance_change.go` (常量定义)
- ✅ `internal/logic/backend/balance/balance.go` (状态名称映射)

### 使用 balanceConsts.StatusSuccess 的地方

- ✅ `internal/logic/frontend/balance/balance.go` (10处)
- ✅ `internal/logic/backend/balance/balance.go` (8处)

### 使用 balanceConsts.StatusFail 的地方

- ✅ `internal/logic/frontend/balance/balance.go` (6处)
- ✅ `internal/logic/backend/balance/balance.go` (3处)

## 特殊状态值

### Status = 0 (未支付)

**位置**: `internal/dao/balance_log_payment.go`

```go
Status: 0, // 待支付
```

**说明**:

- 仅用于充值订单的初始状态
- 表示订单已创建，但用户还未支付
- 不是通用状态，不需要定义常量

### BalanceStatus = 1 (正常)

**位置**: `internal/logic/frontend/balance/balance.go`

```go
BalanceStatus: 1, // 1=正常
```

**说明**:

- 这是用户余额状态，不是订单状态
- 与订单状态是不同的概念
- 建议单独定义用户余额状态常量

## 编译验证

### 编译命令

```bash
cd jh_balance_service
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jh_balance_service .
```

### 编译结果

✅ 编译成功

- 二进制文件: `jh_balance_service`
- 文件大小: 42MB
- 无编译错误

## 后续建议

### 1. 定义用户余额状态常量

```go
// 用户余额状态
const (
    BalanceStatusNormal  = 1 // 正常
    BalanceStatusFrozen  = 2 // 冻结
    BalanceStatusDisable = 3 // 禁用
)
```

### 2. 定义充值订单状态常量

```go
// 充值订单状态
const (
    PaymentStatusUnpaid  = 0 // 未支付
    PaymentStatusPending = 1 // 待处理
    PaymentStatusSuccess = 3 // 成功
    PaymentStatusFail    = 4 // 失败
)
```

### 3. 数据库迁移

如果数据库中已有数据使用旧的状态值，需要创建迁移脚本。
详见: `STATUS_MIGRATION_GUIDE.md`

### 4. API 文档更新

更新 API 文档中的状态说明：

- protobuf 定义
- Swagger 文档
- 前端对接文档

### 5. 单元测试

添加状态转换的单元测试：

```go
func TestStatusTransition(t *testing.T) {
    // 测试状态转换逻辑
}
```

## 相关文档

- [状态迁移指南](STATUS_MIGRATION_GUIDE.md)
- [充值功能重构](../DEPOSIT_DOMAIN_IMPLEMENTATION.md)
- [提现功能重构](./WITHDRAW_DOMAIN_IMPLEMENTATION.md)

---

**完成时间**: 2026-02-02
**状态**: ✅ 已完成并编译验证
