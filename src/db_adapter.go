package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxDB struct {
	pool *pgxpool.Pool
}

func NewPGXDB(pool *pgxpool.Pool) DB {
	return &pgxDB{pool: pool}
}

func (p *pgxDB) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tag, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return int64(tag.RowsAffected()), nil
}

func (p *pgxDB) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return p.pool.QueryRow(ctx, sql, args...)
}
