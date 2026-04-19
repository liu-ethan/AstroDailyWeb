package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/auth"
	"astrodailyweb/backend/internal/response"
)

const UserIDKey = "user_id"

// JWTAuth 校验 Bearer Token、检查 Redis 中是否存在，并写入用户ID到上下文。
// 参数：jwtMgr - JWT 管理器；tokenStore - Token 存储。
// 返回：gin.HandlerFunc - Gin 中间件函数。
func JWTAuth(jwtMgr *auth.JWTManager, tokenStore auth.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(response.HTTPStatusFromBizCode(4010), response.Fail(4010, "未授权"))
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtMgr.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(response.HTTPStatusFromBizCode(4011), response.Fail(4011, "Token无效或已过期"))
			return
		}
		exists, err := tokenStore.Exists(c.Request.Context(), token)
		if err != nil || !exists {
			c.AbortWithStatusJSON(response.HTTPStatusFromBizCode(4011), response.Fail(4011, "Token无效或已过期"))
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}
