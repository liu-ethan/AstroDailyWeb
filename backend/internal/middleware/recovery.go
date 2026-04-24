package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/response"
)

// Recovery 捕获 panic 并输出统一错误响应。
// 参数：log - 日志实例。
// 返回：gin.HandlerFunc - Gin 中间件函数。
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error("panic recovered", "panic", recovered)
		c.AbortWithStatusJSON(response.HTTPStatusFromBizCode(5000), response.Fail(5000, "系统异常"))
	})
}
