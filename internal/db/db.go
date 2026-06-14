package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init(databaseURL string, maxConns int) error {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return err
	}
	cfg.MaxConns = int32(maxConns)
	Pool, err = pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return err
	}
	return Pool.Ping(context.Background())
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
