package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"astrodailyweb/backend/internal/llm"
	"astrodailyweb/backend/internal/repository"
)

type fakeFortuneMapper struct {
	content    string
	getErr     error
	saveErr    error
	saved      bool
	cleanupErr error
}

// DeleteExceptDate implements [repository.FortuneMapper].
func (m *fakeFortuneMapper) DeleteExceptDate(ctx context.Context, date time.Time) error {
	panic("unimplemented")
}

// DeleteByDate implements [repository.FortuneMapper].
func (m *fakeFortuneMapper) DeleteByDate(ctx context.Context, date time.Time) error {
	panic("unimplemented")
}

func (m *fakeFortuneMapper) GetByUserAndDate(ctx context.Context, userID int64, date time.Time) (string, error) {
	_ = ctx
	_ = userID
	_ = date
	if m.getErr != nil {
		return "", m.getErr
	}
	return m.content, nil
}

func (m *fakeFortuneMapper) Save(ctx context.Context, userID int64, date time.Time, content string) error {
	_ = ctx
	_ = userID
	_ = date
	_ = content
	m.saved = true
	return m.saveErr
}

func (m *fakeFortuneMapper) CleanupBefore(ctx context.Context, cutoffDate time.Time) error {
	_ = ctx
	_ = cutoffDate
	return m.cleanupErr
}

type fakeUserMapper struct {
	profile repository.UserProfile
	err     error
}

func (m *fakeUserMapper) UpdateSubscription(ctx context.Context, userID int64, subscribed bool) error {
	_ = ctx
	_ = userID
	_ = subscribed
	return nil
}

func (m *fakeUserMapper) ListSubscribedUsers(ctx context.Context) ([]repository.UserRecord, error) {
	_ = ctx
	return nil, nil
}

func (m *fakeUserMapper) GetProfile(ctx context.Context, userID int64) (repository.UserProfile, error) {
	_ = ctx
	_ = userID
	if m.err != nil {
		return repository.UserProfile{}, m.err
	}
	return m.profile, nil
}

func (m *fakeUserMapper) UpsertProfile(ctx context.Context, profile repository.UserProfile) error {
	_ = ctx
	_ = profile
	return nil
}

type fakeLLMClient struct {
	content string
	err     error
}

func (c *fakeLLMClient) GenerateTodayFortune(ctx context.Context, profile llm.FortuneProfile) (string, error) {
	_ = ctx
	_ = profile
	if c.err != nil {
		return "", c.err
	}
	return c.content, nil
}

type fakeSMTPClient struct {
	sentTo []string
	err    error
}

func (c *fakeSMTPClient) Send(ctx context.Context, to []string, subject, body string) error {
	_ = ctx
	_ = subject
	_ = body
	if c.err != nil {
		return c.err
	}
	c.sentTo = append(c.sentTo, to...)
	return nil
}

func (c *fakeSMTPClient) SendVerifyCode(ctx context.Context, to []string, code string) error {
	_ = ctx
	_ = to
	_ = code
	return nil
}

func TestFortuneServiceGetTodayFromCache(t *testing.T) {
	svc := NewFortuneService(
		&fakeFortuneMapper{content: "cached", getErr: nil},
		&fakeUserMapper{},
		&fakeLLMClient{content: "llm"},
		&fakeSMTPClient{},
	)
	_, content, err := svc.GetToday(context.Background(), 1)
	if err != nil {
		t.Fatalf("get today failed: %v", err)
	}
	if content != "cached" {
		t.Fatalf("unexpected content: %s", content)
	}
}

func TestFortuneServiceGetTodayGenerateAndSave(t *testing.T) {
	mapper := &fakeFortuneMapper{getErr: sql.ErrNoRows}
	svc := NewFortuneService(
		mapper,
		&fakeUserMapper{profile: repository.UserProfile{Birthday: "1999-08-12", Constellation: "狮子座", Gender: "男", City: "上海", Occupation: "产品"}},
		&fakeLLMClient{content: "generated"},
		&fakeSMTPClient{},
	)
	_, content, err := svc.GetToday(context.Background(), 1)
	if err != nil {
		t.Fatalf("get today failed: %v", err)
	}
	if content != "generated" {
		t.Fatalf("unexpected generated content: %s", content)
	}
	if !mapper.saved {
		t.Fatal("expected generated fortune saved")
	}
}

func TestFortuneServiceGetTodayMissingProfile(t *testing.T) {
	svc := NewFortuneService(
		&fakeFortuneMapper{getErr: sql.ErrNoRows},
		&fakeUserMapper{profile: repository.UserProfile{Birthday: "", Constellation: "狮子座", Gender: "男", City: "上海", Occupation: "产品"}},
		&fakeLLMClient{content: "generated"},
		&fakeSMTPClient{},
	)
	_, _, err := svc.GetToday(context.Background(), 1)
	if err == nil {
		t.Fatal("expected profile validation error")
	}
	if !strings.Contains(err.Error(), "请先完善运势资料") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFortuneServiceGenerateForSubscribedUsers(t *testing.T) {
	smtp := &fakeSMTPClient{}
	svc := NewFortuneService(
		&fakeFortuneMapper{getErr: sql.ErrNoRows},
		&fakeUserMapper{profile: repository.UserProfile{Birthday: "1999-08-12", Constellation: "狮子座", Gender: "男", City: "上海", Occupation: "产品"}},
		&fakeLLMClient{content: "generated"},
		smtp,
	)
	users := []repository.UserRecord{{ID: 1, Email: "a@example.com"}, {ID: 2, Email: "b@example.com"}}
	if err := svc.GenerateForSubscribedUsers(context.Background(), users); err != nil {
		t.Fatalf("generate for subscribed users failed: %v", err)
	}
	if len(smtp.sentTo) != 2 {
		t.Fatalf("unexpected send count: %d", len(smtp.sentTo))
	}
}

func TestFortuneServiceGenerateForSubscribedUsersPartialFailed(t *testing.T) {
	svc := NewFortuneService(
		&fakeFortuneMapper{getErr: sql.ErrNoRows},
		&fakeUserMapper{err: errors.New("db error")},
		&fakeLLMClient{content: "generated"},
		&fakeSMTPClient{},
	)
	err := svc.GenerateForSubscribedUsers(context.Background(), []repository.UserRecord{{ID: 1, Email: "a@example.com"}})
	if err == nil {
		t.Fatal("expected aggregate error")
	}
}
