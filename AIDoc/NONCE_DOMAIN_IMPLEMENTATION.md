# Nonce 防重放逻辑封装完成

## 概述

已成功将 Nonce 防重放逻辑从 Logic 层封装到 Domain 层，实现了代码复用和职责分离。

## 实现文件

### 1. Domain 层 - Nonce 服务

**文件**: `internal/domain/common/nonce_domain.go`

**功能**:

- `ValidateAndConsume()`: 验证并消费 nonce（防重放）
- `Generate()`: 生成并存储 nonce

**特点**:

- 通用设计，支持多种业务类型（bizType: "deposit"/"withdraw"）
- 使用 Redis 存储，支持自动过期
- 完整的日志记录
- 防止并发重复提交

### 2. Logic 层 - 调用 Nonce 服务

**文件**: `internal/logic/frontend/balance/balance.go`

**已修改的方法**:

#### 2.1 GetDepositNonce() - 生成充值 nonce

```go
nonceService := &common.NonceService{}
err := nonceService.Generate(ctx, "deposit", int64(siteID), int64(userID), nonce, expiresIn)
```

#### 2.2 GetWithdrawNonce() - 生成提现 nonce

```go
nonceService := &common.NonceService{}
err := nonceService.Generate(ctx, "withdraw", int64(siteID), int64(userID), nonce, expiresIn)
```

#### 2.3 CreateDepositOrder() - 验证充值 nonce

```go
nonceService := &common.NonceService{}
err := nonceService.ValidateAndConsume(ctx, "deposit", int64(siteId), int64(userId), req.Nonce)
```

#### 2.4 Withdraw() - 验证提现 nonce

```go
nonceService := &common.NonceService{}
err := nonceService.ValidateAndConsume(ctx, "withdraw", int64(siteId), int64(userId), req.Nonce)
```

## 技术细节

### Redis Key 格式

```
{bizType}:nonce:{siteId}:{userId}
```

示例:

- 充值: `deposit:nonce:1:1001`
- 提现: `withdraw:nonce:1:1001`

### 防重放机制

1. **生成阶段**:
   - 生成 UUID 作为 nonce
   - 存储到 Redis，设置过期时间（60秒）
   - 返回给前端

2. **验证阶段**:
   - 从 Redis 获取存储的 nonce
   - 对比用户提交的 nonce
   - 验证成功后立即删除（单次使用）
   - 检查是否被并发删除

3. **安全保障**:
   - 时间窗口限制（60秒过期）
   - 单次使用（验证后删除）
   - 并发检测（检查删除计数）
   - 完整的日志记录

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

## 优势

### 1. 代码复用

- 充值和提现共用同一套 nonce 逻辑
- 通过 bizType 参数区分业务类型
- 减少重复代码

### 2. 职责分离

- Domain 层: 封装核心业务逻辑
- Logic 层: 编排业务流程
- 符合分层架构原则

### 3. 易于维护

- 集中管理 nonce 逻辑
- 修改只需在一处进行
- 便于添加新的业务类型

### 4. 易于测试

- Domain 层可独立测试
- 清晰的输入输出
- 便于 Mock

## 使用示例

### 生成 nonce

```go
nonceService := &common.NonceService{}
err := nonceService.Generate(
    ctx,
    "deposit",      // 业务类型
    int64(siteID),  // 站点ID
    int64(userID),  // 用户ID
    nonce,          // nonce值
    60,             // 过期时间（秒）
)
```

### 验证 nonce

```go
nonceService := &common.NonceService{}
err := nonceService.ValidateAndConsume(
    ctx,
    "deposit",      // 业务类型
    int64(siteID),  // 站点ID
    int64(userID),  // 用户ID
    req.Nonce,      // 用户提交的nonce
)
```

## 后续扩展

如需添加新的业务类型（如转账、红利等），只需：

1. 在生成 nonce 时传入新的 bizType（如 "transfer"）
2. 在验证 nonce 时使用相同的 bizType
3. 无需修改 NonceService 代码

## 相关文档

- [提现功能重构](./WITHDRAW_DOMAIN_IMPLEMENTATION.md)
- [gen_pb_clean.sh 脚本优化](../scripts/gen_pb_clean.sh)

---

**完成时间**: 2026-02-02
**状态**: ✅ 已完成并编译验证
