package config

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	ConnectionString string `koanf:"connectionstring"`
	pool             *pgxpool.Pool
}

func (c *DBConfig) Validate() error {
	if c.ConnectionString == "" {
		return fmt.Errorf("db.connectionString is required")
	}
	return nil
}

func (c *DBConfig) GetPool(ctx context.Context) (*pgxpool.Pool, error) {
	if c.pool != nil {
		return c.pool, nil
	}

	cfg, err := pgxpool.ParseConfig(c.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db.connectionString: %w", err)
	}

	// Increase pool size for concurrent scraping workloads
	// Default is max(4, NumCPU) which is too small
	if cfg.MaxConns == 0 || cfg.MaxConns < 100 {
		cfg.MaxConns = 100
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithIncludeQueryParameters(), otelpgx.WithTrimSQLInSpanName())

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	c.pool = pool
	return pool, nil
}
