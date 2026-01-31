package payment

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

// PaymentFactory 支付平台工厂
type PaymentFactory struct {
	platforms map[string]func() PaymentPlatform
	mu        sync.RWMutex
}

var (
	factory *PaymentFactory
	once    sync.Once
)

// GetFactory 获取工厂单例
func GetFactory() *PaymentFactory {
	once.Do(func() {
		factory = &PaymentFactory{
			platforms: make(map[string]func() PaymentPlatform),
		}
	})
	return factory
}

// Register 注册支付平台
// platformCode: 平台代码，如 "MG", "ALIPAY", "WECHAT"
// creator: 创建平台实例的函数
func (f *PaymentFactory) Register(platformCode string, creator func() PaymentPlatform) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.platforms[platformCode] = creator
}

// Create 创建支付平台实例
func (f *PaymentFactory) Create(platformCode string) (PaymentPlatform, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	creator, exists := f.platforms[platformCode]
	ctx := context.Background()

	// 打印所有已注册的平台
	for k := range f.platforms {
		g.Log().Infof(ctx, "已注册支付平台: %s", k)
	}

	if !exists {
		return nil, fmt.Errorf("不支持的支付平台: %s", platformCode)
	}

	return creator(), nil
}

// GetSupportedPlatforms 获取所有支持的平台列表
func (f *PaymentFactory) GetSupportedPlatforms() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	platforms := make([]string, 0, len(f.platforms))
	for code := range f.platforms {
		platforms = append(platforms, code)
	}
	return platforms
}
