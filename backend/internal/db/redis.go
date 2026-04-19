package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"astrodailyweb/backend/internal/config"
)

// NewRedis 创建 Redis 客户端并执行连通性检测。
// 参数：cfg - Redis 配置。
// 返回：*redis.Client - Redis 客户端；error - 初始化失败错误。
func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
