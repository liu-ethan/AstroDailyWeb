package middleware

import (
	"errors"

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
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			c.JSON(response.HTTPStatusFromBizCode(appErr.Code), response.Fail(appErr.Code, appErr.Message))
			return
		}
		c.JSON(response.HTTPStatusFromBizCode(5000), response.Fail(5000, "系统繁忙，请稍后重试"))
	}
}
