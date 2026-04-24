package controller

import (
	"context"
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

type sendCodeMapperStub struct {
	saveErr       error
	savedEmail    string
	savedCode     string
	savedBizType  int
	saveCallCount int
}

func (m *sendCodeMapperStub) FindUserByEmail(ctx context.Context, email string) (repository.UserRecord, error) {
	_ = ctx
	_ = email
	return repository.UserRecord{}, nil
}

func (m *sendCodeMapperStub) CreateUser(ctx context.Context, email, password string) error {
	_ = ctx
	_ = email
	_ = password
	return nil
}

func (m *sendCodeMapperStub) SaveVerificationCode(ctx context.Context, email, code string, businessType int) error {
	_ = ctx
	m.saveCallCount++
	m.savedEmail = email
	m.savedCode = code
	m.savedBizType = businessType
	return m.saveErr
}

func (m *sendCodeMapperStub) VerifyCode(ctx context.Context, email, code string, businessType int) (bool, error) {
	_ = ctx
	_ = email
	_ = code
	_ = businessType
	return false, nil
}

func (m *sendCodeMapperStub) UpdatePassword(ctx context.Context, email, newPassword string) error {
	_ = ctx
	_ = email
	_ = newPassword
	return nil
}

type sendCodeSMTPStub struct {
	sendErr       error
	sentTo        []string
	sentSubject   string
	sentBody      string
	sendCallCount int
}

func (s *sendCodeSMTPStub) Send(ctx context.Context, to []string, subject, body string) error {
	_ = ctx
	s.sendCallCount++
	s.sentTo = append([]string{}, to...)
	s.sentSubject = subject
	s.sentBody = body
	return s.sendErr
}

func (s *sendCodeSMTPStub) SendVerifyCode(ctx context.Context, to []string, code string) error {
	_ = ctx
	s.sendCallCount++
	s.sentTo = append([]string{}, to...)
	s.sentSubject = ""
	s.sentBody = code
	return s.sendErr
}

type sendCodeTokenStoreStub struct{}

func (s *sendCodeTokenStoreStub) Save(ctx context.Context, token string, expireAt time.Time) error {
	_ = ctx
	_ = token
	_ = expireAt
	return nil
}

func (s *sendCodeTokenStoreStub) Exists(ctx context.Context, token string) (bool, error) {
	_ = ctx
	_ = token
	return true, nil
}

func (s *sendCodeTokenStoreStub) Delete(ctx context.Context, token string) error {
	_ = ctx
	_ = token
	return nil
}

type responseEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func buildSendCodeTestEngine(mapper *sendCodeMapperStub, smtp *sendCodeSMTPStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	authSvc := service.NewAuthService(
		mapper,
		smtp,
		auth.NewJWTManager("secret", "issuer", time.Minute),
		&sendCodeTokenStoreStub{},
	)
	ctl := NewAuthController(authSvc)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.POST("/api/v1/auth/send-code", ctl.SendCode)
	return r
}

func TestAuthSendCodeSuccess(t *testing.T) {
	mapper := &sendCodeMapperStub{}
	smtp := &sendCodeSMTPStub{}
	r := buildSendCodeTestEngine(mapper, smtp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-code", strings.NewReader(`{"email":"user@example.com","business_type":1}`))
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
	if resp.Code != 200 || resp.Message != "验证码发送成功" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if mapper.saveCallCount != 1 {
		t.Fatalf("expected save code called once, got=%d", mapper.saveCallCount)
	}
	if mapper.savedEmail != "user@example.com" || mapper.savedBizType != 1 {
		t.Fatalf("unexpected saved args: email=%s businessType=%d", mapper.savedEmail, mapper.savedBizType)
	}
	if len(mapper.savedCode) != 6 {
		t.Fatalf("unexpected verification code length: %s", mapper.savedCode)
	}
	for _, ch := range mapper.savedCode {
		if ch < '0' || ch > '9' {
			t.Fatalf("verification code should be numeric: %s", mapper.savedCode)
		}
	}
	if smtp.sendCallCount != 1 {
		t.Fatalf("expected smtp send once, got=%d", smtp.sendCallCount)
	}
	if len(smtp.sentTo) != 1 || smtp.sentTo[0] != "user@example.com" {
		t.Fatalf("unexpected receiver: %v", smtp.sentTo)
	}
	if !strings.Contains(smtp.sentBody, mapper.savedCode) {
		t.Fatalf("smtp body should contain code, body=%s code=%s", smtp.sentBody, mapper.savedCode)
	}
}

func TestAuthSendCodeBadRequest(t *testing.T) {
	mapper := &sendCodeMapperStub{}
	smtp := &sendCodeSMTPStub{}
	r := buildSendCodeTestEngine(mapper, smtp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-code", strings.NewReader(`{"email":"bad-email","business_type":3}`))
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
	if mapper.saveCallCount != 0 || smtp.sendCallCount != 0 {
		t.Fatalf("bad request should not call downstream, save=%d smtp=%d", mapper.saveCallCount, smtp.sendCallCount)
	}
}

func TestAuthSendCodeMapperError(t *testing.T) {
	mapper := &sendCodeMapperStub{saveErr: errors.New("db down")}
	smtp := &sendCodeSMTPStub{}
	r := buildSendCodeTestEngine(mapper, smtp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-code", strings.NewReader(`{"email":"user@example.com","business_type":2}`))
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
	if resp.Code != 5000 || resp.Message != "系统繁忙，请稍后重试" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if smtp.sendCallCount != 0 {
		t.Fatalf("save code failed should not send mail, smtp=%d", smtp.sendCallCount)
	}
}

func TestAuthSendCodeSMTPError(t *testing.T) {
	mapper := &sendCodeMapperStub{}
	smtp := &sendCodeSMTPStub{sendErr: errors.New("smtp down")}
	r := buildSendCodeTestEngine(mapper, smtp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-code", strings.NewReader(`{"email":"user@example.com","business_type":2}`))
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
	if resp.Code != 5000 || resp.Message != "系统繁忙，请稍后重试" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if mapper.saveCallCount != 1 {
		t.Fatalf("save code should still be called once, got=%d", mapper.saveCallCount)
	}
	if smtp.sendCallCount != 1 {
		t.Fatalf("smtp should be called once, got=%d", smtp.sendCallCount)
	}
}
