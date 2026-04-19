package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"astrodailyweb/backend/internal/llm"
	"astrodailyweb/backend/internal/repository"
)

type FortuneService interface {
	GetToday(ctx context.Context, userID int64, email string) (date string, content string, err error)
	GenerateForSubscribedUsers(ctx context.Context, users []repository.UserRecord) error
	CleanupHistory(ctx context.Context, keepDays int) error
}

type fortuneService struct {
	mapper repository.FortuneMapper
	llm    llm.Client
}

// NewFortuneService 创建运势服务。
// 参数：mapper - 运势数据访问层；llmClient - 大模型客户端。
// 返回：FortuneService - 运势服务接口实现。
func NewFortuneService(mapper repository.FortuneMapper, llmClient llm.Client) FortuneService {
	return &fortuneService{mapper: mapper, llm: llmClient}
}

// GetToday 获取或生成用户今日运势。
// 参数：ctx - 上下文；userID - 用户ID；email - 用户邮箱。
// 返回：string - 日期；string - 运势内容；error - 处理失败错误。
func (s *fortuneService) GetToday(ctx context.Context, userID int64, email string) (string, string, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	content, err := s.mapper.GetByUserAndDate(ctx, userID, now)
	if err == nil {
		return today, content, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}

	content, err = s.llm.GenerateTodayFortune(ctx, email)
	if err != nil {
		return "", "", err
	}
	if err = s.mapper.Save(ctx, userID, now, content); err != nil {
		return "", "", err
	}
	return today, content, nil
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
