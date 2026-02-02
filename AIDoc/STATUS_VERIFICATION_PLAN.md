# 状态常量修改验证方案

## 验证目标

确保所有状态值修改后，业务逻辑正常运行，不会出现状态判断错误。

## 一、快速验证清单

### ✅ 1. 编译验证（已完成）

```bash
cd jh_balance_service
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jh_balance_service .
```

**结果**: ✅ 编译成功，无错误

### ✅ 2. 常量值验证

#### 检查常量定义

**文件**: `internal/consts/backend/balance/balance_change.go`

```go
// 订单状态
StatusPending    = 1  ✅ 待处理
StatusProcessing = 2  ✅ 处理中（新增）
StatusSuccess    = 3  ✅ 成功
StatusFail       = 4  ✅ 失败

// 充值订单
PaymentStatusUnpaid = 0  ✅ 未支付

// 渠道状态
ChannelStatusDisabled = 0  ✅ 禁用
ChannelStatusEnabled  = 1  ✅ 启用

// 用户余额状态
BalanceStatusNormal  = 1  ✅ 正常
BalanceStatusFrozen  = 2  ✅ 冻结
BalanceStatusDisable = 3  ✅ 禁用
```

### 🔍 3. 代码审查验证

#### 3.1 充值回调状态检查

**文件**: `internal/logic/frontend/balance/balance.go` (第1991-2006行)

```go
// ✅ 检查成功状态
if paymentLog.Status == balanceConsts.StatusSuccess {  // 3
    return "订单已完成"
}

// ✅ 检查失败状态
if paymentLog.Status == balanceConsts.StatusFail {  // 4
    return "订单已失败"
}
```

**验证**:

- ✅ 使用 StatusSuccess (3) 判断成功
- ✅ 使用 StatusFail (4) 判断失败

#### 3.2 提现状态名称映射

**文件**: `internal/logic/backend/balance/balance.go` (第768-773行)

```go
names := map[int]string{
    balanceConsts.StatusPending:    "待审核",  // 1
    balanceConsts.StatusProcessing: "处理中",  // 2
    balanceConsts.StatusSuccess:    "成功",    // 3
    balanceConsts.StatusFail:       "失败",    // 4
}
```

**验证**:

- ✅ 所有状态都使用常量
- ✅ 没有硬编码数字

#### 3.3 渠道状态查询

**文件**: `internal/dao/site_payment.go` (第40行)

```go
Where("status", balanceConsts.ChannelStatusEnabled)  // 1
```

**验证**:

- ✅ 查询启用的支付渠道

**文件**: `internal/dao/site_withdraw.go` (第41行)

```go
Where("status", balanceConsts.ChannelStatusEnabled)  // 1
```

**验证**:

- ✅ 查询启用的提现渠道

## 二、功能测试验证

### 测试环境准备

```bash
# 1. 启动服务
cd jh_balance_service
./jh_balance_service

# 2. 检查服务是否正常启动
# 查看日志，确认没有错误
```

### 测试用例

#### 测试 1: 充值功能

```bash
# 测试场景：用户发起充值
# 预期结果：
# - 创建订单时 status = 0 (PaymentStatusUnpaid)
# - 支付成功后 status = 3 (StatusSuccess)
# - 支付失败后 status = 4 (StatusFail)

# 验证点：
1. 创建充值订单
   - 检查 balance_log_payment 表
   - status 应该是 0 (未支付)

2. 充值回调（成功）
   - 检查 balance_log_payment 表
   - status 应该是 3 (成功)
   - 用户余额应该增加

3. 充值回调（失败）
   - 检查 balance_log_payment 表
   - status 应该是 4 (失败)
   - 用户余额不变
```

#### 测试 2: 提现功能

