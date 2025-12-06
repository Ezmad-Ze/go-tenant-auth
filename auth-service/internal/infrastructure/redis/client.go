package redis

import (
	"context"
	"fmt"

	"github.com/ezmad/auth-service/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Client wraps the Redis client with configuration
type Client struct {
	client *redis.Client
	config config.RedisConfig
}

// NewClient creates a new Redis client with connection pooling
func NewClient(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	// Parse Redis options from URL
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Apply configuration
	opts.Password = cfg.Password
	opts.DB = cfg.DB
	opts.MaxRetries = cfg.MaxRetries
	opts.PoolSize = cfg.PoolSize
	opts.MinIdleConns = cfg.MinIdleConns
	opts.ConnMaxIdleTime = cfg.ConnMaxIdleTime
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout

	// Create client
	client := redis.NewClient(opts)

	// Verify connectivity with PING
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Info().
		Str("url", maskURL(cfg.URL)).
		Int("pool_size", cfg.PoolSize).
		Int("min_idle_conns", cfg.MinIdleConns).
		Msg("Redis connection established")

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// Ping verifies the connection to Redis
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close gracefully closes all Redis connections
func (c *Client) Close() error {
	if c.client != nil {
		log.Info().Msg("Closing Redis connection pool")
		return c.client.Close()
	}
	return nil
}

// GetClient returns the underlying Redis client for direct access
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// maskURL masks sensitive information in the Redis URL for logging
func maskURL(url string) string {
	if len(url) < 10 {
		return "****"
	}
	// Show only the protocol and first few characters
	return url[:10] + "..."
}
