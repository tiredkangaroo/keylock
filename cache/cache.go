package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tiredkangaroo/keylock/config"
	"github.com/tiredkangaroo/keylock/vault"
)

var redisClient *redis.Client
var ctx = context.Background()

var (
	ErrCmdNil error = errors.New("redis command returned was nil")
)

func Init() {
	redisClient = redis.NewClient(&redis.Options{
		Network:      config.DefaultConfig.Redis.Network,
		Addr:         config.DefaultConfig.Redis.Hostport,
		Username:     vault.GetRedisUsername(),
		Password:     vault.GetRedisPassword(),
		DB:           config.DefaultConfig.Redis.DB,
		ReadTimeout:  time.Duration(config.DefaultConfig.Redis.Timeout) * time.Second,
		WriteTimeout: time.Duration(config.DefaultConfig.Redis.Timeout) * time.Second,
	})
}

func HGet(key, field string) (string, error) {
	cmd := redisClient.HGet(ctx, key, field)
	if cmd == nil {
		return "", ErrCmdNil // this shouldn't happen but i don't trust redis
	}
	val, err := cmd.Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func HSetWithExpiration(key, field string, value string, expiration time.Duration) error {
	cmd := redisClient.HSetEXWithArgs(ctx, key, &redis.HSetEXOptions{
		Condition:      redis.HSetEXFNX, // if none of the fields exist
		ExpirationType: redis.HSetEXExpirationEX,
		ExpirationVal:  int64(expiration.Seconds()),
	}, field, value)
	if cmd == nil {
		return ErrCmdNil // this shouldn't happen but i don't trust redis
	}
	return cmd.Err()
}
