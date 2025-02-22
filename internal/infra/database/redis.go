package database

import (
	"strconv"

	"github.com/jhonasalves/go-expert-fc-rate-limiter/configs"
	"github.com/redis/go-redis/v9"
)

type RedisDatabase struct {
	Client *redis.Client
}

func NewRedisDatabase(cfg *configs.Conf) *RedisDatabase {
	return &RedisDatabase{
		Client: redis.NewClient(&redis.Options{
			Addr:     cfg.RedisHost + ":" + strconv.Itoa(cfg.RedisPort),
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		}),
	}
}
