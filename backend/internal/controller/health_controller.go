package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/response"
)

type HealthController struct{}

// NewHealthController 创建健康检查控制器。
// 参数：无。
// 返回：*HealthController - 健康检查控制器实例。
func NewHealthController() *HealthController {
	return &HealthController{}
}

// Healthz 返回服务存活状态。
// 参数：c - Gin 请求上下文。
// 返回：无。
func (ctl *HealthController) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, response.SuccessMessage("ok", gin.H{"status": "up"}))
}
