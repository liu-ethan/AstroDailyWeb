package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestID 为请求注入请求ID并回写响应头。
// 参数：无。
// 返回：gin.HandlerFunc - Gin 中间件函数。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}
