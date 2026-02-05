package balance

// SourceType 来源类型
const (
	SourceTypeRecharge         = 1  // 充值
	SourceTypeBonus            = 2  // 红利
	SourceTypeRebate           = 3  // 返水
	SourceTypeManual           = 4  // 人工加款
	SourceTypeRegisterActivity = 5  // 注册送活动
	SourceTypeRechargeActivity = 6  // 充值送活动
	SourceTypeUpgradeActivity  = 7  // 升级送活动
	SourceTypeDailySign        = 8  // 每日签到活动
	SourceTypeLuckySpin        = 9  // 大转盘活动
	SourceTypeInvite           = 10 // 邀请朋友活动
)

// RuleType 规则类型
const (
	RuleTypeRecharge         = 1  // 充值流水
	RuleTypeBonus            = 2  // 红利流水
	RuleTypeRebate           = 3  // 返水流水
	RuleTypeManual           = 4  // 人工加款流水
	RuleTypeRegisterActivity = 5  // 注册送活动流水
	RuleTypeRechargeActivity = 6  // 充值送活动流水
	RuleTypeUpgradeActivity  = 7  // 升级送活动流水
	RuleTypeDailySign        = 8  // 每日签到活动流水
	RuleTypeLuckySpin        = 9  // 大转盘活动流水
	RuleTypeInvite           = 10 // 邀请朋友活动
)
