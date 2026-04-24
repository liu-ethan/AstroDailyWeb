package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"astrodailyweb/backend/internal/auth"
	"astrodailyweb/backend/internal/config"
	"astrodailyweb/backend/internal/controller"
	"astrodailyweb/backend/internal/db"
	"astrodailyweb/backend/internal/llm"
	"astrodailyweb/backend/internal/middleware"
	"astrodailyweb/backend/internal/notify"
	"astrodailyweb/backend/internal/repository"
	"astrodailyweb/backend/internal/router"
	"astrodailyweb/backend/internal/scheduler"
	"astrodailyweb/backend/internal/service"
)

type App struct {
	Config    config.Config
	Logger    *slog.Logger
	Engine    *gin.Engine
	DB        *sql.DB
	Redis     *redis.Client
	Scheduler *scheduler.Scheduler
}

// Build 组装应用所需的通用组件并返回容器对象。
// 参数：cfg - 全局配置；log - 日志实例。
// 返回：*App - 应用容器；error - 组装失败错误。
func Build(cfg config.Config, log *slog.Logger) (*App, error) {
	database, err := db.NewMySQL(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("init db failed: %w", err)
	}
	redisClient, err := db.NewRedis(cfg.Redis)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("init redis failed: %w", err)
	}

	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Issuer, time.Duration(cfg.JWT.ExpireMinutes)*time.Minute)
	tokenStore := auth.NewRedisTokenStore(redisClient, cfg.Redis.KeyPrefix)
	smtp := notify.NewClient(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.From)

	var llmClient llm.Client = &llm.StubClient{}
	provider := strings.TrimSpace(cfg.LLM.Provider)
	if provider == "openai-compatible" || (provider == "" && cfg.LLM.APIKey != "" && cfg.LLM.BaseURL != "") {
		llmClient = llm.NewOpenAICompatibleClient(cfg.LLM.APIKey, cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout)
	}

	authMapper := repository.NewAuthMapper(database)
	fortuneMapper := repository.NewFortuneMapper(database)
	userMapper := repository.NewUserMapper(database)

	authSvc := service.NewAuthService(authMapper, smtp, jwtMgr, tokenStore)
	fortuneSvc := service.NewFortuneService(fortuneMapper, userMapper, llmClient, smtp)
	userSvc := service.NewUserService(userMapper)

	ctrls := router.Controllers{
		Health:  controller.NewHealthController(),
		Auth:    controller.NewAuthController(authSvc),
		Fortune: controller.NewFortuneController(fortuneSvc),
		User:    controller.NewUserController(userSvc),
	}

	mws := []gin.HandlerFunc{
		middleware.RequestID(),
		middleware.Recovery(log),
		middleware.ErrorHandler(),
	}

	engine := router.NewEngine(mws, ctrls, middleware.JWTAuth(jwtMgr, tokenStore))

	s := scheduler.New(log, userSvc, fortuneSvc, authSvc)
	if err = s.RegisterJobs(); err != nil {
		return nil, fmt.Errorf("register cron jobs failed: %w", err)
	}

	return &App{Config: cfg, Logger: log, Engine: engine, DB: database, Redis: redisClient, Scheduler: s}, nil
}
