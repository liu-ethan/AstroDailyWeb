package repository

import (
	"context"
	"database/sql"
)

type AuthMapper interface {
	FindUserByEmail(ctx context.Context, email string) (UserRecord, error)
	CreateUser(ctx context.Context, email, password string) error
	SaveVerificationCode(ctx context.Context, email, code string, businessType int) error
	VerifyCode(ctx context.Context, email, code string, businessType int) (bool, error)
	UpdatePassword(ctx context.Context, email, newPassword string) error
}

type UserRecord struct {
	ID           int64
	Email        string
	Password     string
	IsSubscribed bool
}

type authMapper struct {
	db *sql.DB
}

// NewAuthMapper 创建认证相关数据访问对象。
// 参数：db - 数据库连接池。
// 返回：AuthMapper - 认证 Mapper 接口实现。
func NewAuthMapper(db *sql.DB) AuthMapper {
	return &authMapper{db: db}
}

// FindUserByEmail 按邮箱查询用户（占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱。
// 返回：UserRecord - 用户记录；error - 查询错误。
func (m *authMapper) FindUserByEmail(ctx context.Context, email string) (UserRecord, error) {
	_ = ctx
	_ = email
	return UserRecord{}, sql.ErrNoRows
}

// CreateUser 新增用户记录（占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱；password - 用户密码。
// 返回：error - 写入失败错误。
func (m *authMapper) CreateUser(ctx context.Context, email, password string) error {
	_ = ctx
	_ = email
	_ = password
	return nil
}

// SaveVerificationCode 保存验证码记录（占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱；code - 验证码；businessType - 业务类型。
// 返回：error - 写入失败错误。
func (m *authMapper) SaveVerificationCode(ctx context.Context, email, code string, businessType int) error {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return nil
}

// VerifyCode 校验验证码有效性（占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱；code - 验证码；businessType - 业务类型。
// 返回：bool - 是否通过校验；error - 查询错误。
func (m *authMapper) VerifyCode(ctx context.Context, email, code string, businessType int) (bool, error) {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return false, nil
}

// UpdatePassword 更新用户密码（占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱；newPassword - 新密码。
// 返回：error - 更新失败错误。
func (m *authMapper) UpdatePassword(ctx context.Context, email, newPassword string) error {
	_ = ctx
	_ = email
	_ = newPassword
	return nil
}
