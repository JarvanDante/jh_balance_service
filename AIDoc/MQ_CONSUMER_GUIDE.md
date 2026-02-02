# Balance Service MQ Consumer 使用指南

## 概述

Balance 服务的 MQ 消费者监听用户注册事件，自动初始化用户余额。

## 架构流程

```
App Service (用户注册)
  ↓
发送 MQ 消息到 exchange: user.created
  ↓
Balance Service MQ Consumer 监听
  ↓
调用 InitBalanceUser 初始化余额
  ↓
在 balance_user 表中创建记录
```

## 配置

### 1. 配置 RabbitMQ 连接

在 `manifest/config/config.yaml` 中添加：

```yaml
rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
```

### 2. 确保 RabbitMQ 运行

```bash
# 使用 Docker 启动 RabbitMQ
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

访问管理界面：http://localhost:15672 (guest/guest)

## 启动消费者

### 开发环境

```bash
cd jh_balance_service

# 前台运行（可以看到日志）
./jh_balance_service mq-consumer

# 后台运行
nohup ./jh_balance_service mq-consumer > logs/mq_consumer.log 2>&1 &
```

### 生产环境

#### 使用 Supervisor

创建 `/etc/supervisor/conf.d/jh_balance_mq_consumer.conf`:

```ini
[program:jh_balance_mq_consumer]
command=/path/to/jh_balance_service mq-consumer
directory=/path/to/jh_balance_service
user=www-data
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/jh_balance_mq_consumer.log
```

启动：

```bash
supervisorctl reread
supervisorctl update
supervisorctl start jh_balance_mq_consumer
supervisorctl status jh_balance_mq_consumer
```

#### 使用 Systemd

创建 `/etc/systemd/system/jh_balance_mq_consumer.service`:

```ini
[Unit]
Description=JH Balance MQ Consumer
After=network.target rabbitmq.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/jh_balance_service
ExecStart=/path/to/jh_balance_service mq-consumer
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动：

```bash
systemctl daemon-reload
systemctl enable jh_balance_mq_consumer
systemctl start jh_balance_mq_consumer
systemctl status jh_balance_mq_consumer
```

## 监控

### 查看日志

```bash
# 实时查看日志
tail -f logs/mq_consumer.log

# 或使用 journalctl（如果用 systemd）
journalctl -u jh_balance_mq_consumer -f
```

### 查看 RabbitMQ 队列状态

访问 RabbitMQ 管理界面：http://localhost:15672

查看队列 `balance.init_user`:

- Messages: 待处理消息数
- Message rate: 消息处理速率
- Consumers: 消费者数量

### 命令行查看队列

```bash
# 查看队列信息
docker exec rabbitmq rabbitmqctl list_queues name messages consumers

# 查看 exchange 绑定
docker exec rabbitmq rabbitmqctl list_bindings
```

## 测试

### 1. 测试用户注册

```bash
# 调用 app 服务的注册接口
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

### 2. 查看消息是否被消费

```bash
# 查看 balance 服务日志
tail -f jh_balance_service/logs/mq_consumer.log

# 应该看到类似日志：
# 收到初始化余额消息: site_id=1, user_id=123, username=testuser
# 初始化余额成功: site_id=1, user_id=123, username=testuser, 耗时: 50ms
```

### 3. 查看数据库

```sql
-- 查看用户余额是否创建
SELECT * FROM balance_user WHERE user_id = 123;

-- 查看余额日志
SELECT * FROM balance_log WHERE user_id = 123 ORDER BY id DESC LIMIT 10;
```

## 故障排查

### 问题1: 消费者无法连接 RabbitMQ

**错误日志：**

```
Failed to initialize RabbitMQ: dial tcp: connection refused
```

**解决方案：**

1. 检查 RabbitMQ 是否运行：`docker ps | grep rabbitmq`
2. 检查配置文件中的 URL 是否正确
3. 检查网络连接

### 问题2: 消息未被消费

**检查步骤：**

1. 查看队列中是否有消息：

```bash
docker exec rabbitmq rabbitmqctl list_queues
```

2. 查看消费者是否连接：

```bash
docker exec rabbitmq rabbitmqctl list_consumers
```

3. 查看 exchange 绑定：

```bash
docker exec rabbitmq rabbitmqctl list_bindings
```

### 问题3: 消息处理失败

**错误日志：**

```
处理消息失败: xxx, 消息ID: xxx
```

**解决方案：**

1. 查看详细错误日志
2. 检查数据库连接
3. 检查站点配置是否正确
4. 失败的消息会自动重新入队重试

### 问题4: 重复消费

**原因：**

- 消息处理成功但未 Ack
- 消费者异常退出

**解决方案：**

- 确保消息处理逻辑是幂等的
- 检查是否有重复的消费者实例

## 性能优化

### 1. 调整 QoS

在 `rabbitmq.go` 中修改：

```go
err := channel.Qos(
    20,    // prefetch count (增加并发处理数)
    0,     // prefetch size
    false, // global
)
```

### 2. 启动多个消费者

```bash
# 启动多个消费者实例
./jh_balance_service mq-consumer &
./jh_balance_service mq-consumer &
./jh_balance_service mq-consumer &
```

### 3. 监控指标

关注以下指标：

- 消息处理速率（messages/second）
- 平均处理时间
- 队列积压数量
- 失败重试次数

## 注意事项

1. ✅ 消费者会自动重连 RabbitMQ
2. ✅ 消息处理失败会自动重试
3. ✅ 使用手动 Ack 确保消息不丢失
4. ⚠️ 确保消息处理逻辑是幂等的
5. ⚠️ 监控队列积压，及时扩容
