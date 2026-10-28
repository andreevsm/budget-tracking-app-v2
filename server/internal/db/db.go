package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(connectCtx, dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	pingErr := pool.Ping(connectCtx)

	if pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping: %w", pingErr)
	}

	return pool, nil
}
