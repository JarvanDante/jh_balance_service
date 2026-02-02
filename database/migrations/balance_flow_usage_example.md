# 流水要求系统使用说明（简化版）

## 一、表结构说明

### 1. balance_flow_requirement（流水要求记录表）

**用途**：记录每个用户需要完成的流水要求

**核心字段**：

- `source_type`: 来源类型（充值、红利、活动等）
- `flow_multiple`: 流水倍数（1倍、3倍、5倍等）
- `required_flow`: 需要完成的总流水
- `completed_flow`: 已完成的流水
- `remaining_flow`: 剩余流水
- `game_flow_rates`: 游戏流水贡献率（从规则表复制，避免规则修改影响历史）
- `status`: 状态（进行中、已完成、已过期）

### 2. balance_flow_rule（流水规则配置表）

**用途**：配置不同场景的流水规则

**核心字段**：

- `rule_type`: 规则类型
- `flow_multiple`: 流水倍数
- `game_flow_rates`: 不同游戏的流水贡献率（JSON）
- `valid_game_types`: 有效游戏类型（JSON数组）
- `expire_days`: 有效天数

## 二、流水计算逻辑（在注单同步时处理）

### 定时任务同步注单时的处理流程：

```go
// 伪代码示例
func SyncBetOrders() {
    // 1. 从第三方拉取注单
    betOrders := fetchBetOrdersFromThirdParty()

    for _, bet := range betOrders {
        // 2. 保存注单到数据库
        saveBetOrder(bet)

        // 3. 查询该用户是否有进行中的流水要求
        flowRequirements := getActiveFlowRequirements(bet.UserId)

        if len(flowRequirements) > 0 {
            // 4. 计算并更新流水进度
            updateFlowProgress(bet, flowRequirements)
        }
    }
}

func updateFlowProgress(bet BetOrder, flowRequirements []FlowRequirement) {
    for _, flow := range flowRequirements {
        // 1. 检查游戏类型是否有效
        if !isValidGameType(bet.GameType, flow.ValidGameTypes) {
            continue
        }

        // 2. 获取该游戏类型的流水贡献率
        flowRate := getFlowRate(bet.GameType, flow.GameFlowRates)

        // 3. 计算流水贡献值
        flowContribution := bet.ValidBetAmount * flowRate / 100

        // 4. 更新流水记录
        updateFlowRequirement(flow.Id, flowContribution)
    }
}
```

### SQL 更新流水进度：

```sql
-- 更新流水进度
UPDATE balance_flow_requirement
SET completed_flow = completed_flow + ?,  -- 流水贡献值
    remaining_flow = required_flow - (completed_flow + ?),
    progress = ((completed_flow + ?) / required_flow) * 100,
    status = CASE
        WHEN (completed_flow + ?) >= required_flow THEN 2  -- 已完成
        ELSE status
    END,
    complete_time = CASE
        WHEN (completed_flow + ?) >= required_flow THEN NOW()
        ELSE complete_time
    END,
    updated_at = NOW()
WHERE id = ?
  AND status = 1;  -- 只更新进行中的记录
```

## 三、业务场景示例

### 场景1：用户充值100元，创建1倍流水要求

```sql
-- 1. 查询充值流水规则
SELECT * FROM balance_flow_rule
WHERE site_id = 1
  AND rule_type = 1  -- 充值流水
  AND status = 1
LIMIT 1;

-- 2. 创建流水要求记录（充值成功后）
INSERT INTO balance_flow_requirement (
    site_id, user_id, username,
    source_type, source_id, trade_no,
    amount, bonus_amount,
    flow_multiple, required_flow, remaining_flow,
    game_flow_rates, valid_game_types,
    status, start_time, expire_time
) VALUES (
    1, 10001, 'player001',
    1, 123456, 'D202602030001',
    100.00, 0.00,
    1.00, 100.00, 100.00,
    '{"slot": 100, "live": 50, "sport": 100, "fishing": 80, "lottery": 100, "chess": 60}',
    NULL,  -- 所有游戏都有效
    1, NOW(), DATE_ADD(NOW(), INTERVAL 30 DAY)
);
```

