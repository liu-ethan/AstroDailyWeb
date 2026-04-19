package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

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
	mapper repository.AuthMapper
	smtp   notify.SMTPClient
	jwt    *auth.JWTManager
}

// NewAuthService 创建认证服务。
// 参数：mapper - 认证数据访问层；smtp - 邮件发送客户端；jwt - JWT管理器。
// 返回：AuthService - 认证服务接口实现。
func NewAuthService(mapper repository.AuthMapper, smtp notify.SMTPClient, jwt *auth.JWTManager) AuthService {
	return &authService{mapper: mapper, smtp: smtp, jwt: jwt}
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

// Register 处理注册逻辑（当前为占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱；password - 用户密码；code - 验证码。
// 返回：error - 处理失败错误。
func (s *authService) Register(ctx context.Context, email, password, code string) error {
	_ = ctx
	_ = email
	_ = password
	_ = code
	return nil
}

// Login 处理登录并生成令牌（当前为占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱；password - 用户密码。
// 返回：string - JWT令牌；int64 - 过期秒数；error - 处理失败错误。
func (s *authService) Login(ctx context.Context, email, password string) (string, int64, error) {
	_ = ctx
	_ = email
	_ = password
	token, exp, err := s.jwt.Generate(0)
	if err != nil {
		return "", 0, err
	}
	return token, int64(time.Until(exp).Seconds()), nil
}

// ResetPassword 处理重置密码逻辑（当前为占位实现）。
// 参数：ctx - 上下文；email - 用户邮箱；newPassword - 新密码；code - 验证码。
// 返回：error - 处理失败错误。
func (s *authService) ResetPassword(ctx context.Context, email, newPassword, code string) error {
	_ = ctx
	_ = email
	_ = newPassword
	_ = code
	return nil
}
