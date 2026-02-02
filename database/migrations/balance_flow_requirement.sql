-- 流水要求记录表（打码量要求表）
CREATE TABLE IF NOT EXISTS `balance_flow_requirement` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `site_id` int(11) NOT NULL DEFAULT '1' COMMENT '站点ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `username` varchar(50) NOT NULL DEFAULT '' COMMENT '用户名',
  
  -- 来源信息
  `source_type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '来源类型: 1=充值, 2=红利, 3=返水, 4=活动, 5=人工加款',
  `source_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '来源ID（充值订单ID、红利ID等）',
  `trade_no` varchar(64) NOT NULL DEFAULT '' COMMENT '关联交易号',
  
  -- 金额信息
  `amount` decimal(15,4) NOT NULL DEFAULT '0.0000' COMMENT '触发金额（充值金额、红利金额等）',
  `bonus_amount` decimal(15,4) NOT NULL DEFAULT '0.0000' COMMENT '赠送金额（如果有）',
  
  -- 流水要求
  `flow_multiple` decimal(10,2) NOT NULL DEFAULT '1.00' COMMENT '流水倍数（如：1倍、3倍、5倍）',
  `required_flow` decimal(15,4) NOT NULL DEFAULT '0.0000' COMMENT '需要完成的流水金额 = (amount + bonus_amount) * flow_multiple',
  `completed_flow` decimal(15,4) NOT NULL DEFAULT '0.0000' COMMENT '已完成的流水金额',
  `remaining_flow` decimal(15,4) NOT NULL DEFAULT '0.0000' COMMENT '剩余流水金额 = required_flow - completed_flow',
  
  -- 状态信息
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 1=进行中, 2=已完成, 3=已过期, 4=已取消',
  `progress` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '完成进度百分比（0-100）',
  
  -- 时间限制
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间（NULL表示永久有效）',
  `complete_time` datetime DEFAULT NULL COMMENT '完成时间',
  
  -- 规则配置（从规则表复制过来，避免规则修改影响历史记录）
  `game_flow_rates` text COMMENT '游戏流水贡献率配置（JSON格式）',
  `valid_game_types` varchar(255) DEFAULT NULL COMMENT '有效游戏类型（JSON数组，NULL表示所有游戏）',
  
  -- 备注信息
  `remark` varchar(500) DEFAULT NULL COMMENT '备注说明',
  `admin_id` int(11) NOT NULL DEFAULT '0' COMMENT '操作管理员ID（人工操作时）',
  
  -- 时间戳
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_site_user_status` (`site_id`, `user_id`, `status`),
  KEY `idx_username` (`username`),
  KEY `idx_trade_no` (`trade_no`),
  KEY `idx_status` (`status`),
  KEY `idx_source` (`source_type`, `source_id`),
  KEY `idx_expire_time` (`expire_time`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水要求记录表';

-- 流水规则配置表
CREATE TABLE IF NOT EXISTS `balance_flow_rule` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `site_id` int(11) NOT NULL DEFAULT '1' COMMENT '站点ID',
  
  -- 规则基本信息
  `rule_name` varchar(100) NOT NULL DEFAULT '' COMMENT '规则名称',
  `rule_type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '规则类型: 1=充值流水, 2=红利流水, 3=活动流水, 4=人工加款流水',
  `flow_multiple` decimal(10,2) NOT NULL DEFAULT '1.00' COMMENT '默认流水倍数',
  
  -- 游戏类型配置（不同游戏类型的流水贡献率）
  `game_flow_rates` text COMMENT '游戏流水贡献率配置（JSON格式）',
  -- 示例: {"slot": 100, "live": 50, "sport": 100, "fishing": 80, "lottery": 100}
  -- 100表示100%计入流水，50表示50%计入流水
  
  -- 有效游戏类型（NULL表示所有游戏都有效）
  `valid_game_types` varchar(255) DEFAULT NULL COMMENT '有效游戏类型（JSON数组）',
  -- 示例: ["slot", "sport", "fishing"] 或 NULL（所有游戏）
  
  -- 时间限制
  `expire_days` int(11) NOT NULL DEFAULT '0' COMMENT '有效天数（0表示永久）',
  
  -- 金额限制
  `min_amount` decimal(15,4) NOT NULL DEFAULT '0.0000' COMMENT '最小触发金额',
  `max_amount` decimal(15,4) NOT NULL DEFAULT '0.0000' COMMENT '最大触发金额（0表示不限制）',
  
  -- 状态
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  
  -- 备注
  `remark` varchar(500) DEFAULT NULL COMMENT '备注说明',
  
  -- 时间戳
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_site_type` (`site_id`, `rule_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水规则配置表';

-- 插入默认规则示例
INSERT INTO `balance_flow_rule` (`site_id`, `rule_name`, `rule_type`, `flow_multiple`, `game_flow_rates`, `valid_game_types`, `expire_days`, `min_amount`, `status`, `remark`) VALUES
(1, '充值1倍流水', 1, 1.00, '{"slot": 100, "live": 50, "sport": 100, "fishing": 80, "lottery": 100, "chess": 60}', NULL, 30, 100.00, 1, '充值金额需完成1倍有效投注'),
(1, '红利3倍流水', 2, 3.00, '{"slot": 100, "live": 50, "sport": 100, "fishing": 80, "lottery": 100, "chess": 60}', NULL, 7, 10.00, 1, '红利金额需完成3倍有效投注'),
(1, '活动5倍流水', 3, 5.00, '{"slot": 100, "live": 50, "sport": 100, "fishing": 80, "lottery": 100, "chess": 60}', '["slot", "sport", "fishing"]', 7, 50.00, 1, '活动奖励需完成5倍有效投注，仅限老虎机、体育、捕鱼');
