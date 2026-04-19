package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/middleware"
	"astrodailyweb/backend/internal/response"
	"astrodailyweb/backend/internal/service"
)

type UserController struct {
	svc service.UserService
}

// NewUserController 创建用户设置控制器。
// 参数：svc - 用户服务。
// 返回：*UserController - 用户控制器实例。
func NewUserController(svc service.UserService) *UserController {
	return &UserController{svc: svc}
}

// Subscribe 订阅每日运势邮件。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *UserController) Subscribe(c *gin.Context) {
	uid, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Fail(4010, "未授权"))
		return
	}
	if err := ctl.svc.Subscribe(c.Request.Context(), uid.(int64)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.SuccessMessage("已订阅，以后每日8点将发送当天运势给你", nil))
}

// Unsubscribe 取消每日运势邮件订阅。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *UserController) Unsubscribe(c *gin.Context) {
	uid, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Fail(4010, "未授权"))
		return
	}
	if err := ctl.svc.Unsubscribe(c.Request.Context(), uid.(int64)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.SuccessMessage("已取消订阅，不再发送当日运势邮件给你", nil))
}
