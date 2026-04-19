package service

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"astrodailyweb/backend/internal/apperror"
	"astrodailyweb/backend/internal/auth"
	"astrodailyweb/backend/internal/notify"
	"astrodailyweb/backend/internal/repository"
)

type AuthService interface {
	SendCode(ctx context.Context, email string, businessType int) error
	Register(ctx context.Context, email, password, code string) error
	Login(ctx context.Context, email, password string) (string, int64, error)
	ResetPassword(ctx context.Context, email, newPassword, code string) error
}

type authService struct {
	mapper     repository.AuthMapper
	smtp       notify.SMTPClient
	jwt        *auth.JWTManager
	tokenStore auth.TokenStore
}

const (
	registerBusinessType = 1
	resetBusinessType    = 2
	minPasswordLen       = 8
)

// NewAuthService 创建认证服务。
// 参数：mapper - 认证数据访问层；smtp - 邮件发送客户端；jwt - JWT管理器；tokenStore - Token 存储。
// 返回：AuthService - 认证服务接口实现。
func NewAuthService(mapper repository.AuthMapper, smtp notify.SMTPClient, jwt *auth.JWTManager, tokenStore auth.TokenStore) AuthService {
	return &authService{mapper: mapper, smtp: smtp, jwt: jwt, tokenStore: tokenStore}
}

// SendCode 生成并发送邮箱验证码。
// 参数：ctx - 上下文；email - 用户邮箱；businessType - 业务类型。
// 返回：error - 保存或发送失败错误。
func (s *authService) SendCode(ctx context.Context, email string, businessType int) error {
	code := fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000))
	if err := s.mapper.SaveVerificationCode(ctx, email, code, businessType); err != nil {
		return err
	}
	return s.smtp.Send(ctx, []string{email}, "验证码", "您的验证码是: "+code)
}

// Register 处理注册逻辑。
// 参数：ctx - 上下文；email - 用户邮箱；password - 用户密码；code - 验证码。
// 返回：error - 处理失败错误。
func (s *authService) Register(ctx context.Context, email, password, code string) error {
	if len(password) < minPasswordLen {
		return apperror.New(4002, "密码不得少于8位")
	}
	ok, err := s.mapper.VerifyCode(ctx, email, code, registerBusinessType)
	if err != nil {
		return apperror.Wrap(5000, "验证码校验失败", err)
	}
	if !ok {
		return apperror.New(4001, "请输入正确的邮箱或验证码")
	}

	_, err = s.mapper.FindUserByEmail(ctx, email)
	if err == nil {
		return apperror.New(4004, "邮箱已存在")
	}
	if err != nil && err != sql.ErrNoRows {
		return apperror.Wrap(5000, "查询用户失败", err)
	}

	if err = s.mapper.CreateUser(ctx, email, password); err != nil {
		return apperror.Wrap(5000, "注册失败", err)
	}
	return nil
}

// Login 处理登录并生成令牌。
// 参数：ctx - 上下文；email - 用户邮箱；password - 用户密码。
// 返回：string - JWT令牌；int64 - 过期秒数；error - 处理失败错误。
func (s *authService) Login(ctx context.Context, email, password string) (string, int64, error) {
	user, err := s.mapper.FindUserByEmail(ctx, email)
	if err == sql.ErrNoRows || (err == nil && user.Password != password) {
		return "", 0, apperror.New(4005, "邮箱或密码不正确")
	}
	if err != nil {
		return "", 0, apperror.Wrap(5000, "查询用户失败", err)
	}

	token, exp, err := s.jwt.Generate(user.ID)
	if err != nil {
		return "", 0, apperror.Wrap(5000, "生成Token失败", err)
	}
	if err = s.tokenStore.Save(ctx, token, exp); err != nil {
		return "", 0, apperror.Wrap(5000, "保存Token失败", err)
	}
	return token, int64(time.Until(exp).Seconds()), nil
}

// ResetPassword 处理重置密码逻辑。
// 参数：ctx - 上下文；email - 用户邮箱；newPassword - 新密码；code - 验证码。
// 返回：error - 处理失败错误。
func (s *authService) ResetPassword(ctx context.Context, email, newPassword, code string) error {
	if len(newPassword) < minPasswordLen {
		return apperror.New(4002, "密码不得少于8位")
	}

	if _, err := s.mapper.FindUserByEmail(ctx, email); err != nil {
		if err == sql.ErrNoRows {
			return apperror.New(4001, "请输入正确的邮箱或验证码")
		}
		return apperror.Wrap(5000, "查询用户失败", err)
	}

	ok, err := s.mapper.VerifyCode(ctx, email, code, resetBusinessType)
	if err != nil {
		return apperror.Wrap(5000, "验证码校验失败", err)
	}
	if !ok {
		return apperror.New(4001, "请输入正确的邮箱或验证码")
	}

	if err = s.mapper.UpdatePassword(ctx, email, newPassword); err != nil {
		return apperror.Wrap(5000, "重置密码失败", err)
	}
	return nil
}
