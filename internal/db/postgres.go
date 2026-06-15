package db

import (
	"context"
	"fmt"

	"github.com/dosu-logi/logistics-erp/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func EnsureDatabase(ctx context.Context, cfg *config.Config) error {
	adminPool, err := pgxpool.New(ctx, cfg.AdminDSN())
	if err != nil {
		return fmt.Errorf("connect admin postgres: %w", err)
	}
	defer adminPool.Close()

	var exists bool
	err = adminPool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.DBName,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check database: %w", err)
	}
	if !exists {
		_, err = adminPool.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.DBName))
		if err != nil {
			return fmt.Errorf("create database: %w", err)
		}
	}
	return nil
}
