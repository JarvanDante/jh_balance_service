# 状态常量迁移指南

## 新的状态定义

```go
const (
    StatusPending    = 1 // 待处理
    StatusProcessing = 2 // 处理中
    StatusSuccess    = 3 // 成功
    StatusFail       = 4 // 失败
)
```

## 需要修改的文件

### ✅ 已完成

1. `internal/consts/backend/balance/balance_change.go` - 状态常量定义和映射

### 🔧 需要检查的文件

#### 1. internal/logic/backend/balance/balance.go

**第773行**: 硬编码的状态值

```go
// 旧代码
names := map[int]string{
    balanceConsts.StatusPending: "待审核",
    balanceConsts.StatusSuccess: "成功",
    balanceConsts.StatusFail:    "失败",
    4:                           "处理中",  // ❌ 硬编码
}

// 新代码
names := map[int]string{
    balanceConsts.StatusPending:    "待审核",
    balanceConsts.StatusProcessing: "处理中",  // ✅ 使用常量
    balanceConsts.StatusSuccess:    "成功",
    balanceConsts.StatusFail:       "失败",
}
```

#### 2. internal/dao/balance_log_payment.go

**第56行**: 待支付状态

```go
// 当前代码
Status: 0, // 待支付

// 建议：如果 0 表示待支付，需要定义一个新的常量
// 或者使用 StatusPending
Status: balanceConsts.StatusPending, // 待支付
```

#### 3. internal/logic/frontend/balance/balance.go

**第114行**: BalanceStatus

```go
// 当前代码
BalanceStatus: 1, // 1=正常

// 这个是用户余额状态，不是订单状态，可以保持不变
// 或者定义专门的用户状态常量
```

### 📝 注意事项

#### 1. 订单状态 vs 用户状态

- **订单状态**: 使用 `balanceConsts.StatusXXX`
- **用户余额状态**: `BalanceStatus: 1` (正常)
- **支付状态**: 可能需要单独定义 `StatusUnpaid = 0`

#### 2. 数据库中的现有数据

如果数据库中已有数据使用旧的状态值，需要：

1. 创建数据库迁移脚本
2. 更新所有表中的 status 字段

#### 3. API 兼容性

如果前端或其他服务依赖状态值，需要：

1. 检查 protobuf 定义
2. 更新 API 文档
3. 通知相关团队

## 建议的完整状态常量定义

```go
package balance

// 订单/交易状态
const (
    StatusUnpaid     = 0 // 未支付（仅用于充值订单）
    StatusPending    = 1 // 待处理
    StatusProcessing = 2 // 处理中
    StatusSuccess    = 3 // 成功
    StatusFail       = 4 // 失败
)

// 用户余额状态
const (
    BalanceStatusNormal  = 1 // 正常
    BalanceStatusFrozen  = 2 // 冻结
    BalanceStatusDisable = 3 // 禁用
)

// 状态名称映射
var StatusNames = map[int]string{
    StatusUnpaid:     "未支付",
    StatusPending:    "待处理",
    StatusProcessing: "处理中",
    StatusSuccess:    "成功",
    StatusFail:       "失败",
}

var BalanceStatusNames = map[int]string{
    BalanceStatusNormal:  "正常",
    BalanceStatusFrozen:  "冻结",
    BalanceStatusDisable: "禁用",
}
```

## 数据库迁移脚本示例

```sql
-- 如果需要更新数据库中的状态值
-- 注意：这只是示例，实际执行前请备份数据库

-- 更新 balance_log 表
UPDATE balance_log
SET status = CASE
    WHEN status = 1 THEN 1  -- 待处理保持不变
    WHEN status = 2 THEN 3  -- 成功改为3
    WHEN status = 3 THEN 4  -- 失败改为4
    ELSE status
END;

-- 更新 balance_log_withdraw 表
UPDATE balance_log_withdraw
SET status = CASE
    WHEN status = 1 THEN 1  -- 待处理保持不变
    WHEN status = 2 THEN 3  -- 成功改为3
    WHEN status = 3 THEN 4  -- 失败改为4
    ELSE status
END;

-- 更新 balance_log_payment 表
UPDATE balance_log_payment
SET status = CASE
    WHEN status = 0 THEN 0  -- 未支付保持不变
    WHEN status = 1 THEN 1  -- 待处理保持不变
    WHEN status = 2 THEN 3  -- 成功改为3
    WHEN status = 3 THEN 4  -- 失败改为4
    ELSE status
END;

-- 更新 balance_change 表
UPDATE balance_change
SET status = CASE
    WHEN status = 1 THEN 1  -- 待处理保持不变
    WHEN status = 2 THEN 3  -- 成功改为3
    WHEN status = 3 THEN 4  -- 失败改为4
    ELSE status
END;
```

## 检查清单

- [ ] 更新常量定义
- [ ] 更新状态名称映射
- [ ] 修改硬编码的状态值
- [ ] 检查所有使用状态的地方
- [ ] 更新 protobuf 定义（如果需要）
- [ ] 创建数据库迁移脚本
- [ ] 更新 API 文档
- [ ] 通知前端团队
- [ ] 编译测试
- [ ] 集成测试

---

**创建时间**: 2026-02-02
**状态**: 📝 待执行
