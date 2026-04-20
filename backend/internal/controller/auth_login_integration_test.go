package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"astrodailyweb/backend/internal/auth"
	"astrodailyweb/backend/internal/middleware"
	"astrodailyweb/backend/internal/repository"
	"astrodailyweb/backend/internal/service"
)

type loginMapperStub struct {
	findResult repository.UserRecord
	findErr    error
}

func (m *loginMapperStub) FindUserByEmail(ctx context.Context, email string) (repository.UserRecord, error) {
	_ = ctx
	_ = email
	if m.findErr != nil {
		return repository.UserRecord{}, m.findErr
	}
	return m.findResult, nil
}

func (m *loginMapperStub) CreateUser(ctx context.Context, email, password string) error {
	_ = ctx
	_ = email
	_ = password
	return nil
}

func (m *loginMapperStub) SaveVerificationCode(ctx context.Context, email, code string, businessType int) error {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return nil
}

func (m *loginMapperStub) VerifyCode(ctx context.Context, email, code string, businessType int) (bool, error) {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return false, nil
}

func (m *loginMapperStub) UpdatePassword(ctx context.Context, email, newPassword string) error {
	_ = ctx
	_ = email
	_ = newPassword
	return nil
}

type loginSMTPStub struct{}

func (s *loginSMTPStub) Send(ctx context.Context, to []string, subject, body string) error {
	_ = ctx
	_ = to
	_ = subject
	_ = body
	return nil
}

type loginTokenStoreStub struct {
	saveErr       error
	saveCallCount int
	savedToken    string
}

func (s *loginTokenStoreStub) Save(ctx context.Context, token string, expireAt time.Time) error {
	_ = ctx
	_ = expireAt
	s.saveCallCount++
	s.savedToken = token
	return s.saveErr
}

func (s *loginTokenStoreStub) Exists(ctx context.Context, token string) (bool, error) {
	_ = ctx
	_ = token
	return true, nil
}

func (s *loginTokenStoreStub) Delete(ctx context.Context, token string) error {
	_ = ctx
	_ = token
	return nil
}

func buildLoginTestEngine(mapper *loginMapperStub, tokenStore *loginTokenStoreStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	authSvc := service.NewAuthService(
		mapper,
		&loginSMTPStub{},
		auth.NewJWTManager("secret", "issuer", time.Minute),
		tokenStore,
	)
	ctl := NewAuthController(authSvc)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.POST("/api/v1/auth/login", ctl.Login)
	return r
}

func TestAuthLoginSuccess(t *testing.T) {
	mapper := &loginMapperStub{findResult: repository.UserRecord{ID: 1, Email: "user@example.com", Password: "12345678"}}
	tokenStore := &loginTokenStoreStub{}
	r := buildLoginTestEngine(mapper, tokenStore)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp struct {
		Code    int `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token     string `json:"token"`
			ExpiresIn int64  `json:"expires_in"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 200 || resp.Message != "登录成功" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Data.Token == "" {
		t.Fatal("expected token in response")
	}
	if resp.Data.ExpiresIn <= 0 {
		t.Fatalf("expected positive expires_in, got=%d", resp.Data.ExpiresIn)
	}
	if tokenStore.saveCallCount != 1 {
		t.Fatalf("expected save token once, got=%d", tokenStore.saveCallCount)
	}
	if tokenStore.savedToken == "" {
		t.Fatal("expected saved token in token store")
	}
}

func TestAuthLoginBadRequest(t *testing.T) {
	mapper := &loginMapperStub{}
	tokenStore := &loginTokenStoreStub{}
	r := buildLoginTestEngine(mapper, tokenStore)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"bad-email","password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 4000 || resp.Message != "参数错误" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthLoginInvalidCredentialUserNotFound(t *testing.T) {
	mapper := &loginMapperStub{findErr: sql.ErrNoRows}
	tokenStore := &loginTokenStoreStub{}
	r := buildLoginTestEngine(mapper, tokenStore)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 4005 || resp.Message != "邮箱或密码不正确" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthLoginInvalidCredentialWrongPassword(t *testing.T) {
	mapper := &loginMapperStub{findResult: repository.UserRecord{ID: 1, Email: "user@example.com", Password: "wrong"}}
	tokenStore := &loginTokenStoreStub{}
	r := buildLoginTestEngine(mapper, tokenStore)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 4005 || resp.Message != "邮箱或密码不正确" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthLoginMapperError(t *testing.T) {
	mapper := &loginMapperStub{findErr: errors.New("db down")}
	tokenStore := &loginTokenStoreStub{}
	r := buildLoginTestEngine(mapper, tokenStore)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 5000 || resp.Message != "查询用户失败" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthLoginSaveTokenError(t *testing.T) {
	mapper := &loginMapperStub{findResult: repository.UserRecord{ID: 1, Email: "user@example.com", Password: "12345678"}}
	tokenStore := &loginTokenStoreStub{saveErr: errors.New("redis down")}
	r := buildLoginTestEngine(mapper, tokenStore)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 5000 || resp.Message != "保存Token失败" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

