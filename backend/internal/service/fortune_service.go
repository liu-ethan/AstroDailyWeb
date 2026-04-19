package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"astrodailyweb/backend/internal/apperror"
	"astrodailyweb/backend/internal/llm"
	"astrodailyweb/backend/internal/repository"
)

type FortuneService interface {
	GetToday(ctx context.Context, userID int64) (date string, content string, err error)
	GenerateForSubscribedUsers(ctx context.Context, users []repository.UserRecord) error
	CleanupHistory(ctx context.Context, keepDays int) error
}

type fortuneService struct {
	mapper     repository.FortuneMapper
	userMapper repository.UserMapper
	llm        llm.Client
}

// NewFortuneService 创建运势服务。
// 参数：mapper - 运势数据访问层；userMapper - 用户数据访问层；llmClient - 大模型客户端。
// 返回：FortuneService - 运势服务接口实现。
func NewFortuneService(mapper repository.FortuneMapper, userMapper repository.UserMapper, llmClient llm.Client) FortuneService {
	return &fortuneService{mapper: mapper, userMapper: userMapper, llm: llmClient}
}

// GetToday 获取或生成用户今日运势。
// 参数：ctx - 上下文；userID - 用户ID。
// 返回：string - 日期；string - 运势内容；error - 处理失败错误。
func (s *fortuneService) GetToday(ctx context.Context, userID int64) (string, string, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	content, err := s.mapper.GetByUserAndDate(ctx, userID, now)
	if err == nil {
		return today, content, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}

	profile, err := s.userMapper.GetProfile(ctx, userID)
	if err != nil {
		return "", "", apperror.New(4003, "请先完善运势资料")
	}
	missing := validateProfile(profile)
	if len(missing) > 0 {
		return "", "", apperror.New(4003, "请先完善运势资料: "+strings.Join(missing, ","))
	}

	content, err = s.llm.GenerateTodayFortune(ctx, llm.FortuneProfile{
		Birthday:      profile.Birthday,
		Today:         today,
		Constellation: profile.Constellation,
		Gender:        profile.Gender,
		City:          profile.City,
		Occupation:    profile.Occupation,
	})
	if err != nil {
		return "", "", err
	}
	if err = s.mapper.Save(ctx, userID, now, content); err != nil {
		return "", "", err
	}
	return today, content, nil
}

func validateProfile(p repository.UserProfile) []string {
	missing := make([]string, 0, 6)
	if strings.TrimSpace(p.Birthday) == "" {
		missing = append(missing, "birthday")
	}
	if strings.TrimSpace(p.Constellation) == "" {
		missing = append(missing, "constellation")
	}
	if strings.TrimSpace(p.Gender) == "" {
		missing = append(missing, "gender")
	}
	if strings.TrimSpace(p.City) == "" {
		missing = append(missing, "city")
	}
	if strings.TrimSpace(p.Occupation) == "" {
		missing = append(missing, "occupation")
	}
	return missing
}

// GenerateForSubscribedUsers 为订阅用户批量生成运势（当前为占位实现）。
// 参数：ctx - 上下文；users - 订阅用户列表。
// 返回：error - 处理失败错误。
func (s *fortuneService) GenerateForSubscribedUsers(ctx context.Context, users []repository.UserRecord) error {
	_ = ctx
	_ = users
	return nil
}

// CleanupHistory 清理历史运势数据。
// 参数：ctx - 上下文；keepDays - 保留天数。
// 返回：error - 清理失败错误。
func (s *fortuneService) CleanupHistory(ctx context.Context, keepDays int) error {
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	return s.mapper.CleanupBefore(ctx, cutoff)
}
