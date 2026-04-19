package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/response"
	"astrodailyweb/backend/internal/service"
)

type AuthController struct {
	svc service.AuthService
}

// NewAuthController 创建认证控制器。
// 参数：svc - 认证服务。
// 返回：*AuthController - 认证控制器实例。
func NewAuthController(svc service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

type SendCodeReq struct {
	Email        string `json:"email" binding:"required,email"`
	BusinessType int    `json:"business_type" binding:"required,oneof=1 2"`
}

type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordReq struct {
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

// SendCode 处理验证码发送请求。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *AuthController) SendCode(c *gin.Context) {
	var req SendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(4000, "参数错误"))
		return
	}
	if err := ctl.svc.SendCode(c.Request.Context(), req.Email, req.BusinessType); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.SuccessMessage("验证码发送成功", nil))
}

// Register 处理注册请求。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *AuthController) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(4000, "参数错误"))
		return
	}
	if err := ctl.svc.Register(c.Request.Context(), req.Email, req.Password, req.Code); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.SuccessMessage("注册成功", nil))
}

// Login 处理登录请求并返回令牌。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *AuthController) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(4000, "参数错误"))
		return
	}
	token, expiresIn, err := ctl.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.SuccessMessage("登录成功", gin.H{"token": token, "expires_in": expiresIn}))
}

// ResetPassword 处理重置密码请求。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *AuthController) ResetPassword(c *gin.Context) {
	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail(4000, "参数错误"))
		return
	}
	if err := ctl.svc.ResetPassword(c.Request.Context(), req.Email, req.NewPassword, req.Code); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response.SuccessMessage("密码重置成功", nil))
}
