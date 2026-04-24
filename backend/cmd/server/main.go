package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"astrodailyweb/backend/internal/app"
	"astrodailyweb/backend/internal/config"
	"astrodailyweb/backend/internal/logger"
)

// main 加载配置、装配组件并启动 HTTP 服务。
// 参数：无。
// 返回：无。
func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		panic(fmt.Errorf("load config failed: %w", err))
	}

	log := logger.New(cfg.App.LogLevel)
	application, err := app.Build(cfg, log)
	if err != nil {
		panic(fmt.Errorf("build app failed: %w", err))
	}
	defer func() { _ = application.DB.Close() }()
	defer func() {
		if application.Redis != nil {
			_ = application.Redis.Close()
		}
	}()

	if cfg.App.EnableCron {
		application.Scheduler.Start()
		defer application.Scheduler.Stop()
		log.Info("cron scheduler enabled")
	}

	srv := &http.Server{
		Addr:         cfg.App.Host + ":" + cfg.App.Port,
		Handler:      application.Engine,
		ReadTimeout:  cfg.App.ReadTO,
		WriteTimeout: cfg.App.WriteTO,
		IdleTimeout:  cfg.App.IdleTO,
	}

	go func() {
		log.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("server start failed: %w", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown failed", "err", err)
	}
	log.Info("server exited")
}