### 场景2：定时任务同步注单并更新流水

```sql
-- 假设用户在老虎机投注50元（100%计入流水）
-- 在同步注单时执行：

-- 1. 查询用户进行中的流水要求
SELECT id, game_flow_rates, valid_game_types, required_flow, completed_flow
FROM balance_flow_requirement
WHERE user_id = 10001
  AND status = 1  -- 进行中
  AND (expire_time IS NULL OR expire_time > NOW());

-- 2. 解析 game_flow_rates，获取 slot 的贡献率 = 100%
-- 3. 计算流水贡献 = 50 * 100% = 50

-- 4. 更新流水进度
UPDATE balance_flow_requirement
SET completed_flow = completed_flow + 50,
    remaining_flow = required_flow - (completed_flow + 50),
    progress = ((completed_flow + 50) / required_flow) * 100,
    status = CASE
        WHEN (completed_flow + 50) >= required_flow THEN 2
        ELSE 1
    END,
    complete_time = CASE
        WHEN (completed_flow + 50) >= required_flow THEN NOW()
        ELSE NULL
    END,
    updated_at = NOW()
WHERE id = 1
  AND status = 1;
```

### 场景3：用户在真人游戏投注100元（50%计入流水）

```sql
-- 在同步注单时：
-- 1. 获取 live 的贡献率 = 50%
-- 2. 计算流水贡献 = 100 * 50% = 50

-- 3. 更新流水进度
UPDATE balance_flow_requirement
SET completed_flow = completed_flow + 50,
    remaining_flow = remaining_flow - 50,
    progress = (completed_flow / required_flow) * 100,
    status = CASE
        WHEN completed_flow >= required_flow THEN 2
        ELSE 1
    END,
    complete_time = CASE
        WHEN completed_flow >= required_flow THEN NOW()
        ELSE NULL
    END,
    updated_at = NOW()
WHERE id = 1
  AND status = 1;

-- 此时：completed_flow = 100, remaining_flow = 0, status = 2（已完成）
```

## 四、常用查询SQL

### 1. 查询用户当前流水要求

```sql
SELECT
    id,
    CASE source_type
        WHEN 1 THEN '充值'
        WHEN 2 THEN '红利'
        WHEN 3 THEN '返水'
        WHEN 4 THEN '活动'
        WHEN 5 THEN '人工加款'
    END as source_name,
    amount,
    bonus_amount,
    flow_multiple,
    required_flow,
    completed_flow,
    remaining_flow,
    progress,
    CASE status
        WHEN 1 THEN '进行中'
        WHEN 2 THEN '已完成'
        WHEN 3 THEN '已过期'
        WHEN 4 THEN '已取消'
    END as status_name,
    start_time,
    expire_time,
    created_at
FROM balance_flow_requirement
WHERE user_id = 10001
  AND status = 1  -- 进行中
ORDER BY created_at DESC;
```

### 2. 检查用户是否可以提现

```sql
SELECT
    COUNT(*) as pending_count,
    SUM(remaining_flow) as total_remaining
FROM balance_flow_requirement
WHERE user_id = 10001
  AND status = 1  -- 进行中
  AND remaining_flow > 0;

-- 如果 pending_count > 0，则不允许提现
-- 或者提示用户还需完成 total_remaining 的流水
```

### 3. 自动过期处理（定时任务）

```sql
-- 将过期的流水要求标记为已过期
UPDATE balance_flow_requirement
SET status = 3,  -- 已过期
    updated_at = NOW()
WHERE status = 1  -- 进行中
  AND expire_time IS NOT NULL
  AND expire_time < NOW();
```

### 4. 统计用户流水完成情况

```sql
SELECT
    user_id,
    username,
    COUNT(*) as total_count,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as in_progress,
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as completed,
    SUM(required_flow) as total_required,
    SUM(completed_flow) as total_completed
FROM balance_flow_requirement
WHERE user_id = 10001
GROUP BY user_id, username;
```

## 五、定时任务实现建议

