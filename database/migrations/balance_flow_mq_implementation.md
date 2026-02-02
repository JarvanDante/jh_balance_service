# 流水要求系统 - MQ 实现方案

## 一、架构设计

```
注单同步服务                    MQ (RabbitMQ)              流水更新服务
    │                              │                          │
    ├─ 1. 拉取注单                 │                          │
    ├─ 2. 保存注单到DB             │                          │
    ├─ 3. 发送消息到MQ ──────────> │                          │
    │                              │                          │
    │                              ├─ Queue: flow_update      │
    │                              │   - user_id              │
    │                              │   - bet_id               │
    │                              │   - game_type            │
    │                              │   - valid_bet_amount     │
    │                              │                          │
    │                              │ <──────────────────────  ├─ 4. 消费消息
    │                              │                          ├─ 5. 查询流水要求
    │                              │                          ├─ 6. 计算流水贡献
    │                              │                          ├─ 7. 更新流水记录
    │                              │                          ├─ 8. ACK 消息
```

## 二、消息结构设计

### 消息格式（JSON）

```json
{
  "message_id": "msg_20260203_123456",
  "timestamp": 1706918400,
  "data": {
    "site_id": 1,
    "user_id": 10001,
    "username": "player001",
    "bet_id": "BET202602030001",
    "game_type": "slot",
    "game_name": "幸运老虎机",
    "bet_amount": 100.00,
    "valid_bet_amount": 100.00,
    "bet_time": "2026-02-03 12:00:00",
    "settle_time": "2026-02-03 12:00:05"
  }
}
```

### Go 结构体定义

```go
// FlowUpdateMessage 流水更新消息
type FlowUpdateMessage struct {
    MessageID string    `json:"message_id"`
    Timestamp int64     `json:"timestamp"`
    Data      BetData   `json:"data"`
}

type BetData struct {
    SiteId          int64   `json:"site_id"`
    UserId          int64   `json:"user_id"`
    Username        string  `json:"username"`
    BetId           string  `json:"bet_id"`
    GameType        string  `json:"game_type"`
    GameName        string  `json:"game_name"`
    BetAmount       float64 `json:"bet_amount"`
    ValidBetAmount  float64 `json:"valid_bet_amount"`
    BetTime         string  `json:"bet_time"`
    SettleTime      string  `json:"settle_time"`
}
```

## 三、生产者实现（注单同步服务）

### 1. 初始化 RabbitMQ 连接

```go
package mq

import (
    "encoding/json"
    "fmt"
    "github.com/streadway/amqp"
    "log"
)

type FlowUpdateProducer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    queue   string
}

func NewFlowUpdateProducer(amqpURL, queueName string) (*FlowUpdateProducer, error) {
    // 连接 RabbitMQ
    conn, err := amqp.Dial(amqpURL)
    if err != nil {
        return nil, fmt.Errorf("连接 RabbitMQ 失败: %v", err)
    }

    // 创建 Channel
    channel, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("创建 Channel 失败: %v", err)
    }

    // 声明队列
    _, err = channel.QueueDeclare(
        queueName, // 队列名称
        true,      // durable: 持久化
        false,     // autoDelete
        false,     // exclusive
        false,     // noWait
        nil,       // arguments
    )
    if err != nil {
        channel.Close()
        conn.Close()
        return nil, fmt.Errorf("声明队列失败: %v", err)
    }

    return &FlowUpdateProducer{
        conn:    conn,
        channel: channel,
        queue:   queueName,
    }, nil
}

func (p *FlowUpdateProducer) Publish(msg *FlowUpdateMessage) error {
    // 序列化消息
    body, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("序列化消息失败: %v", err)
    }

    // 发布消息
    err = p.channel.Publish(
        "",       // exchange
        p.queue,  // routing key
        false,    // mandatory
        false,    // immediate
        amqp.Publishing{
            DeliveryMode: amqp.Persistent, // 持久化消息
            ContentType:  "application/json",
            Body:         body,
        },
    )
    if err != nil {
        return fmt.Errorf("发布消息失败: %v", err)
    }

    log.Printf("发送流水更新消息: user_id=%d, bet_id=%s", msg.Data.UserId, msg.Data.BetId)
    return nil
}

func (p *FlowUpdateProducer) Close() {
    if p.channel != nil {
        p.channel.Close()
    }
    if p.conn != nil {
        p.conn.Close()
    }
}
```

