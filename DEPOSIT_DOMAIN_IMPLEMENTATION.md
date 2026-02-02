# 充值功能 Domain 层封装完成

## 概述

参考提现功能的封装方式，将充值逻辑从 Logic 层封装到 Domain 层，实现了代码复用和职责分离。

## 实现文件

### 1. Domain 层 - 充值服务

**文件**: `internal/domain/deposit/deposit_domain.go`

**核心方法**:

- `CreateDepositOrder()`: 创建充值订单（核心业务逻辑）
- `validateDepositRange()`: 验证充值金额范围
- `UpdateDepositOrderInfo()`: 更新充值订单信息

**业务流程**:

```
1. 设置默认订单类型
2. 查询支付渠道配置
3. 验证充值金额范围
4. 生成交易流水号
5. 创建充值订单记录
6. 创建支付日志记录
```

### 2. DAO 层 - 数据访问封装

#### 2.1 site_payment.go

**文件**: `internal/dao/site_payment.go`

**新增方法**:

```go
// GetByPaymentId 根据支付渠道ID查询支付渠道配置
func (d sitePaymentDao) GetByPaymentId(ctx context.Context, siteId, paymentId int64) (*entity.SitePayment, error)
```

**功能**: 查询支付渠道配置，验证渠道是否可用（status=1）

#### 2.2 balance_log_payment.go

**文件**: `internal/dao/balance_log_payment.go`

**新增方法**:

```go
// CreateDepositOrder 创建充值订单记录
func (d balanceLogPaymentDao) CreateDepositOrder(ctx context.Context, params *CreateDepositOrderParams) error

// UpdateDepositOrderInfo 更新充值订单信息
func (d balanceLogPaymentDao) UpdateDepositOrderInfo(ctx context.Context, tradeNo, sysTradeNo, channelCardNo, channelName, imageUrl string) error
```

**功能**:

- 创建充值订单记录（status=0 待支付）
- 更新订单的支付平台返回信息

#### 2.3 balance_log.go

**文件**: `internal/dao/balance_log.go`

**新增方法**:

```go
// CreateDepositLog 创建充值日志
func (d balanceLogDao) CreateDepositLog(ctx context.Context, params *CreateDepositLogParams) error
```

**功能**: 创建充值日志记录（status=1 待支付）

### 3. Logic 层 - 简化后的实现

**文件**: `internal/logic/frontend/balance/balance.go`

**重构后的 CreateDepositOrder()**:

```go
func (s *sBalance) CreateDepositOrder(ctx context.Context, req *v1.CreateDepositOrderReq) (*v1.CreateDepositOrderRes, error) {
    // 1. 获取用户信息
    userId, siteId, username := ...

    // 2. 验证 nonce（防重放）
    nonceService := &common.NonceService{}
    err := nonceService.ValidateAndConsume(ctx, "deposit", int64(siteId), int64(userId), req.Nonce)

    // 3. 调用 Domain 层创建充值订单
    depositDomain := &deposit.DepositDomain{}
    domainResult, err := depositDomain.CreateDepositOrder(ctx, &deposit.CreateDepositOrderRequest{
        SiteId:    int64(siteId),
        UserId:    int64(userId),
        Username:  username,
        PaymentId: int64(req.PaymentId),
        Money:     req.Money,
        // ... 其他参数
    })

    // 4. 调用支付平台
    payment := s.getPaymentPlatform(ctx, domainResult.SitePayment.Code)
    result, err := s.createPaymentOrder(ctx, payment, orderParams)

    // 5. 更新订单信息
    err = depositDomain.UpdateDepositOrderInfo(ctx, domainResult.TradeNo, result.SysTradeNo, ...)

    // 6. 返回结果
    return &v1.CreateDepositOrderRes{...}, nil
}
```

**代码行数对比**:

- **重构前**: ~180 行（包含所有业务逻辑和数据库操作）
- **重构后**: ~100 行（只负责编排和调用）
- **减少**: ~45% 代码量

## 架构对比

### 重构前（面向过程）

```
Logic 层
├─ 获取用户信息
├─ 验证 nonce
├─ 查询支付渠道（直接 SQL）
├─ 验证金额范围
├─ 生成流水号
├─ 创建订单记录（直接 SQL）
├─ 创建日志记录（直接 SQL）
├─ 调用支付平台
└─ 更新订单信息（直接 SQL）
```

### 重构后（分层架构）

```
Controller 层
├─ 参数格式验证
└─ 调用 Logic 层

Logic 层
├─ 获取用户信息
├─ 验证 nonce（防重放）
├─ 调用 Domain 层
├─ 调用支付平台
└─ 更新订单信息

Domain 层
├─ 查询支付渠道（调用 DAO）
├─ 验证金额范围（业务规则）
├─ 生成流水号
├─ 创建订单记录（调用 DAO）
└─ 创建日志记录（调用 DAO）

DAO 层
├─ GetByPaymentId()
├─ CreateDepositOrder()
├─ CreateDepositLog()
└─ UpdateDepositOrderInfo()
```

