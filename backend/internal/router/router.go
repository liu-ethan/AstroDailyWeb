package router

import (
	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/controller"
)

type Controllers struct {
	Health  *controller.HealthController
	Auth    *controller.AuthController
	Fortune *controller.FortuneController
	User    *controller.UserController
}

// NewEngine 构建 HTTP 路由并绑定中间件和控制器。
// 参数：mws - 全局中间件列表；ctrls - 控制器集合；authMW - 鉴权中间件。
// 返回：*gin.Engine - 初始化完成的 Gin 引擎。
func NewEngine(mws []gin.HandlerFunc, ctrls Controllers, authMW gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	for _, mw := range mws {
		r.Use(mw)
	}

	r.GET("/healthz", ctrls.Health.Healthz)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/send-code", ctrls.Auth.SendCode)
		auth.POST("/register", ctrls.Auth.Register)
		auth.POST("/login", ctrls.Auth.Login)
		auth.POST("/reset-password", ctrls.Auth.ResetPassword)

		secured := v1.Group("")
		secured.Use(authMW)
		secured.GET("/fortune/today", ctrls.Fortune.Today)
		secured.POST("/user/subscribe", ctrls.User.Subscribe)
		secured.POST("/user/unsubscribe", ctrls.User.Unsubscribe)
	}
	return r
}
