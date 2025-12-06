package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Cache key patterns
const (
	FailedLoginKeyPattern = "failed_login:%s:%s"  // failed_login:{tenant_id}:{user_id}
	AccountLockKeyPattern = "account_lock:%s:%s"  // account_lock:{tenant_id}:{user_id}
)

// FailedLoginTracker defines the interface for tracking failed login attempts
type FailedLoginTracker interface {
	RecordFailedAttempt(ctx context.Context, tenantID, userID uuid.UUID) error
	GetFailedAttempts(ctx context.Context, tenantID, userID uuid.UUID) (int, error)
	ResetFailedAttempts(ctx context.Context, tenantID, userID uuid.UUID) error
	IsAccountLocked(ctx context.Context, tenantID, userID uuid.UUID) (bool, time.Time, error)
}

// failedLoginTracker implements the FailedLoginTracker interface
type failedLoginTracker struct {
	cache        repository.CacheRepository
	maxAttempts  int
	lockDuration time.Duration
	attemptTTL   time.Duration
}

// NewFailedLoginTracker creates a new failed login tracker
func NewFailedLoginTracker(
	cache repository.CacheRepository,
	maxAttempts int,
	lockDuration time.Duration,
	attemptTTL time.Duration,
) FailedLoginTracker {
	return &failedLoginTracker{
		cache:        cache,
		maxAttempts:  maxAttempts,
		lockDuration: lockDuration,
		attemptTTL:   attemptTTL,
	}
}

// RecordFailedAttempt increments the failed attempt counter and locks account if threshold is reached
func (t *failedLoginTracker) RecordFailedAttempt(ctx context.Context, tenantID, userID uuid.UUID) error {
	key := fmt.Sprintf(FailedLoginKeyPattern, tenantID.String(), userID.String())

	// Increment the counter with TTL
	count, err := t.cache.Increment(ctx, key, t.attemptTTL)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Failed to record failed login attempt")
		return err
	}

	// Check if we should lock the account
	if int(count) >= t.maxAttempts {
		lockKey := fmt.Sprintf(AccountLockKeyPattern, tenantID.String(), userID.String())
		lockUntil := time.Now().Add(t.lockDuration)

		// Store lock with expiration
		if err := t.cache.Set(ctx, lockKey, lockUntil.Unix(), t.lockDuration); err != nil {
			log.Error().
				Err(err).
				Str("tenant_id", tenantID.String()).
				Str("user_id", userID.String()).
				Msg("Failed to lock account")
			return err
		}

		log.Warn().
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Int("attempts", int(count)).
			Time("locked_until", lockUntil).
			Msg("Account locked due to failed login attempts")
	}

	return nil
}

// GetFailedAttempts returns the number of failed login attempts
func (t *failedLoginTracker) GetFailedAttempts(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	key := fmt.Sprintf(FailedLoginKeyPattern, tenantID.String(), userID.String())

	var countStr string
	err := t.cache.Get(ctx, key, &countStr)
	if err != nil {
		// If key doesn't exist, return 0
		if err.Error() == "cache: key not found" {
			return 0, nil
		}
		log.Error().
			Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Failed to get failed login attempts")
		return 0, err
	}

	// Parse the count
	count, err := strconv.Atoi(countStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Str("value", countStr).
			Msg("Failed to parse failed login count")
		return 0, err
	}

	return count, nil
}

// ResetFailedAttempts clears the failed attempt counter
func (t *failedLoginTracker) ResetFailedAttempts(ctx context.Context, tenantID, userID uuid.UUID) error {
	key := fmt.Sprintf(FailedLoginKeyPattern, tenantID.String(), userID.String())

	if err := t.cache.Delete(ctx, key); err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Failed to reset failed login attempts")
		return err
	}

	return nil
}

// IsAccountLocked checks if an account is currently locked
func (t *failedLoginTracker) IsAccountLocked(ctx context.Context, tenantID, userID uuid.UUID) (bool, time.Time, error) {
	lockKey := fmt.Sprintf(AccountLockKeyPattern, tenantID.String(), userID.String())

	var lockUntilUnix int64
	err := t.cache.Get(ctx, lockKey, &lockUntilUnix)
	if err != nil {
		// If key doesn't exist, account is not locked
		if err.Error() == "cache: key not found" {
			return false, time.Time{}, nil
		}
		log.Error().
			Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Failed to check account lock status")
		return false, time.Time{}, err
	}

	lockUntil := time.Unix(lockUntilUnix, 0)

	// Check if lock has expired
	if time.Now().After(lockUntil) {
		// Lock has expired, clean up
		_ = t.cache.Delete(ctx, lockKey)
		return false, time.Time{}, nil
	}

	return true, lockUntil, nil
}
