package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	amqp "github.com/rabbitmq/amqp091-go"
	v1 "jh_balance_service/api/frontend/balance/v1"
	"jh_balance_service/internal/middleware"
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
