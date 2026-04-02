// placeholder file 
// TODO: extract sql commands into separate file for better organization and testing

package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context) *pgxpool.Pool {
	dbUrl := "postgres://postgres:password@localhost:5432/postgres"
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		panic(err)
	}
	return pool
}	