### 1. 注单同步任务（每分钟执行）

```go
func SyncBetOrdersTask() {
    // 1. 拉取最近的注单（根据时间戳）
    betOrders := fetchRecentBetOrders()

    // 2. 批量处理
    for _, bet := range betOrders {
        // 保存注单
        saveBetOrder(bet)

        // 更新流水进度（异步处理，避免阻塞）
        go updateUserFlowProgress(bet)
    }
}

func updateUserFlowProgress(bet BetOrder) {
    // 1. 查询用户进行中的流水要求
    flows := getActiveFlowRequirements(bet.UserId)

    if len(flows) == 0 {
        return
    }

    // 2. 遍历每个流水要求
    for _, flow := range flows {
        // 检查游戏类型是否有效
        if !isValidGame(bet.GameType, flow.ValidGameTypes) {
            continue
        }

        // 获取流水贡献率
        rate := getGameFlowRate(bet.GameType, flow.GameFlowRates)

        // 计算流水贡献
        contribution := bet.ValidBetAmount * rate / 100

        // 更新流水记录
        updateFlowRecord(flow.Id, contribution)
    }
}
```

### 2. 流水过期检查任务（每小时执行）

```go
func CheckExpiredFlowTask() {
    // 将过期的流水要求标记为已过期
    db.Exec(`
        UPDATE balance_flow_requirement
        SET status = 3, updated_at = NOW()
        WHERE status = 1
          AND expire_time IS NOT NULL
          AND expire_time < NOW()
    `)
}
```

## 六、游戏流水贡献率说明

```json
{
  "slot": 100, // 老虎机：100%计入流水
  "live": 50, // 真人游戏：50%计入流水
  "sport": 100, // 体育博彩：100%计入流水
  "fishing": 80, // 捕鱼游戏：80%计入流水
  "lottery": 100, // 彩票：100%计入流水
  "chess": 60 // 棋牌：60%计入流水
}
```

**计算公式**：

```
流水贡献 = 有效投注金额 × 游戏流水贡献率 ÷ 100
```

**示例**：

- 老虎机投注100元 → 流水贡献 = 100 × 100% = 100元
- 真人游戏投注100元 → 流水贡献 = 100 × 50% = 50元
- 捕鱼游戏投注100元 → 流水贡献 = 100 × 80% = 80元

## 七、关键点总结

1. ✅ **在注单同步时计算流水**：定时任务拉取注单后，立即计算并更新流水进度
2. ✅ **无需明细表**：直接在 `balance_flow_requirement` 表中累加 `completed_flow`
3. ✅ **规则快照**：创建流水要求时，将规则配置复制到记录中，避免规则修改影响历史
4. ✅ **并发处理**：注单同步可以异步更新流水，提高性能
5. ✅ **提现控制**：提现前检查是否有未完成的流水要求

## 八、完整流程图

```
充值成功
   ↓
查询流水规则（balance_flow_rule）
   ↓
创建流水要求（balance_flow_requirement）
   ↓
用户开始游戏
   ↓
定时任务拉取注单
   ↓
查询用户进行中的流水要求
   ↓
计算流水贡献（根据游戏类型和贡献率）
   ↓
更新流水进度（completed_flow += 贡献值）
   ↓
检查是否完成（completed_flow >= required_flow）
   ↓
完成 → status = 2, complete_time = NOW()
未完成 → 继续累计
   ↓
用户申请提现
   ↓
检查是否有未完成的流水
   ↓
有 → 拒绝提现或提示剩余流水
无 → 允许提现
```

## 九、性能优化方案

### ⚠️ 问题：同步处理会很卡

如果在保存注单时同步计算流水，会导致：

- 注单同步变慢
- 阻塞主流程
- 影响系统性能

### ✅ 推荐方案

#### 方案1：异步处理（简单）

```go
func SyncBetOrdersTask() {
    betOrders := fetchRecentBetOrders()

    for _, bet := range betOrders {
        // 1. 先保存注单（快速完成）
        saveBetOrder(bet)

        // 2. 异步更新流水（不阻塞）
        go func(b BetOrder) {
            updateUserFlowProgress(b)
        }(bet)
    }
}
```

