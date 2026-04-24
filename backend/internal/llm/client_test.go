package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAICompatibleClientGenerateTodayFortune(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"今日运势星级：★★★☆"}}]}`))
	}))
	defer srv.Close()

	client := NewOpenAICompatibleClient("test-key", srv.URL, "gpt-4o-mini", 3*time.Second)
	client.templatePath = "prompt_template.yaml"

	content, err := client.GenerateTodayFortune(context.Background(), FortuneProfile{
		Birthday:      "1999-08-12",
		Today:         "2026-04-19",
		Constellation: "狮子座",
		Gender:        "男",
		City:          "上海",
		Occupation:    "产品经理",
	})
	if err != nil {
		t.Fatalf("generate fortune failed: %v", err)
	}
	if !strings.Contains(content, "今日运势星级") {
		t.Fatalf("unexpected content: %s", content)
	}
}

func TestOpenAICompatibleClientGenerateTodayFortuneBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := NewOpenAICompatibleClient("test-key", srv.URL, "gpt-4o-mini", 3*time.Second)
	client.templatePath = "prompt_template.yaml"

	_, err := client.GenerateTodayFortune(context.Background(), FortuneProfile{Today: "2026-04-19"})
	if err == nil {
		t.Fatal("expected error on bad status")
	}
}
