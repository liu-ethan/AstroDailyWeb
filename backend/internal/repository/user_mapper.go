package repository

import (
	"context"
	"database/sql"
)

type UserMapper interface {
	UpdateSubscription(ctx context.Context, userID int64, subscribed bool) error
	ListSubscribedUsers(ctx context.Context) ([]UserRecord, error)
}

type userMapper struct {
	db *sql.DB
}

// NewUserMapper 创建用户设置相关数据访问对象。
// 参数：db - 数据库连接池。
// 返回：UserMapper - 用户 Mapper 接口实现。
func NewUserMapper(db *sql.DB) UserMapper {
	return &userMapper{db: db}
}

// UpdateSubscription 更新用户订阅状态（占位实现）。
// 参数：ctx - 上下文；userID - 用户ID；subscribed - 是否订阅。
// 返回：error - 更新失败错误。
func (m *userMapper) UpdateSubscription(ctx context.Context, userID int64, subscribed bool) error {
	_ = ctx
	_ = userID
	_ = subscribed
	return nil
}

// ListSubscribedUsers 查询已订阅用户列表（占位实现）。
// 参数：ctx - 上下文。
// 返回：[]UserRecord - 用户列表；error - 查询失败错误。
func (m *userMapper) ListSubscribedUsers(ctx context.Context) ([]UserRecord, error) {
	_ = ctx
	return []UserRecord{}, nil
}
