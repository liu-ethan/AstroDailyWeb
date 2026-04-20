package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"astrodailyweb/backend/internal/auth"
	"astrodailyweb/backend/internal/repository"
)

type fakeAuthMapper struct {
	user            repository.UserRecord
	findErr         error
	verifyResult    bool
	verifyErr       error
	createErr       error
	updatePassword  string
	updatePassErr   error
	createdEmail    string
	createdPassword string
}

func (m *fakeAuthMapper) FindUserByEmail(ctx context.Context, email string) (repository.UserRecord, error) {
	_ = ctx
	_ = email
	if m.findErr != nil {
		return repository.UserRecord{}, m.findErr
	}
	return m.user, nil
}

func (m *fakeAuthMapper) CreateUser(ctx context.Context, email, password string) error {
	_ = ctx
	m.createdEmail = email
	m.createdPassword = password
	return m.createErr
}

func (m *fakeAuthMapper) SaveVerificationCode(ctx context.Context, email, code string, businessType int) error {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return nil
}

func (m *fakeAuthMapper) VerifyCode(ctx context.Context, email, code string, businessType int) (bool, error) {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return m.verifyResult, m.verifyErr
}

func (m *fakeAuthMapper) UpdatePassword(ctx context.Context, email, newPassword string) error {
	_ = ctx
	_ = email
	m.updatePassword = newPassword
	return m.updatePassErr
}

type fakeSMTP struct{}

func (f *fakeSMTP) Send(ctx context.Context, to []string, subject, body string) error {
	_ = ctx
	_ = to
	_ = subject
	_ = body
	return nil
}

type fakeTokenStore struct {
	savedToken string
	expireAt   time.Time
	saveErr    error
}

func (s *fakeTokenStore) Save(ctx context.Context, token string, expireAt time.Time) error {
	_ = ctx
	s.savedToken = token
	s.expireAt = expireAt
	return s.saveErr
}

func (s *fakeTokenStore) Exists(ctx context.Context, token string) (bool, error) {
	_ = ctx
	_ = token
	return true, nil
}

func (s *fakeTokenStore) Delete(ctx context.Context, token string) error {
	_ = ctx
	_ = token
	return nil
}

func TestAuthServiceRegister(t *testing.T) {
	mapper := &fakeAuthMapper{findErr: sql.ErrNoRows, verifyResult: true}
	svc := NewAuthService(mapper, &fakeSMTP{}, auth.NewJWTManager("secret", "issuer", time.Minute), &fakeTokenStore{})

	if err := svc.Register(context.Background(), "user@example.com", "12345678", "123456"); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if mapper.createdEmail != "user@example.com" {
		t.Fatalf("unexpected created email: %s", mapper.createdEmail)
	}
}

func TestAuthServiceLoginInvalidCredential(t *testing.T) {
	mapper := &fakeAuthMapper{findErr: sql.ErrNoRows}
	svc := NewAuthService(mapper, &fakeSMTP{}, auth.NewJWTManager("secret", "issuer", time.Minute), &fakeTokenStore{})

	_, _, err := svc.Login(context.Background(), "notfound@example.com", "pwd")
	if err == nil {
		t.Fatal("expected login error")
	}
}

func TestAuthServiceResetPassword(t *testing.T) {
	mapper := &fakeAuthMapper{
		user:         repository.UserRecord{ID: 1, Email: "user@example.com"},
		verifyResult: true,
	}
	svc := NewAuthService(mapper, &fakeSMTP{}, auth.NewJWTManager("secret", "issuer", time.Minute), &fakeTokenStore{})

	if err := svc.ResetPassword(context.Background(), "user@example.com", "12345678", "123456"); err != nil {
		t.Fatalf("reset password failed: %v", err)
	}
	if mapper.updatePassword != "12345678" {
		t.Fatalf("unexpected password update: %s", mapper.updatePassword)
	}
}

func TestAuthServiceLoginSaveTokenFailed(t *testing.T) {
	mapper := &fakeAuthMapper{user: repository.UserRecord{ID: 9, Email: "u@x.com", Password: "12345678"}}
	store := &fakeTokenStore{saveErr: errors.New("redis down")}
	svc := NewAuthService(mapper, &fakeSMTP{}, auth.NewJWTManager("secret", "issuer", time.Minute), store)

	_, _, err := svc.Login(context.Background(), "u@x.com", "12345678")
	if err == nil {
		t.Fatal("expected save token error")
	}
}
