package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type goRedisClient struct {
	c *redis.Client
}

func NewGoRedisClient(c *redis.Client) RedisClient {
	return &goRedisClient{c: c}
}

func (g *goRedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return g.c.LPush(ctx, key, values...).Err()
}

func (g *goRedisClient) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return g.c.BLPop(ctx, timeout, keys...).Result()
}

func (g *goRedisClient) Close() error             { return g.c.Close() }
func (g *goRedisClient) Context() context.Context { return context.Background() }