```bash
# 测试场景：用户申请提现
# 预期结果：
# - 创建订单时 status = 1 (StatusPending)
# - 审核通过后 status = 2 (StatusProcessing)
# - 提现成功后 status = 3 (StatusSuccess)
# - 提现拒绝后 status = 4 (StatusFail)

# 验证点：
1. 申请提现
   - 检查 balance_log_withdraw 表
   - status 应该是 1 (待处理)
   - 余额应该被冻结

2. 后台审核通过
   - status 应该是 2 (处理中)

3. 提现成功
   - status 应该是 3 (成功)
   - 冻结金额减少

4. 提现拒绝
   - status 应该是 4 (失败)
   - 冻结金额返回到可用余额
```

#### 测试 3: 渠道管理

```bash
# 测试场景：查询支付/提现渠道
# 预期结果：
# - 只返回 status = 1 (ChannelStatusEnabled) 的渠道

# 验证点：
1. 查询支付渠道列表
   - 只显示启用的渠道
   - 禁用的渠道不显示

2. 查询提现渠道列表
   - 只显示启用的渠道
   - 禁用的渠道不显示
```

## 三、数据库验证

### SQL 验证脚本

#### 1. 检查充值订单状态

```sql
-- 查看充值订单的状态分布
SELECT
    status,
    CASE
        WHEN status = 0 THEN '未支付'
        WHEN status = 1 THEN '待处理'
        WHEN status = 3 THEN '成功'
        WHEN status = 4 THEN '失败'
        ELSE '未知'
    END as status_name,
    COUNT(*) as count
FROM balance_log_payment
GROUP BY status
ORDER BY status;

-- 预期结果：
-- status | status_name | count
-- 0      | 未支付      | X
-- 1      | 待处理      | X
-- 3      | 成功        | X
-- 4      | 失败        | X
```

#### 2. 检查提现订单状态

```sql
-- 查看提现订单的状态分布
SELECT
    status,
    CASE
        WHEN status = 1 THEN '待处理'
        WHEN status = 2 THEN '处理中'
        WHEN status = 3 THEN '成功'
        WHEN status = 4 THEN '失败'
        ELSE '未知'
    END as status_name,
    COUNT(*) as count
FROM balance_log_withdraw
GROUP BY status
ORDER BY status;

-- 预期结果：
-- status | status_name | count
-- 1      | 待处理      | X
-- 2      | 处理中      | X
-- 3      | 成功        | X
-- 4      | 失败        | X
```

#### 3. 检查渠道状态

```sql
-- 查看支付渠道状态
SELECT
    id,
    name,
    status,
    CASE
        WHEN status = 0 THEN '禁用'
        WHEN status = 1 THEN '启用'
        ELSE '未知'
    END as status_name
FROM site_payment
ORDER BY status DESC, sort DESC;

-- 查看提现渠道状态
SELECT
    id,
    name,
    status,
    CASE
        WHEN status = 0 THEN '禁用'
        WHEN status = 1 THEN '启用'
        ELSE '未知'
    END as status_name
FROM site_withdraw
ORDER BY status DESC, sort DESC;
```

## 四、日志验证

### 关键日志检查点

#### 1. 充值回调日志

```bash
# 查找充值回调日志
grep "充值订单已完成" logs/*.log
grep "充值订单已失败" logs/*.log

# 预期：
# - 成功的订单应该显示 "充值订单已完成"
# - 失败的订单应该显示 "充值订单已失败"
```

#### 2. 提现审核日志

```bash
# 查找提现审核日志
grep "提现确认成功" logs/*.log
grep "提现已拒绝" logs/*.log

# 预期：
# - 审核通过应该显示 "提现确认成功"
# - 审核拒绝应该显示 "提现已拒绝"
```

## 五、回滚方案

### 如果发现问题，快速回滚

#### Git 回滚

```bash
# 查看最近的提交
git log --oneline -5

# 回滚到修改前的版本
git reset --hard <commit_hash>

# 重新编译
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o jh_balance_service .
```

#### 手动回滚（修改常量值）

