package db

import (
	"testing"

	"github.com/alicebob/miniredis/v2"

	"astrodailyweb/backend/internal/config"
)

func TestNewRedis(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	defer mini.Close()

	client, err := NewRedis(config.RedisConfig{Addr: mini.Addr(), DB: 0})
	if err != nil {
		t.Fatalf("new redis failed: %v", err)
	}
	defer func() { _ = client.Close() }()
}
