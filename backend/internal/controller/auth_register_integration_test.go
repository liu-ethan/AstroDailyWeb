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

type registerMapperStub struct {
	findResult repository.UserRecord
	findErr    error
	verifyOK   bool
	verifyErr  error
	createErr  error

	createdEmail    string
	createdPassword string
	createCallCount int
}

// CleanupVerificationCodesBefore implements [repository.AuthMapper].
func (m *registerMapperStub) CleanupVerificationCodesBefore(ctx context.Context, cutoff time.Time) error {
	panic("unimplemented")
}

func (m *registerMapperStub) FindUserByEmail(ctx context.Context, email string) (repository.UserRecord, error) {
	_ = ctx
	_ = email
	if m.findErr != nil {
		return repository.UserRecord{}, m.findErr
	}
	return m.findResult, nil
}

func (m *registerMapperStub) CreateUser(ctx context.Context, email, password string) error {
	_ = ctx
	m.createCallCount++
	m.createdEmail = email
	m.createdPassword = password
	return m.createErr
}

func (m *registerMapperStub) SaveVerificationCode(ctx context.Context, email, code string, businessType int) error {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return nil
}

func (m *registerMapperStub) VerifyCode(ctx context.Context, email, code string, businessType int) (bool, error) {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	if m.verifyErr != nil {
		return false, m.verifyErr
	}
	return m.verifyOK, nil
}

func (m *registerMapperStub) UpdatePassword(ctx context.Context, email, newPassword string) error {
	_ = ctx
	_ = email
	_ = newPassword
	return nil
}

type registerSMTPStub struct{}

func (s *registerSMTPStub) Send(ctx context.Context, to []string, subject, body string) error {
	_ = ctx
	_ = to
	_ = subject
	_ = body
	return nil
}

func (s *registerSMTPStub) SendVerifyCode(ctx context.Context, to []string, code string) error {
	_ = ctx
	_ = to
	_ = code
	return nil
}

func buildRegisterTestEngine(mapper *registerMapperStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	authSvc := service.NewAuthService(
		mapper,
		&registerSMTPStub{},
		auth.NewJWTManager("secret", "issuer", time.Minute),
		&sendCodeTokenStoreStub{},
	)
	ctl := NewAuthController(authSvc)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.POST("/api/v1/auth/register", ctl.Register)
	return r
}

func TestAuthRegisterSuccess(t *testing.T) {
	mapper := &registerMapperStub{verifyOK: true, findErr: sql.ErrNoRows}
	r := buildRegisterTestEngine(mapper)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"new@example.com","password":"12345678","code":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var resp responseEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 200 || resp.Message != "注册成功" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if mapper.createCallCount != 1 {
		t.Fatalf("expected create user called once, got=%d", mapper.createCallCount)
	}
	if mapper.createdEmail != "new@example.com" || mapper.createdPassword != "12345678" {
		t.Fatalf("unexpected create args: email=%s password=%s", mapper.createdEmail, mapper.createdPassword)
	}
}

func TestAuthRegisterBadRequest(t *testing.T) {
	mapper := &registerMapperStub{}
	r := buildRegisterTestEngine(mapper)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"bad-email","password":"12345678","code":"123456"}`))
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
	if mapper.createCallCount != 0 {
		t.Fatalf("bad request should not create user, got=%d", mapper.createCallCount)
	}
}

func TestAuthRegisterPasswordTooShort(t *testing.T) {
	mapper := &registerMapperStub{verifyOK: true, findErr: sql.ErrNoRows}
	r := buildRegisterTestEngine(mapper)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"new@example.com","password":"1234567","code":"123456"}`))
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
	if resp.Code != 4002 || resp.Message != "密码不得少于8位" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthRegisterVerifyCodeFailed(t *testing.T) {
	mapper := &registerMapperStub{verifyOK: false, findErr: sql.ErrNoRows}
	r := buildRegisterTestEngine(mapper)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"new@example.com","password":"12345678","code":"000000"}`))
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
	if resp.Code != 4001 || resp.Message != "请输入正确的邮箱或验证码" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if mapper.createCallCount != 0 {
		t.Fatalf("verify code failed should not create user, got=%d", mapper.createCallCount)
	}
}

func TestAuthRegisterEmailExists(t *testing.T) {
	mapper := &registerMapperStub{verifyOK: true, findResult: repository.UserRecord{ID: 1, Email: "new@example.com"}}
	r := buildRegisterTestEngine(mapper)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"new@example.com","password":"12345678","code":"123456"}`))
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
	if resp.Code != 4004 || resp.Message != "邮箱已存在" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if mapper.createCallCount != 0 {
		t.Fatalf("email exists should not create user, got=%d", mapper.createCallCount)
	}
}

func TestAuthRegisterCreateUserError(t *testing.T) {
	mapper := &registerMapperStub{verifyOK: true, findErr: sql.ErrNoRows, createErr: errors.New("insert failed")}
	r := buildRegisterTestEngine(mapper)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"new@example.com","password":"12345678","code":"123456"}`))
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
	if resp.Code != 5000 || resp.Message != "注册失败" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
