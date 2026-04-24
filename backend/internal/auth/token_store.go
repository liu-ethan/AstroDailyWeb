package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenStore interface {
	Save(ctx context.Context, token string, expireAt time.Time) error
	Exists(ctx context.Context, token string) (bool, error)
	Delete(ctx context.Context, token string) error
}

type RedisTokenStore struct {
	client    *redis.Client
	keyPrefix string
}

// NewRedisTokenStore 创建基于 Redis 的 Token 存储。
// 参数：client - Redis 客户端；keyPrefix - 键前缀。
// 返回：*RedisTokenStore - Token 存储实例。
func NewRedisTokenStore(client *redis.Client, keyPrefix string) *RedisTokenStore {
	if keyPrefix == "" {
		keyPrefix = "astro:jwt:"
	}
	return &RedisTokenStore{client: client, keyPrefix: keyPrefix}
}

// Save 保存 JWT 到 Redis，并设置过期时间。
// 参数：ctx - 上下文；token - JWT 字符串；expireAt - 过期时间。
// 返回：error - 保存失败错误。
func (s *RedisTokenStore) Save(ctx context.Context, token string, expireAt time.Time) error {
	ttl := time.Until(expireAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}
	return s.client.Set(ctx, s.key(token), "1", ttl).Err()
}

// Exists 判断 JWT 是否存在于 Redis。
// 参数：ctx - 上下文；token - JWT 字符串。
// 返回：bool - 是否存在；error - 查询失败错误。
func (s *RedisTokenStore) Exists(ctx context.Context, token string) (bool, error) {
	n, err := s.client.Exists(ctx, s.key(token)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Delete 删除 Redis 中的 JWT。
// 参数：ctx - 上下文；token - JWT 字符串。
// 返回：error - 删除失败错误。
func (s *RedisTokenStore) Delete(ctx context.Context, token string) error {
	return s.client.Del(ctx, s.key(token)).Err()
}

func (s *RedisTokenStore) key(token string) string {
	return s.keyPrefix + token
}
