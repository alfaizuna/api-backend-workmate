package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	// Normalize scheme for pgx: allow both postgresql:// and postgres://
	if strings.HasPrefix(databaseURL, "postgresql://") {
		databaseURL = "postgres://" + strings.TrimPrefix(databaseURL, "postgresql://")
	}
	if !strings.Contains(strings.ToLower(databaseURL), "sslmode=") {
		sep := "?"
		if strings.Contains(databaseURL, "?") {
			sep = "&"
		}
		databaseURL = databaseURL + sep + "sslmode=require"
	}
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	// Use default resolver/dialer; ensure sslmode=require is enabled above.
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}
