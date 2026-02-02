# 状态常量重构完成

## 概述

已成功将所有硬编码的状态值替换为常量，提高代码可维护性和可读性。

## 新增的状态常量

### 1. 订单/交易状态

```go
const (
    StatusPending    = 1 // 待处理
    StatusProcessing = 2 // 处理中
    StatusSuccess    = 3 // 成功
    StatusFail       = 4 // 失败
)
```

### 2. 充值订单状态

```go
const (
    PaymentStatusUnpaid = 0 // 未支付
)
```

### 3. 渠道状态（支付渠道、提现渠道）

```go
const (
    ChannelStatusDisabled = 0 // 禁用
    ChannelStatusEnabled  = 1 // 启用
)
```

### 4. 用户余额状态

```go
const (
    BalanceStatusNormal  = 1 // 正常
    BalanceStatusFrozen  = 2 // 冻结
    BalanceStatusDisable = 3 // 禁用
)
```

## 修改的文件清单

### ✅ 1. 常量定义文件

**文件**: `internal/consts/backend/balance/balance_change.go`

- 新增 `StatusProcessing = 2` (处理中)
- 新增 `PaymentStatusUnpaid = 0` (未支付)
- 新增 `ChannelStatusDisabled/Enabled` (渠道状态)
- 新增 `BalanceStatusNormal/Frozen/Disable` (用户余额状态)
- 更新 `StatusNames` 映射

### ✅ 2. 前台逻辑文件

**文件**: `internal/logic/frontend/balance/balance.go`

**修改内容**:

```go
// 修改前
if paymentLog.Status == 2 { // STATUS_SUCCESS
if paymentLog.Status == 3 { // STATUS_FAIL
Where("status", 1)
BalanceStatus: 1, // 1=正常

// 修改后
if paymentLog.Status == balanceConsts.StatusSuccess {
if paymentLog.Status == balanceConsts.StatusFail {
Where("status", balanceConsts.ChannelStatusEnabled)
BalanceStatus: balanceConsts.BalanceStatusNormal,
```

**修改位置**:

- 第114行: `BalanceStatus` 初始化用户余额
- 第977行: 查询支付渠道列表
- 第1035行: 查询提现渠道列表
- 第1991行: 充值回调检查订单状态（成功）
- 第1999行: 充值回调检查订单状态（失败）

### ✅ 3. 后台逻辑文件

**文件**: `internal/logic/backend/balance/balance.go`

**修改内容**:

```go
// 修改前
names := map[int]string{
    balanceConsts.StatusPending: "待审核",
    balanceConsts.StatusSuccess: "成功",
    balanceConsts.StatusFail:    "失败",
    4:                           "处理中",  // 硬编码
}

// 修改后
names := map[int]string{
    balanceConsts.StatusPending:    "待审核",
    balanceConsts.StatusProcessing: "处理中",  // 使用常量
    balanceConsts.StatusSuccess:    "成功",
    balanceConsts.StatusFail:       "失败",
}
```

**修改位置**:

- 第768-773行: `getWithdrawStatusName()` 方法

### ✅ 4. DAO 层文件

#### 4.1 site_payment.go

**修改内容**:

```go
// 修改前
Where("status", 1)

// 修改后
Where("status", balanceConsts.ChannelStatusEnabled)
```

- 添加 `balanceConsts` import
- 第40行: `GetByPaymentId()` 方法

#### 4.2 site_withdraw.go

**修改内容**:

```go
// 修改前
Where("status", 1)

// 修改后
Where("status", balanceConsts.ChannelStatusEnabled)
```

- 添加 `balanceConsts` import
- 第41行: `GetByWithdrawId()` 方法

#### 4.3 balance_log_payment.go

**修改内容**:

```go
// 修改前
Status: 0, // 待支付

// 修改后
Status: balanceConsts.PaymentStatusUnpaid, // 未支付
```

- 添加 `balanceConsts` import
- 第56行: `CreateDepositOrder()` 方法

## 状态使用场景

### 订单状态流转

#### 充值订单

```
PaymentStatusUnpaid (0) → StatusPending (1) → StatusSuccess (3) / StatusFail (4)
     未支付                    待处理              成功           失败
```

#### 提现订单

```
StatusPending (1) → StatusProcessing (2) → StatusSuccess (3) / StatusFail (4)
    待处理              处理中                  成功           失败
```

#### 游戏转账

```
StatusPending (1) → StatusSuccess (3) / StatusFail (4)
    待确认              成功           失败
```

### 渠道状态

```
ChannelStatusDisabled (0) - 禁用
ChannelStatusEnabled (1)  - 启用
```

### 用户余额状态

```
BalanceStatusNormal (1)  - 正常
BalanceStatusFrozen (2)  - 冻结
BalanceStatusDisable (3) - 禁用
```

## 修改统计

| 类型     | 修改数量     | 文件数      |
| -------- | ------------ | ----------- |
| 常量定义 | 13个常量     | 1           |
| 前台逻辑 | 5处          | 1           |
| 后台逻辑 | 1处          | 1           |
| DAO 层   | 3处          | 3           |
| **总计** | **22处修改** | **6个文件** |

## 优势

### 1. 代码可读性

```go
// 修改前 - 不清楚 2 代表什么
if status == 2 {
    // ...
}

// 修改后 - 一目了然
if status == balanceConsts.StatusSuccess {
    // ...
}
```

### 2. 易于维护

- 修改状态值只需修改常量定义
- 不需要在多个文件中查找和替换
- 减少出错的可能性

### 3. 类型安全

- 使用常量避免魔法数字
- IDE 可以提供自动补全
- 编译时检查，避免运行时错误

### 4. 文档化

- 常量名称即文档
- 注释说明状态含义
- 便于新人理解代码

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
- 无警告信息

## 测试建议

### 1. 单元测试

```go
func TestStatusConstants(t *testing.T) {
    assert.Equal(t, 1, balanceConsts.StatusPending)
    assert.Equal(t, 2, balanceConsts.StatusProcessing)
    assert.Equal(t, 3, balanceConsts.StatusSuccess)
    assert.Equal(t, 4, balanceConsts.StatusFail)
}
```

### 2. 集成测试

- 测试充值订单状态流转
- 测试提现订单状态流转
- 测试渠道启用/禁用
- 测试用户余额状态

### 3. 回归测试

- 测试充值功能
- 测试提现功能
- 测试游戏转账
- 测试后台管理

## 注意事项

### 1. 数据库数据

由于项目未上线，数据可控，不需要数据库迁移。
如果后续需要迁移，参考 `STATUS_MIGRATION_GUIDE.md`

### 2. API 兼容性

前端和其他服务需要知道新的状态值：

- StatusPending = 1
- StatusProcessing = 2
- StatusSuccess = 3
- StatusFail = 4

### 3. 文档更新

需要更新以下文档：

- API 文档
- 数据库设计文档
- 前端对接文档

## 相关文档

- [状态常量更新](./STATUS_CONSTANTS_UPDATE.md)
- [状态迁移指南](./STATUS_MIGRATION_GUIDE.md)
- [充值功能重构](../DEPOSIT_DOMAIN_IMPLEMENTATION.md)
- [提现功能重构](./WITHDRAW_DOMAIN_IMPLEMENTATION.md)

---

**完成时间**: 2026-02-02
**状态**: ✅ 已完成并编译验证
**修改文件**: 6个
**修改行数**: 22处
