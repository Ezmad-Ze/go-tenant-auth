package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Store struct {
	*pgxpool.Pool
	queries *db.Queries
}

func NewStore(ctx context.Context, cfg config.DatabaseConfig) (*Store, error) {
	pool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	pool.Config().MaxConns = int32(cfg.MaxOpenConns)
	pool.Config().MinConns = 2
	pool.Config().MaxConnLifetime = cfg.ConnMaxLifetime
	pool.Config().MaxConnIdleTime = 30 * time.Minute

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	queries := db.New(pool)

	log.Info().
		Str("url", maskURL(cfg.URL)).
		Int("max_conns", cfg.MaxOpenConns).
		Msg("Database connection established")

	return &Store{Pool: pool, queries: queries}, nil
}

func (s *Store) Querier() *db.Queries {
	return s.queries
}

func (s *Store) Close() {
	if s.Pool != nil {
		s.Pool.Close()
		log.Info().Msg("Database connection pool closed")
	}
}

func maskURL(url string) string {
	if len(url) < 10 {
		return "****"
	}
	return url[:10] + "..."
}