package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ezmad/auth-service/internal/repository"
	"github.com/rs/zerolog/log"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error)
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

// rateLimiter implements the RateLimiter interface using Redis
type rateLimiter struct {
	cache repository.CacheRepository
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cache repository.CacheRepository) RateLimiter {
	return &rateLimiter{
		cache: cache,
	}
}

// Allow checks if a request is allowed based on the rate limit
// Uses a sliding window counter algorithm
func (rl *rateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error) {
	// Calculate the current window start time
	now := time.Now()
	windowStart := now.Truncate(window)
	
	// Create a unique key for this window
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	// Increment the counter for this window
	count, err := rl.cache.Increment(ctx, windowKey, window)
	if err != nil {
		// On error, fail open (allow the request) for availability
		log.Warn().
			Err(err).
			Str("key", key).
			Msg("Rate limiter failed, allowing request (fail-open)")
		
		return &RateLimitResult{
			Allowed:    true,
			Limit:      limit,
			Remaining:  limit,
			ResetAt:    windowStart.Add(window),
			RetryAfter: 0,
		}, nil
	}

	// Calculate reset time (end of current window)
	resetAt := windowStart.Add(window)
	
	// Check if limit is exceeded
	if count > int64(limit) {
		retryAfter := time.Until(resetAt)
		if retryAfter < 0 {
			retryAfter = 0
		}

		return &RateLimitResult{
			Allowed:    false,
			Limit:      limit,
			Remaining:  0,
			ResetAt:    resetAt,
			RetryAfter: retryAfter,
		}, nil
	}

	// Request is allowed
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return &RateLimitResult{
		Allowed:    true,
		Limit:      limit,
		Remaining:  remaining,
		ResetAt:    resetAt,
		RetryAfter: 0,
	}, nil
}
