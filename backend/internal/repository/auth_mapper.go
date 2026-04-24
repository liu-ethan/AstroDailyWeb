package repository

import (
	"context"
	"database/sql"
	"time"
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

// FindUserByEmail 按邮箱查询用户。
// 参数：ctx - 上下文；email - 用户邮箱。
// 返回：UserRecord - 用户记录；error - 查询错误。
func (m *authMapper) FindUserByEmail(ctx context.Context, email string) (UserRecord, error) {
	const query = `SELECT id, email, password, is_subscribed FROM users WHERE email = ? LIMIT 1`
	var user UserRecord
	if err := m.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.IsSubscribed); err != nil {
		return UserRecord{}, err
	}
	return user, nil
}

// CreateUser 新增用户记录。
// 参数：ctx - 上下文；email - 用户邮箱；password - 用户密码。
// 返回：error - 写入失败错误。
func (m *authMapper) CreateUser(ctx context.Context, email, password string) error {
	const stmt = `INSERT INTO users (email, password, is_subscribed) VALUES (?, ?, 0)`
	_, err := m.db.ExecContext(ctx, stmt, email, password)
	return err
}

// SaveVerificationCode 保存验证码记录。
// 参数：ctx - 上下文；email - 用户邮箱；code - 验证码；businessType - 业务类型。
// 返回：error - 写入失败错误。
func (m *authMapper) SaveVerificationCode(ctx context.Context, email, code string, businessType int) error {
	const stmt = `INSERT INTO verification_codes (email, code, business_type, expires_at) VALUES (?, ?, ?, ?)`
	expireAt := time.Now().Add(5 * time.Minute)
	_, err := m.db.ExecContext(ctx, stmt, email, code, businessType, expireAt)
	return err
}

// VerifyCode 校验验证码有效性。
// 参数：ctx - 上下文；email - 用户邮箱；code - 验证码；businessType - 业务类型。
// 返回：bool - 是否通过校验；error - 查询错误。
func (m *authMapper) VerifyCode(ctx context.Context, email, code string, businessType int) (bool, error) {
	const query = `
SELECT id
FROM verification_codes
WHERE email = ? AND code = ? AND business_type = ? AND expires_at > NOW()
ORDER BY id DESC
LIMIT 1`
	var id int64
	err := m.db.QueryRowContext(ctx, query, email, code, businessType).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdatePassword 更新用户密码。
// 参数：ctx - 上下文；email - 用户邮箱；newPassword - 新密码。
// 返回：error - 更新失败错误。
func (m *authMapper) UpdatePassword(ctx context.Context, email, newPassword string) error {
	const stmt = `UPDATE users SET password = ? WHERE email = ?`
	result, err := m.db.ExecContext(ctx, stmt, newPassword, email)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
