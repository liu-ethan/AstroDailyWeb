package middleware

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/apperror"
	"astrodailyweb/backend/internal/response"
)

// ErrorHandler 将上下文中的错误统一转换为标准响应。
// 参数：无。
// 返回：gin.HandlerFunc - Gin 中间件函数。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		requestID := c.GetString(RequestIDKey)
		path := c.Request.URL.Path
		method := c.Request.Method
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			status := response.HTTPStatusFromBizCode(appErr.Code)
			slog.Error(
				"request failed",
				"request_id", requestID,
				"method", method,
				"path", path,
				"status", status,
				"code", appErr.Code,
				"msg", appErr.Message,
				"err", appErr.Err,
			)
			c.JSON(status, response.Fail(appErr.Code, appErr.Message))
			return
		}
		status := response.HTTPStatusFromBizCode(5000)
		slog.Error(
			"request failed",
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", status,
			"code", 5000,
			"msg", err.Error(),
			"err", err,
		)
		c.JSON(status, response.Fail(5000, "系统繁忙，请稍后重试"))
	}
}