### 2. 注单同步时发送消息

```go
package sync

import (
    "fmt"
    "time"
)

var flowProducer *mq.FlowUpdateProducer

func init() {
    // 初始化生产者
    var err error
    flowProducer, err = mq.NewFlowUpdateProducer(
        "amqp://guest:guest@localhost:5672/",
        "flow_update_queue",
    )
    if err != nil {
        panic(fmt.Sprintf("初始化 MQ 生产者失败: %v", err))
    }
}

// SyncBetOrdersTask 同步注单任务
func SyncBetOrdersTask() {
    // 1. 从第三方拉取注单
    betOrders := fetchBetOrdersFromThirdParty()
    
    // 2. 批量保存注单到数据库
    err := batchSaveBetOrders(betOrders)
    if err != nil {
        log.Errorf("保存注单失败: %v", err)
        return
    }
    
    // 3. 发送消息到 MQ
    for _, bet := range betOrders {
        msg := &mq.FlowUpdateMessage{
            MessageID: fmt.Sprintf("msg_%d_%s", time.Now().Unix(), bet.BetId),
            Timestamp: time.Now().Unix(),
            Data: mq.BetData{
                SiteId:         bet.SiteId,
                UserId:         bet.UserId,
                Username:       bet.Username,
                BetId:          bet.BetId,
                GameType:       bet.GameType,
                GameName:       bet.GameName,
                BetAmount:      bet.BetAmount,
                ValidBetAmount: bet.ValidBetAmount,
                BetTime:        bet.BetTime,
                SettleTime:     bet.SettleTime,
            },
        }
        
        // 发送消息（异步，不阻塞）
        go func(m *mq.FlowUpdateMessage) {
            if err := flowProducer.Publish(m); err != nil {
                log.Errorf("发送流水更新消息失败: %v", err)
                // 可以记录到失败队列，后续重试
            }
        }(msg)
    }
    
    log.Infof("同步注单完成，共 %d 条", len(betOrders))
}
```

## 四、消费者实现（流水更新服务）

### 1. 消费者结构

