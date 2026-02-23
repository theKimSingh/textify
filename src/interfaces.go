package main

import (
	"context"
	"time"
)

type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (int64, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
}

type Row interface {
	Scan(dest ...any) error
}

type RedisClient interface {
	LPush(ctx context.Context, key string, values ...interface{}) error
	BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error)
	Close() error
	Context() context.Context
}
