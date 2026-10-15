-- =================================================================================
-- 彩金/红利发放去重唯一索引
--
-- 背景：注册彩金 / 充值送彩金 / 邀请红利 的发放逻辑均为「先 count() 判重、再 grant」，
-- 该「检查-发放」两步在并发下非原子，存在重复发放（同一用户被入账多次）的风险。
--
-- 修复：为三张发放记录表按「站点 + 受益用户」建立唯一索引，配合业务侧的
-- InsertIgnore / 唯一键冲突判重（占位插入），使「同一用户仅能发放一次」由数据库强制保证。
--
-- 注意：
--   1) 若历史数据中已存在重复记录，需先清理重复行再执行 ADD UNIQUE，否则会失败。
--   2) 各站点为独立库，需对每个站点库分别执行本迁移。
--   3) MySQL 不支持 IF NOT EXISTS 建索引，重复执行会报 1061(Duplicate key name)，可忽略。
-- =================================================================================

-- 注册送彩金：同一站点同一用户仅发放一次
ALTER TABLE `balance_register_reward`
  ADD UNIQUE KEY `uk_site_user` (`site_id`, `user_id`);

-- 充值送彩金：同一站点同一用户仅发放一次
ALTER TABLE `balance_deposit_reward`
  ADD UNIQUE KEY `uk_site_user` (`site_id`, `user_id`);

-- 邀请红利：同一站点同一被邀请人仅触发一次
ALTER TABLE `balance_invite_reward`
  ADD UNIQUE KEY `uk_site_invitee` (`site_id`, `invitee_id`);
