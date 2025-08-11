package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	ConnectionString string `koanf:"connectionString"`
	pool             *pgxpool.Pool
}

func (c *DBConfig) validate() error {
	if c.ConnectionString == "" {
		return fmt.Errorf("db.connectionString is required")
	}
	return nil
}

func (c *DBConfig) GetPool(ctx context.Context) (*pgxpool.Pool, error) {
	if c.pool != nil {
		return c.pool, nil
	}

	pool, err := pgxpool.New(ctx, c.ConnectionString)
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

func (c *DBConfig) GetConnection(ctx context.Context) (*pgx.Conn, error) {
	pool, err := c.GetPool(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection from pool: %w", err)
	}

	return conn.Conn(), nil
}

func (c *DBConfig) Close() {
	if c.pool != nil {
		c.pool.Close()
		c.pool = nil
	}
}
