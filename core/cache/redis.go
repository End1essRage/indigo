package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const defaultExpiration = 1 * time.Minute

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, pwd string, db int) (*RedisCache, error) {
	c := &RedisCache{}
	c.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	_, err := c.client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *RedisCache) GetString(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	switch {
	case err == redis.Nil:
		logrus.Infof("no entry with key %s", key)
	case err != nil:
		logrus.Errorf("error reading from redis %v", err)
	case val == "":
		logrus.Info("value is empty")
	}

	return val
}

func (c *RedisCache) SetString(key string, val string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.client.Set(ctx, key, val, defaultExpiration).Err()
}

func (c *RedisCache) Exists(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	return err == nil && val != ""
}
