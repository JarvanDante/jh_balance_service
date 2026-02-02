# 提现功能重构总结

## 重构目标

将提现功能从面向过程的代码重构为分层架构，提高代码可维护性和可测试性。

## 架构变化

### 重构前

```
Logic 层 (balance.go)
  └── 所有逻辑都在一个方法里
      ├── 参数验证
      ├── 查询银行卡
      ├── 查询提现渠道
      ├── 检查待审核订单
      ├── 直接操作数据库（事务内）
      └── 返回结果
```

### 重构后

```
Logic 层 (balance.go)
  └── 薄薄的编排层
      ├── 获取用户信息
      ├── Nonce 验证
      └── 调用 Domain 层

Domain 层 (withdraw_domain.go)
  └── 核心业务逻辑
      ├── 业务规则验证
      ├── 调用 DAO 层
      └── 事务管理

DAO 层 (各个 dao 文件)
  └── 封装的数据库操作
      ├── GetDefaultBankCard()
      ├── GetByWithdrawId()
      ├── FreezeBalance()
      ├── CreateWithdrawOrder()
      └── ...
```

## 新增文件

### 1. Domain 层

- `internal/domain/withdraw/withdraw_domain.go`
  - `ApplyWithdraw()` - 申请提现核心逻辑
  - `validateWithdrawAmount()` - 验证提现金额
  - `validateWithdrawRange()` - 验证金额范围
  - `executeWithdrawTransaction()` - 执行提现事务
  - `checkBalance()` - 检查余额

## 修改文件

### 1. DAO 层封装

#### `internal/dao/balance_user.go`

```go
// 新增方法
- GetByUserIdWithLock()      // 查询并锁定用户余额
- FreezeBalance()             // 冻结余额
- UnfreezeBalance()           // 解冻余额
- DeductFrozen()              // 扣除冻结金额
- Transaction()               // 事务包装器
```

#### `internal/dao/user_bank.go`

```go
// 新增方法
- GetDefaultBankCard()        // 获取默认银行卡
```

#### `internal/dao/site_withdraw.go`

```go
// 新增方法
- GetByWithdrawId()           // 根据提现渠道ID查询
```

#### `internal/dao/balance_log_withdraw.go`

```go
// 新增方法
- CreateWithdrawOrder()       // 创建提现订单
- HasPendingOrder()           // 检查是否有待审核订单
```

#### `internal/dao/balance_log.go`

```go
// 新增方法
- CreateWithdrawLog()         // 创建提现日志
```

#### `internal/dao/balance_change.go`

```go
// 新增方法
- CreateWithdrawChange()      // 创建提现账变记录
```

### 2. Logic 层简化

#### `internal/logic/frontend/balance/balance.go`

- `Withdraw()` 方法从 200+ 行简化到 100 行
- 只保留：
  - 用户信息获取
  - Nonce 验证
  - 参数验证
  - 调用 Domain 层
  - 返回结果

## 重构收益

### ✅ 1. 代码清晰

- Logic 层只负责编排，不包含业务逻辑
- Domain 层集中管理业务规则
- DAO 层封装数据库操作

### ✅ 2. 易于维护

- 修改业务规则只需改 Domain 层
- 修改数据库操作只需改 DAO 层
- 不会影响其他层

### ✅ 3. 易于测试

- 可以单独测试 Domain 层的业务逻辑
- 可以 Mock DAO 层进行单元测试
- 不需要依赖数据库

### ✅ 4. 易于复用

- DAO 层的方法可以在其他地方复用
- Domain 层的业务逻辑可以被前台和后台共用
- 减少重复代码

### ✅ 5. 易于扩展

- 新增提现规则：只需在 Domain 层添加验证方法
- 新增数据库操作：只需在 DAO 层添加方法
- 不影响现有代码

## 使用示例

### Logic 层调用

```go
// 简洁的调用
withdrawDomain := &withdraw.WithdrawDomain{}
result, err := withdrawDomain.ApplyWithdraw(ctx, &withdraw.ApplyWithdrawRequest{
    SiteId:     siteId,
    UserId:     userId,
    Username:   username,
    Money:      req.Money,
    WithdrawId: req.WithdrawId,
    BankCardId: req.BankCardId,
})
```

### DAO 层调用

```go
// 封装好的方法，语义清晰
userBank, err := dao.UserBank.GetDefaultBankCard(ctx, siteId, userId)
siteWithdraw, err := dao.SiteWithdraw.GetByWithdrawId(ctx, siteId, withdrawId)
err := dao.BalanceUser.FreezeBalance(ctx, siteId, userId, amount)
```

## 后续优化建议

### 1. 状态机管理

可以参考之前讨论的状态机模式，将提现状态转换也封装到 Domain 层：

```go
// internal/domain/withdraw/withdraw_state.go
type WithdrawStateMachine struct{}
func (sm *WithdrawStateMachine) Approve(ctx, orderId, adminId, remark) error
func (sm *WithdrawStateMachine) Reject(ctx, orderId, adminId, reason) error
```

### 2. 其他功能重构

可以用同样的方式重构：

- 充值功能 → `internal/domain/deposit/`
- 转账功能 → `internal/domain/transfer/`
- 红利功能 → `internal/domain/bonus/`

### 3. 单元测试

为 Domain 层添加单元测试：

```go
// internal/domain/withdraw/withdraw_domain_test.go
func TestWithdrawDomain_ApplyWithdraw(t *testing.T) {
    // 测试业务逻辑
}
```

## 注意事项

1. **事务管理**：事务边界在 Domain 层的 `executeWithdrawTransaction` 方法中
2. **错误处理**：使用 `gerror.Wrap()` 包装错误，保留错误链
3. **日志记录**：关键步骤都有日志记录，方便排查问题
4. **浮点数精度**：使用四舍五入到2位小数避免精度问题

## 总结

这次重构将提现功能从面向过程的代码重构为清晰的分层架构：

- **Logic 层**：薄薄的编排层（100行）
- **Domain 层**：核心业务逻辑（200行）
- **DAO 层**：封装的数据库操作（每个方法 10-20行）

代码更清晰、更易维护、更易测试、更易扩展。这是一个可以长期稳定使用的架构。