```go
package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/streadway/amqp"
    "log"
    "time"
)

type FlowUpdateConsumer struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    queue    string
    workers  int
}

func NewFlowUpdateConsumer(amqpURL, queueName string, workers int) (*FlowUpdateConsumer, error) {
    conn, err := amqp.Dial(amqpURL)
    if err != nil {
        return nil, fmt.Errorf("连接 RabbitMQ 失败: %v", err)
    }

    channel, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("创建 Channel 失败: %v", err)
    }

    // 设置 QoS（每次只预取 1 条消息）
    err = channel.Qos(
        1,     // prefetch count
        0,     // prefetch size
        false, // global
    )
    if err != nil {
        channel.Close()
        conn.Close()
        return nil, fmt.Errorf("设置 QoS 失败: %v", err)
    }

    return &FlowUpdateConsumer{
        conn:    conn,
        channel: channel,
        queue:   queueName,
        workers: workers,
    }, nil
}

func (c *FlowUpdateConsumer) Start(ctx context.Context) error {
    // 获取消息通道
    msgs, err := c.channel.Consume(
        c.queue, // queue
        "",      // consumer
        false,   // auto-ack: 手动确认
        false,   // exclusive
        false,   // no-local
        false,   // no-wait
        nil,     // args
    )
    if err != nil {
        return fmt.Errorf("开始消费失败: %v", err)
    }

    log.Printf("流水更新消费者启动，workers=%d", c.workers)

    // 启动多个 worker
    for i := 0; i < c.workers; i++ {
        go c.worker(ctx, i, msgs)
    }

    // 等待退出信号
    <-ctx.Done()
    return nil
}

func (c *FlowUpdateConsumer) worker(ctx context.Context, id int, msgs <-chan amqp.Delivery) {
    log.Printf("Worker %d 启动", id)

    for {
        select {
        case <-ctx.Done():
            log.Printf("Worker %d 退出", id)
            return
        case msg, ok := <-msgs:
            if !ok {
                log.Printf("Worker %d 消息通道关闭", id)
                return
            }

            // 处理消息
            if err := c.handleMessage(&msg); err != nil {
                log.Errorf("Worker %d 处理消息失败: %v", id, err)
                // 拒绝消息，重新入队
                msg.Nack(false, true)
            } else {
                // 确认消息
                msg.Ack(false)
            }
        }
    }
}

func (c *FlowUpdateConsumer) handleMessage(msg *amqp.Delivery) error {
    // 解析消息
    var flowMsg mq.FlowUpdateMessage
    if err := json.Unmarshal(msg.Body, &flowMsg); err != nil {
        return fmt.Errorf("解析消息失败: %v", err)
    }

    log.Printf("处理流水更新: user_id=%d, bet_id=%s", 
        flowMsg.Data.UserId, flowMsg.Data.BetId)

    // 更新流水
    return updateUserFlowProgress(&flowMsg.Data)
}

func (c *FlowUpdateConsumer) Close() {
    if c.channel != nil {
        c.channel.Close()
    }
    if c.conn != nil {
        c.conn.Close()
    }
}
```

### 2. 流水更新逻辑

```go
func updateUserFlowProgress(bet *mq.BetData) error {
    ctx := context.Background()

    // 1. 查询用户进行中的流水要求
    flows, err := getActiveFlowRequirements(ctx, bet.UserId)
    if err != nil {
        return fmt.Errorf("查询流水要求失败: %v", err)
    }

    if len(flows) == 0 {
        log.Printf("用户 %d 没有进行中的流水要求", bet.UserId)
        return nil
    }

    // 2. 遍历每个流水要求
    for _, flow := range flows {
        // 检查游戏类型是否有效
        if !isValidGameType(bet.GameType, flow.ValidGameTypes) {
            log.Printf("游戏类型 %s 不在有效范围内", bet.GameType)
            continue
        }

        // 获取流水贡献率
        rate := getGameFlowRate(bet.GameType, flow.GameFlowRates)
        if rate == 0 {
            log.Printf("游戏类型 %s 流水贡献率为 0", bet.GameType)
            continue
        }

        // 计算流水贡献
        contribution := bet.ValidBetAmount * rate / 100

        // 更新流水记录
        err = updateFlowRecord(ctx, flow.Id, contribution)
        if err != nil {
            return fmt.Errorf("更新流水记录失败: %v", err)
        }

        log.Printf("更新流水成功: flow_id=%d, contribution=%.2f", flow.Id, contribution)
    }

    return nil
}

func updateFlowRecord(ctx context.Context, flowId int64, contribution float64) error {
    sql := `
        UPDATE balance_flow_requirement
        SET completed_flow = completed_flow + ?,
            remaining_flow = GREATEST(required_flow - (completed_flow + ?), 0),
            progress = LEAST(((completed_flow + ?) / required_flow) * 100, 100),
            status = CASE 
                WHEN (completed_flow + ?) >= required_flow THEN 2
                ELSE status 
            END,
            complete_time = CASE 
                WHEN (completed_flow + ?) >= required_flow AND complete_time IS NULL THEN NOW()
                ELSE complete_time 
            END,
            updated_at = NOW()
        WHERE id = ?
          AND status = 1
    `

    result, err := db.ExecContext(ctx, sql, 
        contribution, contribution, contribution, 
        contribution, contribution, flowId)
    if err != nil {
        return err
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        log.Printf("流水记录 %d 未更新（可能已完成或不存在）", flowId)
    }

    return nil
}
```

