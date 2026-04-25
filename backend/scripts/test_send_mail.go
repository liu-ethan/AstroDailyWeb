package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"astrodailyweb/backend/internal/config"
	"astrodailyweb/backend/internal/db"
	"astrodailyweb/backend/internal/llm"
	"astrodailyweb/backend/internal/notify"
	"astrodailyweb/backend/internal/repository"
	"astrodailyweb/backend/internal/service"
)

// TODO: change to your target email
var targetEmail = "2361352642@qq.com"

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
	authMapper := repository.NewAuthMapper(database)

	provider := strings.TrimSpace(cfg.LLM.Provider)
	if provider != "openai-compatible" && !(provider == "" && cfg.LLM.APIKey != "" && cfg.LLM.BaseURL != "") {
		log.Fatal("LLM config missing: please set LLM provider/api_key/base_url")
	}
	llmClient := llm.NewOpenAICompatibleClient(cfg.LLM.APIKey, cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout)
	smtpClient := notify.NewClient(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.From)
	fortuneSvc := service.NewFortuneService(fortuneMapper, userMapper, llmClient, smtpClient)

	ctx := context.Background()
	user, err := authMapper.FindUserByEmail(ctx, targetEmail)
	if err != nil {
		log.Fatal(err)
	}

	date, content, err := fortuneSvc.GetToday(ctx, user.ID)
	if err != nil {
		log.Fatal(err)
	}

	if err := smtpClient.Send(ctx, []string{targetEmail}, "每日运势", fmt.Sprintf("%s\n\n%s", date, content)); err != nil {
		log.Fatal(err)
	}
	log.Printf("sent to %s", targetEmail)
}
