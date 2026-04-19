package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/middleware"
	"astrodailyweb/backend/internal/response"
	"astrodailyweb/backend/internal/service"
)

type FortuneController struct {
	svc service.FortuneService
}

// NewFortuneController 创建运势控制器。
// 参数：svc - 运势服务。
// 返回：*FortuneController - 运势控制器实例。
func NewFortuneController(svc service.FortuneService) *FortuneController {
	return &FortuneController{svc: svc}
}

// Today 获取当前用户今日运势。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *FortuneController) Today(c *gin.Context) {
	uid, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Fail(4010, "未授权"))
		return
	}

	date, content, err := ctl.svc.GetToday(c.Request.Context(), uid.(int64), "")
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"date": date, "content": content}))
}