### 3. 启动消费者

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // 创建消费者
    consumer, err := consumer.NewFlowUpdateConsumer(
        "amqp://guest:guest@localhost:5672/",
        "flow_update_queue",
        10, // 10 个并发 worker
    )
    if err != nil {
        log.Fatalf("创建消费者失败: %v", err)
    }
    defer consumer.Close()

    // 创建上下文
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 启动消费者
    go func() {
        if err := consumer.Start(ctx); err != nil {
            log.Fatalf("启动消费者失败: %v", err)
        }
    }()

    // 等待退出信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    log.Println("收到退出信号，正在关闭...")
    cancel()
}
```

## 五、配置示例

### RabbitMQ 配置

```yaml
# config.yaml
rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
  queues:
    flow_update:
      name: "flow_update_queue"
      durable: true
      prefetch: 1
      workers: 10
```

### Docker Compose 启动 RabbitMQ

```yaml
version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3-management
    container_name: rabbitmq
    ports:
      - "5672:5672"   # AMQP 端口
      - "15672:15672" # 管理界面
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

volumes:
  rabbitmq_data:
```

## 六、监控和告警

### 1. 监控指标

```go
// 监控指标
type FlowUpdateMetrics struct {
    MessagesSent      int64 // 发送的消息数
    MessagesProcessed int64 // 处理的消息数
    MessagesFailed    int64 // 失败的消息数
    ProcessingTime    time.Duration // 平均处理时间
}

// Prometheus 指标
var (
    messagesSent = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flow_update_messages_sent_total",
        Help: "Total number of flow update messages sent",
    })

    messagesProcessed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flow_update_messages_processed_total",
        Help: "Total number of flow update messages processed",
    })

    messagesFailed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flow_update_messages_failed_total",
        Help: "Total number of flow update messages failed",
    })

    processingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "flow_update_processing_duration_seconds",
        Help: "Flow update processing duration in seconds",
    })
)
```

### 2. 死信队列（DLQ）

```go
// 声明死信队列
func declareDLQ(channel *amqp.Channel) error {
    // 死信交换机
    err := channel.ExchangeDeclare(
        "flow_update_dlx", // name
        "direct",          // type
        true,              // durable
        false,             // auto-deleted
        false,             // internal
        false,             // no-wait
        nil,               // arguments
    )
    if err != nil {
        return err
    }

    // 死信队列
    _, err = channel.QueueDeclare(
        "flow_update_dlq", // name
        true,              // durable
        false,             // delete when unused
        false,             // exclusive
        false,             // no-wait
        nil,               // arguments
    )
    if err != nil {
        return err
    }

    // 绑定
    err = channel.QueueBind(
        "flow_update_dlq", // queue name
        "flow_update",     // routing key
        "flow_update_dlx", // exchange
        false,
        nil,
    )
    return err
}
```

## 七、优势总结

✅ **解耦**：注单同步和流水更新完全分离
✅ **可靠**：消息持久化，不会丢失
✅ **可扩展**：可以水平扩展消费者数量
✅ **重试**：失败消息可以重新入队
✅ **监控**：可以监控队列长度、处理速度等
✅ **异步**：不阻塞注单同步流程
✅ **顺序**：可以保证同一用户的消息顺序处理（使用 routing key）

## 八、注意事项

1. **幂等性**：确保重复消费不会导致流水重复计算
2. **消息去重**：可以使用 Redis 记录已处理的 bet_id
3. **错误处理**：失败消息进入死信队列，人工介入
4. **监控告警**：监控队列积压、消费速度、失败率
5. **资源限制**：控制消费者数量，避免数据库压力过大