**优点**：

- 实现简单
- 不阻塞主流程
- 并发处理快

**缺点**：

- goroutine 数量不可控
- 没有重试机制

#### 方案2：Worker Pool（推荐）

```go
// 创建 Worker Pool
type FlowUpdateWorker struct {
    jobQueue chan BetOrder
    workers  int
}

func NewFlowUpdateWorker(workers int) *FlowUpdateWorker {
    w := &FlowUpdateWorker{
        jobQueue: make(chan BetOrder, 1000),
        workers:  workers,
    }

    // 启动 workers
    for i := 0; i < workers; i++ {
        go w.worker()
    }

    return w
}

func (w *FlowUpdateWorker) worker() {
    for bet := range w.jobQueue {
        updateUserFlowProgress(bet)
    }
}

func (w *FlowUpdateWorker) Submit(bet BetOrder) {
    w.jobQueue <- bet
}

// 使用
var flowWorker = NewFlowUpdateWorker(10) // 10个并发worker

func SyncBetOrdersTask() {
    betOrders := fetchRecentBetOrders()

    for _, bet := range betOrders {
        saveBetOrder(bet)
        flowWorker.Submit(bet) // 提交到 worker pool
    }
}
```

**优点**：

- 控制并发数量
- 避免 goroutine 爆炸
- 性能稳定

#### 方案3：消息队列（最佳）

```go
// 1. 注单同步时发送消息
func SyncBetOrdersTask() {
    betOrders := fetchRecentBetOrders()

    for _, bet := range betOrders {
        // 保存注单
        saveBetOrder(bet)

        // 发送到 RabbitMQ/Redis Stream
        publishFlowUpdateMessage(bet)
    }
}

// 2. 独立的消费者处理流水更新
func FlowUpdateConsumer() {
    for {
        msg := consumeFlowUpdateMessage()
        bet := parseBet(msg)

        // 更新流水
        err := updateUserFlowProgress(bet)
        if err != nil {
            // 重试或记录失败
            retryOrLog(msg, err)
        } else {
            // 确认消息
            ackMessage(msg)
        }
    }
}
```

**优点**：

- 完全解耦
- 可靠性高（消息持久化）
- 支持重试
- 可以水平扩展消费者

**适用场景**：

- 注单量大（每分钟 > 1000 条）
- 需要高可靠性
- 已有消息队列基础设施

#### 方案4：批量处理（高性能）

```go
func SyncBetOrdersTask() {
    betOrders := fetchRecentBetOrders()

    // 1. 批量保存注单
    batchSaveBetOrders(betOrders)

    // 2. 按用户分组
    userBets := make(map[int64][]BetOrder)
    for _, bet := range betOrders {
        userBets[bet.UserId] = append(userBets[bet.UserId], bet)
    }

    // 3. 批量更新流水
    go batchUpdateFlowProgress(userBets)
}

func batchUpdateFlowProgress(userBets map[int64][]BetOrder) {
    for userId, bets := range userBets {
        // 查询该用户的流水要求
        flows := getActiveFlowRequirements(userId)
        if len(flows) == 0 {
            continue
        }

        // 计算总流水贡献
        totalContribution := 0.0
        for _, bet := range bets {
            for _, flow := range flows {
                rate := getGameFlowRate(bet.GameType, flow.GameFlowRates)
                totalContribution += bet.ValidBetAmount * rate / 100
            }
        }

        // 一次性更新
        updateFlowRecord(flows[0].Id, totalContribution)
    }
}
```

**优点**：

- 减少数据库查询
- 批量更新效率高
- 适合大量注单

### 📊 性能对比

| 方案        | 注单量      | 处理时间 | 并发控制 | 可靠性 | 复杂度 |
| ----------- | ----------- | -------- | -------- | ------ | ------ |
| 同步处理    | 100/分钟    | 慢       | ❌       | 高     | 低     |
| 异步处理    | 1000/分钟   | 快       | ⚠️       | 中     | 低     |
| Worker Pool | 5000/分钟   | 快       | ✅       | 中     | 中     |
| 消息队列    | 10000+/分钟 | 快       | ✅       | 高     | 高     |
| 批量处理    | 5000/分钟   | 很快     | ✅       | 中     | 中     |

