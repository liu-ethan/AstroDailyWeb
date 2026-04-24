package main

import (
	"context"
	"log"
	"os"

	"astrodailyweb/backend/internal/config"
	"astrodailyweb/backend/internal/db"
	"astrodailyweb/backend/internal/llm"
	"astrodailyweb/backend/internal/notify"
	"astrodailyweb/backend/internal/repository"
	"astrodailyweb/backend/internal/service"
)

type fakeLLM struct {
	content string
}

func (f *fakeLLM) GenerateTodayFortune(ctx context.Context, profile llm.FortuneProfile) (string, error) {
	_ = ctx
	_ = profile
	return f.content, nil
}

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.NewMySQL(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = database.Close() }()

	userMapper := repository.NewUserMapper(database)
	fortuneMapper := repository.NewFortuneMapper(database)

	smtpClient := notify.NewClient(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.From)
	fortuneSvc := service.NewFortuneService(
		fortuneMapper,
		userMapper,
		&fakeLLM{content: "测试运势：一切顺利。"},
		smtpClient,
	)
	userSvc := service.NewUserService(userMapper)

	ctx := context.Background()
	users, err := userSvc.ListSubscribedUsers(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(users) == 0 {
		log.Print("no subscribed users found")
		return
	}

	if err := fortuneSvc.GenerateForSubscribedUsers(ctx, users); err != nil {
		log.Fatal(err)
	}
	log.Printf("sent to %d users", len(users))
}
