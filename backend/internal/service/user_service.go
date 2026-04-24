package service

import (
	"context"

	"astrodailyweb/backend/internal/repository"
)

type UserService interface {
	Subscribe(ctx context.Context, userID int64) error
	Unsubscribe(ctx context.Context, userID int64) error
	ListSubscribedUsers(ctx context.Context) ([]repository.UserRecord, error)
	GetProfile(ctx context.Context, userID int64) (repository.UserProfile, error)
	SaveProfile(ctx context.Context, profile repository.UserProfile) error
}

type userService struct {
	mapper repository.UserMapper
}

// NewUserService 创建用户服务。
// 参数：mapper - 用户数据访问层。
// 返回：UserService - 用户服务接口实现。
func NewUserService(mapper repository.UserMapper) UserService {
	return &userService{mapper: mapper}
}

// Subscribe 设置用户为订阅状态。
// 参数：ctx - 上下文；userID - 用户ID。
// 返回：error - 更新失败错误。
func (s *userService) Subscribe(ctx context.Context, userID int64) error {
	return s.mapper.UpdateSubscription(ctx, userID, true)
}

// Unsubscribe 设置用户为取消订阅状态。
// 参数：ctx - 上下文；userID - 用户ID。
// 返回：error - 更新失败错误。
func (s *userService) Unsubscribe(ctx context.Context, userID int64) error {
	return s.mapper.UpdateSubscription(ctx, userID, false)
}

// ListSubscribedUsers 查询订阅用户列表。
// 参数：ctx - 上下文。
// 返回：[]repository.UserRecord - 用户列表；error - 查询失败错误。
func (s *userService) ListSubscribedUsers(ctx context.Context) ([]repository.UserRecord, error) {
	return s.mapper.ListSubscribedUsers(ctx)
}

// GetProfile 查询用户资料。
// 参数：ctx - 上下文；userID - 用户ID。
// 返回：repository.UserProfile - 用户资料；error - 查询失败错误。
func (s *userService) GetProfile(ctx context.Context, userID int64) (repository.UserProfile, error) {
	return s.mapper.GetProfile(ctx, userID)
}

// SaveProfile 保存用户资料。
// 参数：ctx - 上下文；profile - 用户资料。
// 返回：error - 保存失败错误。
func (s *userService) SaveProfile(ctx context.Context, profile repository.UserProfile) error {
	return s.mapper.UpsertProfile(ctx, profile)
}