### 🎯 推荐选择

**小型项目（< 1000 注单/分钟）**：

- 使用 **Worker Pool** 方案
- 简单可靠，性能足够

**中型项目（1000-5000 注单/分钟）**：

- 使用 **Worker Pool + 批量处理**
- 平衡性能和复杂度

**大型项目（> 5000 注单/分钟）**：

- 使用 **消息队列**
- 高可靠性，可水平扩展

### 💡 实现建议

```go
// 推荐实现：Worker Pool + 批量处理

// 1. 初始化 Worker Pool
var flowWorker *FlowUpdateWorker

func init() {
    flowWorker = NewFlowUpdateWorker(20) // 20个并发worker
}

// 2. 注单同步
func SyncBetOrdersTask() {
    betOrders := fetchRecentBetOrders()

    // 批量保存注单
    batchSaveBetOrders(betOrders)

    // 按用户分组并提交到 worker
    userBets := groupBetsByUser(betOrders)
    for userId, bets := range userBets {
        flowWorker.Submit(FlowUpdateJob{
            UserId: userId,
            Bets:   bets,
        })
    }
}

// 3. Worker 处理
func (w *FlowUpdateWorker) worker() {
    for job := range w.jobQueue {
        // 批量计算该用户的流水
        updateUserFlowBatch(job.UserId, job.Bets)
    }
}

func updateUserFlowBatch(userId int64, bets []BetOrder) {
    // 查询用户流水要求
    flows := getActiveFlowRequirements(userId)
    if len(flows) == 0 {
        return
    }

    // 计算总贡献
    for _, flow := range flows {
        totalContribution := 0.0
        for _, bet := range bets {
            if !isValidGame(bet.GameType, flow.ValidGameTypes) {
                continue
            }
            rate := getGameFlowRate(bet.GameType, flow.GameFlowRates)
            totalContribution += bet.ValidBetAmount * rate / 100
        }

        // 更新流水
        if totalContribution > 0 {
            updateFlowRecord(flow.Id, totalContribution)
        }
    }
}
```

### ⚡ 性能优化技巧

1. **使用索引**：

```sql
-- 确保有这些索引
CREATE INDEX idx_site_user_status ON balance_flow_requirement(site_id, user_id, status);
CREATE INDEX idx_expire_time ON balance_flow_requirement(expire_time);
```

2. **批量更新**：

```sql
-- 使用 CASE WHEN 批量更新多个用户
UPDATE balance_flow_requirement
SET completed_flow = CASE id
    WHEN 1 THEN completed_flow + 50
    WHEN 2 THEN completed_flow + 30
    WHEN 3 THEN completed_flow + 80
END,
remaining_flow = required_flow - completed_flow,
updated_at = NOW()
WHERE id IN (1, 2, 3);
```

3. **缓存流水要求**：

```go
// 使用 Redis 缓存用户的流水要求
func getActiveFlowRequirements(userId int64) []FlowRequirement {
    // 先从 Redis 获取
    cacheKey := fmt.Sprintf("flow:user:%d", userId)
    if cached := redis.Get(cacheKey); cached != nil {
        return parseFlowRequirements(cached)
    }

    // 从数据库查询
    flows := queryFromDB(userId)

    // 缓存 5 分钟
    redis.Set(cacheKey, flows, 5*time.Minute)

    return flows
}
```

4. **限流保护**：

```go
// 使用令牌桶限流
var limiter = rate.NewLimiter(1000, 2000) // 每秒1000个，突发2000个

func updateUserFlowProgress(bet BetOrder) {
    // 等待令牌
    if err := limiter.Wait(context.Background()); err != nil {
        log.Error("rate limit exceeded")
        return
    }

    // 执行更新
    doUpdate(bet)
}
```
