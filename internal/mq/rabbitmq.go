package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	amqp "github.com/rabbitmq/amqp091-go"
	v1 "jh_balance_service/api/frontend/balance/v1"
	"jh_balance_service/internal/middleware"
	"jh_balance_service/internal/model/entity"
	"jh_balance_service/internal/service/frontend"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
)

// InitBalanceUserMessage 初始化余额用户消息
type InitBalanceUserMessage struct {
	SiteID   int32  `json:"site_id"`
	UserID   int32  `json:"user_id"`
	Username string `json:"username"`
	Currency string `json:"currency"`
}

// BetSettledMessage 注单结算消息（用于流水消费）
type BetSettledMessage struct {
	SiteID      int32   `json:"site_id"`
	SiteCode    string  `json:"site_code"`
	UserID      int64   `json:"user_id"`
	Username    string  `json:"username"`
	Platform    string  `json:"platform"`
	GameType    string  `json:"game_type"`
	OrderNo     string  `json:"order_no"`
	ValidAmount float64 `json:"valid_amount"`
	BetTime     string  `json:"bet_time"`
	SettleTime  string  `json:"settle_time"`
	Status      uint8   `json:"status"`
}

// InitRabbitMQ 初始化RabbitMQ连接
func InitRabbitMQ(ctx context.Context) error {
	url := g.Cfg().MustGet(ctx, "rabbitmq.url").String()
	if url == "" {
		g.Log().Warning(ctx, "RabbitMQ URL未配置，跳过初始化")
		return nil
	}

	var err error
	conn, err = amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("连接RabbitMQ失败: %v", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("创建RabbitMQ channel失败: %v", err)
	}

	// 声明 exchange
	err = channel.ExchangeDeclare(
		"user.created", // exchange name
		"topic",        // exchange type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("声明exchange失败: %v", err)
	}

	// 声明队列
	queue, err := channel.QueueDeclare(
		"balance.init_user", // 队列名称
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 绑定队列到 exchange
	err = channel.QueueBind(
		queue.Name,          // queue name
		"user.init_balance", // routing key
		"user.created",      // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("绑定队列失败: %v", err)
	}

	g.Log().Info(ctx, "RabbitMQ初始化成功")
	return nil
}

// CloseRabbitMQ 关闭RabbitMQ连接
func CloseRabbitMQ() {
	if channel != nil {
		channel.Close()
	}
	if conn != nil {
		conn.Close()
	}
}

// StartConsumer 启动消费者
func StartConsumer(ctx context.Context) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel未初始化")
	}

	// 设置 QoS
	err := channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("设置QoS失败: %v", err)
	}

	// 开始消费
	msgs, err := channel.Consume(
		"balance.init_user", // queue
		"",                  // consumer
		false,               // auto-ack (设置为false，手动确认)
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		return fmt.Errorf("开始消费失败: %v", err)
	}

	g.Log().Info(ctx, "RabbitMQ消费者启动成功，等待消息...")

	// 启动消费协程
	go func() {
		for msg := range msgs {
			if err := handleInitBalanceUser(ctx, msg); err != nil {
				g.Log().Errorf(ctx, "处理消息失败: %v, 消息ID: %s", err, msg.MessageId)
				// 消息处理失败，拒绝并重新入队
				msg.Nack(false, true)
			} else {
				// 消息处理成功，确认
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// StartFlowConsumeConsumer 启动流水消费消费者（按用户串行，跨用户并行）
func StartFlowConsumeConsumer(ctx context.Context) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel未初始化")
	}

	// 声明 exchange
	err := channel.ExchangeDeclare(
		"bet.settled", // exchange name
		"topic",       // exchange type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return fmt.Errorf("声明exchange失败: %v", err)
	}

	// 声明队列
	queue, err := channel.QueueDeclare(
		"balance.flow_consume", // queue name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 绑定队列到 exchange
	err = channel.QueueBind(
		queue.Name,               // queue name
		"bet.settled_to_balance", // routing key
		"bet.settled",            // exchange
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return fmt.Errorf("绑定队列失败: %v", err)
	}

	// 设置 QoS
	err = channel.Qos(
		1000, // prefetch count
		0,    // prefetch size
		false,
	)
	if err != nil {
		return fmt.Errorf("设置QoS失败: %v", err)
	}

	// 开始消费
	msgs, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("开始消费失败: %v", err)
	}

	g.Log().Info(ctx, "流水消费消费者启动成功，等待消息...")

	// 启动分片工作协程
	const workerCount = 16
	workers := startFlowConsumeWorkers(ctx, workerCount)

	// 分发消息
	go func() {
		for msg := range msgs {
			var betMsg BetSettledMessage
			if err := json.Unmarshal(msg.Body, &betMsg); err != nil {
				g.Log().Errorf(ctx, "解析注单消息失败: %v, body: %s", err, string(msg.Body))
				msg.Nack(false, false)
				continue
			}

			if betMsg.SiteID <= 0 || betMsg.UserID <= 0 || betMsg.OrderNo == "" || betMsg.SiteCode == "" {
				g.Log().Warningf(ctx, "无效注单消息，跳过: site_id=%d, user_id=%d, order_no=%s, site_code=%s",
					betMsg.SiteID, betMsg.UserID, betMsg.OrderNo, betMsg.SiteCode)
				msg.Ack(false)
				continue
			}
			if betMsg.Status != 1 || betMsg.ValidAmount <= 0 {
				g.Log().Infof(ctx, "注单消息被过滤: order_no=%s, status=%d, valid_amount=%.4f",
					betMsg.OrderNo, betMsg.Status, betMsg.ValidAmount)
				msg.Ack(false)
				continue
			}

			g.Log().Infof(ctx, "收到流水消费消息: order_no=%s, user_id=%d, site=%s, valid_amount=%.4f",
				betMsg.OrderNo, betMsg.UserID, betMsg.SiteCode, betMsg.ValidAmount)

			shard := flowConsumeShard(betMsg.SiteID, betMsg.UserID, workerCount)
			workers[shard] <- flowConsumeDelivery{msg: betMsg, delivery: msg}
		}
	}()

	return nil
}

type flowConsumeDelivery struct {
	msg      BetSettledMessage
	delivery amqp.Delivery
}

func flowConsumeShard(siteID int32, userID int64, workerCount int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%d", siteID, userID)))
	return int(h.Sum32()) % workerCount
}

func startFlowConsumeWorkers(ctx context.Context, workerCount int) []chan flowConsumeDelivery {
	workers := make([]chan flowConsumeDelivery, workerCount)
	for i := 0; i < workerCount; i++ {
		ch := make(chan flowConsumeDelivery, 2000)
		workers[i] = ch
		go flowConsumeWorker(ctx, i, ch)
	}
	return workers
}

func flowConsumeWorker(ctx context.Context, idx int, ch <-chan flowConsumeDelivery) {
	const (
		batchSize    = 200
		flushTimeout = 1 * time.Second
	)

	var batch []flowConsumeDelivery
	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		// 按用户分组（同用户串行）
		type userKey struct {
			siteID   int32
			userID   int64
			siteCode string
		}
		groupMap := make(map[userKey][]flowConsumeDelivery)
		order := make([]userKey, 0, 32)

		for _, d := range batch {
			key := userKey{siteID: d.msg.SiteID, userID: d.msg.UserID, siteCode: d.msg.SiteCode}
			if _, ok := groupMap[key]; !ok {
				order = append(order, key)
				groupMap[key] = make([]flowConsumeDelivery, 0, 4)
			}
			groupMap[key] = append(groupMap[key], d)
		}

		for _, key := range order {
			deliveries := groupMap[key]
			if err := processUserFlowBatch(ctx, key.siteID, key.userID, key.siteCode, deliveries); err != nil {
				g.Log().Errorf(ctx, "流水消费失败: site_id=%d, user_id=%d, err=%v", key.siteID, key.userID, err)
				for _, d := range deliveries {
					d.delivery.Nack(false, true)
				}
			} else {
				for _, d := range deliveries {
					d.delivery.Ack(false)
				}
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case d, ok := <-ch:
			if !ok {
				flushBatch()
				return
			}
			batch = append(batch, d)
			if len(batch) >= batchSize {
				flushBatch()
				ticker.Reset(flushTimeout)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch()
			}
		}
	}
}

func processUserFlowBatch(ctx context.Context, siteID int32, userID int64, siteCode string, deliveries []flowConsumeDelivery) error {
	groupName := "site_" + siteCode
	if gdb.GetConfig(groupName) == nil {
		return fmt.Errorf("动态数据库组不存在: %s", groupName)
	}

	siteCtx := context.WithValue(ctx, "dynamic_db_group", groupName)

	orderNos := make([]string, 0, len(deliveries))
	for _, d := range deliveries {
		orderNos = append(orderNos, d.msg.OrderNo)
	}

	// 查询已存在的订单号（幂等）
	var existed []struct {
		OrderNo string `json:"order_no"`
	}
	if err := g.DB(groupName).Model("balance_flow_consumed_bets").
		Where("site_id", siteID).
		Where("user_id", userID).
		WhereIn("order_no", orderNos).
		Fields("order_no").
		Scan(&existed); err != nil {
		return err
	}
	existedSet := make(map[string]struct{}, len(existed))
	for _, e := range existed {
		existedSet[e.OrderNo] = struct{}{}
	}

	// 过滤新订单
	newDeliveries := make([]flowConsumeDelivery, 0, len(deliveries))
	for _, d := range deliveries {
		if _, ok := existedSet[d.msg.OrderNo]; ok {
			continue
		}
		newDeliveries = append(newDeliveries, d)
	}
	if len(newDeliveries) == 0 {
		return nil
	}

	// 批量插入幂等记录
	insertData := make([]g.Map, 0, len(newDeliveries))
	now := gtime.Now()
	for _, d := range newDeliveries {
		insertData = append(insertData, g.Map{
			"site_id":     siteID,
			"user_id":     userID,
			"order_no":    d.msg.OrderNo,
			"consumed_at": now,
		})
	}
	if _, err := g.DB(groupName).Model("balance_flow_consumed_bets").Data(insertData).InsertIgnore(); err != nil {
		return err
	}

	// FIFO 消费流水（逐条事务，失败写入 retry 表并继续）
	for _, d := range newDeliveries {
		err := g.DB(groupName).Transaction(siteCtx, func(ctx context.Context, tx gdb.TX) error {
			return ApplyFlowRequirementTx(ctx, tx, int(siteID), userID, d.msg.ValidAmount)
		})
		if err != nil {
			g.Log().Errorf(ctx, "流水消费失败，写入重试表: order_no=%s, user_id=%d, site=%s, err=%v",
				d.msg.OrderNo, userID, siteCode, err)
			_ = insertFlowConsumeRetry(siteCtx, groupName, d.msg, err)
			// 继续处理其它订单
			continue
		}
	}

	return nil
}

func ApplyFlowRequirementTx(ctx context.Context, tx gdb.TX, siteID int, userID int64, amount float64) error {
	if amount <= 0 {
		return nil
	}

	var rules []*entity.BalanceFlowRequirement
	now := gtime.Now()
	err := tx.Model("balance_flow_requirement").
		Where("site_id", siteID).
		Where("user_id", userID).
		Where("status", 1).
		Where("remaining_flow > 0").
		Where("(expire_time IS NULL OR expire_time > ?)", now).
		Order("created_at ASC, id ASC").
		Scan(&rules)
	if err != nil {
		return err
	}

	remaining := amount
	for _, rule := range rules {
		if remaining <= 0 {
			break
		}
		if rule.RemainingFlow <= 0 {
			continue
		}

		consume := math.Min(rule.RemainingFlow, remaining)
		newCompleted := rule.CompletedFlow + consume
		newRemaining := rule.RemainingFlow - consume

		progress := 0.0
		if rule.RequiredFlow > 0 {
			progress = (newCompleted / rule.RequiredFlow) * 100
			if progress > 100 {
				progress = 100
			}
		}

		status := rule.Status
		completeTime := rule.CompleteTime
		if newCompleted >= rule.RequiredFlow {
			status = 2
			if completeTime == nil {
				completeTime = now
			}
		}

		updateData := g.Map{
			"completed_flow": newCompleted,
			"remaining_flow": newRemaining,
			"progress":       progress,
			"status":         status,
			"updated_at":     now,
		}
		if status == 2 {
			updateData["complete_time"] = completeTime
		}

		if _, err := tx.Model("balance_flow_requirement").
			Where("id", rule.Id).
			Where("status", 1).
			Update(updateData); err != nil {
			return err
		}

		g.Log().Infof(ctx, "流水消费更新: req_id=%d, user_id=%d, consume=%.4f, completed=%.4f, remaining=%.4f, status=%d",
			rule.Id, userID, consume, newCompleted, newRemaining, status)

		remaining -= consume
	}

	return nil
}

func insertFlowConsumeRetry(ctx context.Context, groupName string, msg BetSettledMessage, err error) error {
	_, dbErr := g.DB(groupName).Model("balance_flow_consume_retry").Data(g.Map{
		"site_id":      msg.SiteID,
		"user_id":      msg.UserID,
		"order_no":     msg.OrderNo,
		"valid_amount": msg.ValidAmount,
		"error_msg":    err.Error(),
		"created_at":   gtime.Now(),
	}).InsertIgnore()
	return dbErr
}

// handleInitBalanceUser 处理初始化余额消息
func handleInitBalanceUser(ctx context.Context, msg amqp.Delivery) error {
	startTime := time.Now()

	// 解析消息
	var message InitBalanceUserMessage
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		g.Log().Errorf(ctx, "解析消息失败: %v, body: %s", err, string(msg.Body))
		return err
	}

	g.Log().Infof(ctx, "收到初始化余额消息: site_id=%d, user_id=%d, username=%s, currency=%s, message_id=%s",
		message.SiteID, message.UserID, message.Username, message.Currency, msg.MessageId)

	// 创建带站点信息的 context
	ctx = context.WithValue(ctx, middleware.SiteIdContextKey, uint(message.SiteID))

	// 调用 balance 服务初始化余额
	req := &v1.InitBalanceUserReq{
		SiteId:   message.SiteID,
		UserId:   message.UserID,
		Username: message.Username,
		Currency: message.Currency,
	}

	res, err := frontend.Balance().InitBalanceUser(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "初始化余额失败: %v, site_id=%d, user_id=%d", err, message.SiteID, message.UserID)
		return err
	}

	if !res.Success {
		g.Log().Errorf(ctx, "初始化余额返回失败: %s, site_id=%d, user_id=%d", res.Message, message.SiteID, message.UserID)
		return fmt.Errorf("初始化余额失败: %s", res.Message)
	}

	duration := time.Since(startTime)
	g.Log().Infof(ctx, "初始化余额成功: site_id=%d, user_id=%d, username=%s, 耗时: %v",
		message.SiteID, message.UserID, message.Username, duration)

	return nil
}
