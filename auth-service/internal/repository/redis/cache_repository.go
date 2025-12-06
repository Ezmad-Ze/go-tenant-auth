package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ezmad/auth-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Error definitions
var (
	ErrCacheNotFound      = errors.New("cache: key not found")
	ErrCacheSerialization = errors.New("cache: serialization failed")
	ErrCacheConnection    = errors.New("cache: connection failed")
	ErrInvalidTTL         = errors.New("cache: TTL must be positive")
)

// cacheRepository implements the CacheRepository interface
type cacheRepository struct {
	client *redis.Client
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(client *redis.Client) repository.CacheRepository {
	return &cacheRepository{
		client: client,
	}
}

// Set stores a key-value pair with TTL
func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Validate TTL
	if ttl <= 0 {
		return ErrInvalidTTL
	}

	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_set").
			Str("key", key).
			Msg("Failed to serialize value")
		return fmt.Errorf("%w: %v", ErrCacheSerialization, err)
	}

	// Set in Redis with TTL
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_set").
			Str("key", key).
			Dur("ttl", ttl).
			Msg("Redis SET operation failed")
		return fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return nil
}

// Get retrieves a value by key and deserializes it into dest
func (r *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	// Get from Redis
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheNotFound
		}
		log.Error().
			Err(err).
			Str("operation", "cache_get").
			Str("key", key).
			Msg("Redis GET operation failed")
		return fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	// Deserialize JSON into dest
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_get").
			Str("key", key).
			Msg("Failed to deserialize value")
		return fmt.Errorf("%w: %v", ErrCacheSerialization, err)
	}

	return nil
}

// Delete removes one or more keys from the cache
func (r *cacheRepository) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_delete").
			Strs("keys", keys).
			Msg("Redis DEL operation failed")
		return fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return nil
}

// Exists checks if a key exists in the cache
func (r *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_exists").
			Str("key", key).
			Msg("Redis EXISTS operation failed")
		return false, fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return count > 0, nil
}

// Increment atomically increments a counter and sets TTL if it's a new key
func (r *cacheRepository) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	// Validate TTL
	if ttl <= 0 {
		return 0, ErrInvalidTTL
	}

	// Use a pipeline for atomic operations
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_increment").
			Str("key", key).
			Dur("ttl", ttl).
			Msg("Redis INCR operation failed")
		return 0, fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return incrCmd.Val(), nil
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (r *cacheRepository) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	// Validate TTL
	if ttl <= 0 {
		return false, ErrInvalidTTL
	}

	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_setnx").
			Str("key", key).
			Msg("Failed to serialize value")
		return false, fmt.Errorf("%w: %v", ErrCacheSerialization, err)
	}

	// SetNX in Redis with TTL
	success, err := r.client.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "cache_setnx").
			Str("key", key).
			Dur("ttl", ttl).
			Msg("Redis SETNX operation failed")
		return false, fmt.Errorf("%w: %v", ErrCacheConnection, err)
	}

	return success, nil
}
