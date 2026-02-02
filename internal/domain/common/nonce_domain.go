package common

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// NonceService Nonce 防重放服务
type NonceService struct{}

// ValidateAndConsume 验证并消费 nonce（防重放）
// bizType: 业务类型，如 "deposit"、"withdraw"
// siteId: 站点ID
// userId: 用户ID
// providedNonce: 用户提供的 nonce
// 返回 error 如果验证失败
func (s *NonceService) ValidateAndConsume(ctx context.Context, bizType string, siteId, userId int64, providedNonce string) error {
	// 1. 构建 Redis Key
	redisKey := fmt.Sprintf("%s:nonce:%d:%d", bizType, siteId, userId)
	g.Log().Infof(ctx, "开始验证 nonce: key=%s, provided_nonce=%s", redisKey, providedNonce)

	// 2. 从 Redis 获取存储的 nonce
	storedNonce, err := g.Redis().Get(ctx, redisKey)
	if err != nil {
		g.Log().Errorf(ctx, "获取 nonce 失败: %v, key=%s", err, redisKey)
		return gerror.New("nonce 已过期，请刷新页面重试")
	}

	// 3. 检查 nonce 是否存在
	if storedNonce.IsNil() {
		g.Log().Errorf(ctx, "nonce 不存在或已过期: key=%s", redisKey)
		return gerror.New("nonce 不存在，请刷新页面重试")
	}

	// 4. 对比 nonce
	storedNonceStr := storedNonce.String()
	g.Log().Infof(ctx, "nonce 对比: stored=[%s], provided=[%s], equal=%v",
		storedNonceStr, providedNonce, storedNonceStr == providedNonce)

	if storedNonceStr != providedNonce {
		g.Log().Errorf(ctx, "nonce 验证失败: stored=%s, provided=%s", storedNonceStr, providedNonce)
		return gerror.New("nonce 验证失败，请刷新页面重试")
	}

	// 5. 删除 nonce（单次使用，防止重复提交）
	delCount, err := g.Redis().Del(ctx, redisKey)
	if err != nil {
		g.Log().Warningf(ctx, "删除 nonce 失败: %v", err)
	}

	// 6. 检查是否被并发删除
	if delCount == 0 {
		// nonce 已被删除（可能是并发请求）
		g.Log().Warningf(ctx, "nonce 已被使用: biz_type=%s, user_id=%d", bizType, userId)
		return gerror.New("请勿重复提交")
	}

	g.Log().Infof(ctx, "nonce 验证成功并已删除: biz_type=%s, user_id=%d", bizType, userId)
	return nil
}

// Generate 生成并存储 nonce
// bizType: 业务类型，如 "deposit"、"withdraw"
// siteId: 站点ID
// userId: 用户ID
// nonce: 生成的 nonce 值
// expiresIn: 过期时间（秒）
func (s *NonceService) Generate(ctx context.Context, bizType string, siteId, userId int64, nonce string, expiresIn int64) error {
	redisKey := fmt.Sprintf("%s:nonce:%d:%d", bizType, siteId, userId)

	g.Log().Infof(ctx, "存储 nonce 到 Redis: key=%s, value=%s, expiresIn=%d", redisKey, nonce, expiresIn)

	err := g.Redis().SetEX(ctx, redisKey, nonce, expiresIn)
	if err != nil {
		g.Log().Errorf(ctx, "存储 nonce 失败: %v", err)
		return gerror.Wrap(err, "生成 nonce 失败")
	}

	return nil
}
