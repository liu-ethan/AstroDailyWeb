package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/middleware"
	"astrodailyweb/backend/internal/repository"
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

type SaveProfileReq struct {
	Birthday      string `json:"birthday" binding:"required"`
	Constellation string `json:"constellation" binding:"required"`
	Gender        string `json:"gender" binding:"required"`
	City          string `json:"city" binding:"required"`
	Occupation    string `json:"occupation" binding:"required"`
}

// GetProfile 获取当前用户资料。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *UserController) GetProfile(c *gin.Context) {
	uid, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Fail(4010, "未授权"))
		return
	}
	profile, err := ctl.svc.GetProfile(c.Request.Context(), uid.(int64))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.Success(profile))
}

// SaveProfile 保存当前用户资料。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *UserController) SaveProfile(c *gin.Context) {
	uid, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Fail(4010, "未授权"))
		return
	}
	var req SaveProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(4000, "参数错误"))
		return
	}
	profile := repository.UserProfile{
		UserID:        uid.(int64),
		Birthday:      req.Birthday,
		Constellation: req.Constellation,
		Gender:        req.Gender,
		City:          req.City,
		Occupation:    req.Occupation,
	}
	if err := ctl.svc.SaveProfile(c.Request.Context(), profile); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.SuccessMessage("用户资料已保存", nil))
}
