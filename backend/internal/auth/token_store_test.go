package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisTokenStoreSaveExistsDelete(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer func() { _ = client.Close() }()

	store := NewRedisTokenStore(client, "test:jwt:")
	ctx := context.Background()
	token := "abc.def.ghi"
	if err = store.Save(ctx, token, time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("save token failed: %v", err)
	}

	exists, err := store.Exists(ctx, token)
	if err != nil {
		t.Fatalf("exists token failed: %v", err)
	}
	if !exists {
		t.Fatal("expected token exists")
	}

	if err = store.Delete(ctx, token); err != nil {
		t.Fatalf("delete token failed: %v", err)
	}

	exists, err = store.Exists(ctx, token)
	if err != nil {
		t.Fatalf("exists token after delete failed: %v", err)
	}
	if exists {
		t.Fatal("expected token deleted")
	}
}

func TestRedisTokenStoreSaveExpired(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer func() { _ = client.Close() }()

	store := NewRedisTokenStore(client, "")
	if err = store.Save(context.Background(), "expired", time.Now().Add(-time.Second)); err == nil {
		t.Fatal("expected error for expired token")
	}
}