如果只是状态值有问题，可以快速修改常量：

**文件**: `internal/consts/backend/balance/balance_change.go`

```go
// 如果需要回滚到旧的状态值
const (
    StatusPending    = 1 // 待处理
    StatusSuccess    = 2 // 成功（旧值）
    StatusFail       = 3 // 失败（旧值）
)
```

## 六、验证检查表

### 代码层面

- [x] 编译成功，无错误
- [x] 所有硬编码都替换为常量
- [x] 常量值定义正确
- [x] import 语句正确

### 功能层面

- [ ] 充值功能正常
  - [ ] 创建订单 (status=0)
  - [ ] 支付成功 (status=3)
  - [ ] 支付失败 (status=4)
- [ ] 提现功能正常
  - [ ] 申请提现 (status=1)
  - [ ] 审核通过 (status=2)
  - [ ] 提现成功 (status=3)
  - [ ] 提现拒绝 (status=4)
- [ ] 渠道查询正常
  - [ ] 只显示启用的渠道 (status=1)

### 数据层面

- [ ] 数据库状态值正确
- [ ] 状态流转正常
- [ ] 日志记录正确

## 七、建议的验证顺序

### 第一步：本地验证（5分钟）

1. ✅ 编译验证（已完成）
2. ✅ 代码审查（已完成）
3. 查看常量定义是否正确

### 第二步：开发环境验证（30分钟）

1. 启动服务
2. 测试充值功能
3. 测试提现功能
4. 测试渠道查询
5. 检查日志

### 第三步：数据库验证（10分钟）

1. 执行 SQL 验证脚本
2. 检查状态分布
3. 确认数据一致性

### 第四步：集成测试（1小时）

1. 完整的充值流程
2. 完整的提现流程
3. 后台管理功能
4. 异常场景测试

## 八、常见问题排查

### 问题 1: 充值回调失败

**症状**: 充值成功但余额没增加

**排查**:

```sql
-- 检查订单状态
SELECT * FROM balance_log_payment WHERE trade_no = 'xxx';

-- 检查状态是否正确
-- status 应该是 3 (成功)
```

**解决**: 检查回调逻辑中的状态判断

### 问题 2: 提现审核失败

**症状**: 审核通过但状态没变

**排查**:

```sql
-- 检查提现订单状态
SELECT * FROM balance_log_withdraw WHERE trade_no = 'xxx';

-- 检查状态流转
-- 1 (待处理) → 2 (处理中) → 3 (成功)
```

**解决**: 检查后台审核逻辑

### 问题 3: 渠道不显示

**症状**: 前端看不到支付/提现渠道

**排查**:

```sql
-- 检查渠道状态
SELECT * FROM site_payment WHERE status = 1;
SELECT * FROM site_withdraw WHERE status = 1;
```

**解决**: 确认渠道 status = 1 (启用)

## 九、信心保证

### 为什么可以放心？

1. **编译验证** ✅
   - Go 是强类型语言
   - 编译通过说明语法正确
   - 类型匹配正确

2. **常量使用** ✅
   - 所有状态都用常量
   - 避免魔法数字
   - IDE 可以追踪引用

3. **逻辑一致** ✅
   - 状态值定义清晰
   - 状态流转合理
   - 没有冲突

4. **可回滚** ✅
   - Git 版本控制
   - 可以快速回滚
   - 风险可控

### 建议的上线策略

1. **灰度发布**
   - 先在测试环境验证
   - 再在开发环境验证
   - 最后生产环境

2. **监控告警**
   - 监控错误日志
   - 监控订单状态
   - 监控用户反馈

3. **快速响应**
   - 准备回滚方案
   - 保持沟通畅通
   - 及时处理问题

---

**创建时间**: 2026-02-02
**验证状态**:

- ✅ 编译验证通过
- ✅ 代码审查通过
- ⏳ 功能测试待执行
- ⏳ 数据库验证待执行
