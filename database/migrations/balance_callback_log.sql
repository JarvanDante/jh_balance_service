-- 支付回调失败日志表
CREATE TABLE `balance_callback_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `site_id` int NOT NULL DEFAULT '0' COMMENT '站点ID',
  `callback_type` tinyint NOT NULL DEFAULT '1' COMMENT '回调类型 1=充值 2=提现',
  `order_sn` varchar(64) NOT NULL DEFAULT '' COMMENT '订单号',
  `trade_status` varchar(32) DEFAULT NULL COMMENT '交易状态',
  `actual_amount` decimal(15,4) DEFAULT NULL COMMENT '实际金额',
  `failure_reason` varchar(255) NOT NULL DEFAULT '' COMMENT '失败原因',
  `failure_stage` varchar(64) NOT NULL DEFAULT '' COMMENT '失败阶段',
  `request_data` text COMMENT '请求数据(JSON)',
  `error_message` text COMMENT '错误详情',
  `retry_count` int NOT NULL DEFAULT '0' COMMENT '重试次数',
  `is_resolved` tinyint NOT NULL DEFAULT '0' COMMENT '是否已解决 0=未解决 1=已解决',
  `resolved_at` datetime DEFAULT NULL COMMENT '解决时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_sn` (`order_sn`),
  KEY `idx_site_callback` (`site_id`, `callback_type`, `created_at`),
  KEY `idx_is_resolved` (`is_resolved`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付回调失败日志表';
