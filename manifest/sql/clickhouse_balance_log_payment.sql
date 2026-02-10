CREATE DATABASE IF NOT EXISTS game_analytics;

CREATE TABLE game_analytics.balance_log_payment
(
    event_id UInt64,
    site_id UInt32,
    id UInt64,
    user_id UInt64,
    username String,
    activity_recharge_id UInt32,
    gateway UInt8,
    payment_id UInt32,
    payment_account_id UInt32,
    bank_value String,
    trade_no String,
    money Decimal(15, 4),
    actual_money Decimal(15, 4),
    fee Decimal(15, 4),
    status UInt8,
    admin_id UInt32,
    remark String,
    created_at DateTime,
    updated_at DateTime
)
    ENGINE = ReplacingMergeTree(event_id)
PARTITION BY toYYYYMM(created_at)
ORDER BY (site_id, id)
SETTINGS index_granularity = 8192;