## 封装的 DAO 方法

### 1. SitePayment DAO

| 方法           | 功能             | 参数              |
| -------------- | ---------------- | ----------------- |
| GetByPaymentId | 查询支付渠道配置 | siteId, paymentId |

### 2. BalanceLogPayment DAO

| 方法                   | 功能         | 参数                     |
| ---------------------- | ------------ | ------------------------ |
| CreateDepositOrder     | 创建充值订单 | CreateDepositOrderParams |
| UpdateDepositOrderInfo | 更新订单信息 | tradeNo, sysTradeNo, ... |

### 3. BalanceLog DAO

| 方法             | 功能         | 参数                   |
| ---------------- | ------------ | ---------------------- |
| CreateDepositLog | 创建充值日志 | CreateDepositLogParams |

## 业务流程图

```
前端请求
    ↓
Controller 层
    ├─ 验证 PaymentId > 0
    ├─ 验证 Money >= 0.01
    └─ 验证 Nonce 不为空
    ↓
Logic 层
    ├─ 获取用户信息（Context）
    ├─ 验证 Nonce（Redis）
    └─ 调用 Domain 层
    ↓
Domain 层
    ├─ 设置默认订单类型
    ├─ 查询支付渠道配置（DAO）
    ├─ 验证金额范围
    ├─ 生成交易流水号
    ├─ 创建充值订单（DAO）
    └─ 创建支付日志（DAO）
    ↓
Logic 层
    ├─ 调用支付平台
    └─ 更新订单信息（Domain）
    ↓
返回结果
```

## 优势

### 1. 代码复用

- 充值订单创建逻辑可在多处复用
- DAO 方法可被其他服务调用
- 减少重复代码

### 2. 职责清晰

- **Controller**: 参数验证
- **Logic**: 业务编排、调用支付平台
- **Domain**: 核心业务逻辑
- **DAO**: 数据访问

### 3. 易于维护

- 修改充值逻辑只需修改 Domain 层
- 数据库操作集中在 DAO 层
- 便于单元测试

### 4. 易于扩展

- 新增支付渠道只需扩展支付平台
- 新增充值类型只需扩展 Domain 层
- 不影响现有代码

## 与提现功能对比

| 特性        | 提现               | 充值               |
| ----------- | ------------------ | ------------------ |
| Domain 文件 | withdraw_domain.go | deposit_domain.go  |
| 核心方法    | ApplyWithdraw      | CreateDepositOrder |
| DAO 封装    | 6个方法            | 3个方法            |
| 事务处理    | 是（冻结余额）     | 否（待支付）       |
| 外部调用    | 无                 | 支付平台           |

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

## 测试建议

### 1. 单元测试

```go
// 测试 Domain 层
func TestDepositDomain_CreateDepositOrder(t *testing.T) {
    // Mock DAO 层
    // 测试业务逻辑
}

// 测试 DAO 层
func TestSitePaymentDao_GetByPaymentId(t *testing.T) {
    // 测试数据库查询
}
```

### 2. 集成测试

- 测试完整的充值流程
- 测试支付平台调用
- 测试订单状态更新

### 3. 性能测试

- 测试并发充值
- 测试数据库连接池
- 测试 Redis 性能

## 后续优化建议

### 1. 支付平台抽象

```go
// 定义支付平台接口
type PaymentPlatform interface {
    CreateOrder(ctx context.Context, params *PaymentOrderParams) (*PaymentOrderResult, error)
}

// 实现工厂模式
func GetPaymentPlatform(code string) PaymentPlatform {
    switch code {
    case "MG":
        return &MGPayment{}
    case "Alipay":
        return &AlipayPayment{}
    default:
        return nil
    }
}
```

### 2. 事件驱动

```go
// 充值成功后发送事件
type DepositSuccessEvent struct {
    TradeNo string
    UserId  int64
    Money   float64
}

// 监听事件，更新用户余额
func OnDepositSuccess(event *DepositSuccessEvent) {
    // 更新余额
    // 发送通知
}
```

### 3. 缓存优化

```go
// 缓存支付渠道配置
func (d *DepositDomain) GetSitePaymentWithCache(ctx context.Context, siteId, paymentId int64) (*entity.SitePayment, error) {
    // 先从 Redis 获取
    // 如果不存在，从数据库查询并缓存
}
```

## 相关文档

- [提现功能重构](./WITHDRAW_DOMAIN_IMPLEMENTATION.md)
- [Nonce 防重放实现](./NONCE_DOMAIN_IMPLEMENTATION.md)
- [Controller 层参数验证](./CONTROLLER_VALIDATION_REFACTOR.md)

---

**完成时间**: 2026-02-02
**状态**: ✅ 已完成并编译验证
