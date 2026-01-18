package middleware

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const (
	SiteIdContextKey   = "site_id"
	SiteCodeContextKey = "site_code"
	UserIdContextKey   = "user_id"
)

// GetSiteIdFromContext 从上下文中获取站点ID
func GetSiteIdFromContext(ctx context.Context) (uint, bool) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if siteIDs := md.Get("site_id"); len(siteIDs) > 0 {
			siteID, _ := strconv.ParseInt(siteIDs[0], 10, 64)
			return uint(siteID), true
		}
	}

	if siteId, ok := ctx.Value(SiteIdContextKey).(uint); ok {
		return siteId, true
	}
	return 0, false
}

// GetUserIdFromContext 从上下文中获取用户ID
func GetUserIdFromContext(ctx context.Context) (uint, bool) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userIds := md.Get("user_id"); len(userIds) > 0 {
			userId, err := strconv.ParseUint(userIds[0], 10, 32)
			if err == nil {
				return uint(userId), true
			}
		}
	}

	if userId, ok := ctx.Value(UserIdContextKey).(uint); ok {
		return userId, true
	}
	return 0, false
}

// GetSiteCodeFromContext 从上下文中获取站点代码
func GetSiteCodeFromContext(ctx context.Context) (string, bool) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if siteCodes := md.Get("site_code"); len(siteCodes) > 0 {
			return siteCodes[0], true
		}
	}

	if siteCode, ok := ctx.Value(SiteCodeContextKey).(string); ok {
		return siteCode, true
	}
	return "", false
}
