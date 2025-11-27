package redis

import (
	"bell_best/setting"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func Init(cfg *setting.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Host,
			cfg.Port,
		),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	_, err = client.Ping(context.Background()).Result()
	return err
}

func Close() {
	_ = client.Close()
}
