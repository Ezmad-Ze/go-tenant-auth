package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
)

// TOTPRateLimiter handles rate limiting for TOTP verification attempts
type TOTPRateLimiter struct {
	cacheRepo         repository.CacheRepository
	maxAttempts       int
	lockoutDuration   time.Duration
	attemptWindowTTL  time.Duration
}

// NewTOTPRateLimiter creates a new TOTP rate limiter
func NewTOTPRateLimiter(
	cacheRepo repository.CacheRepository,
	maxAttempts int,
	lockoutDuration time.Duration,
	attemptWindowTTL time.Duration,
) *TOTPRateLimiter {
	return &TOTPRateLimiter{
		cacheRepo:        cacheRepo,
		maxAttempts:      maxAttempts,
		lockoutDuration:  lockoutDuration,
		attemptWindowTTL: attemptWindowTTL,
	}
}

// RecordFailedTOTPAttempt records a failed TOTP verification attempt
func (r *TOTPRateLimiter) RecordFailedTOTPAttempt(ctx context.Context, tenantID, userID uuid.UUID) error {
	key := fmt.Sprintf("totp:failed:%s:%s", tenantID.String(), userID.String())
	
	// Increment failed attempts counter
	attempts, err := r.cacheRepo.Increment(ctx, key, r.attemptWindowTTL)
	if err != nil {
		return fmt.Errorf("failed to record TOTP attempt: %w", err)
	}

	// If max attempts exceeded, set lockout
	if attempts >= int64(r.maxAttempts) {
		lockoutKey := fmt.Sprintf("totp:locked:%s:%s", tenantID.String(), userID.String())
		if err := r.cacheRepo.Set(ctx, lockoutKey, true, r.lockoutDuration); err != nil {
			return fmt.Errorf("failed to set TOTP lockout: %w", err)
		}
	}

	return nil
}

// IsAccountLockedForTOTP checks if the account is locked due to too many failed TOTP attempts
func (r *TOTPRateLimiter) IsAccountLockedForTOTP(ctx context.Context, tenantID, userID uuid.UUID) (bool, time.Time, error) {
	lockoutKey := fmt.Sprintf("totp:locked:%s:%s", tenantID.String(), userID.String())
	
	var locked bool
	err := r.cacheRepo.Get(ctx, lockoutKey, &locked)
	if err != nil {
		// If key doesn't exist, account is not locked
		return false, time.Time{}, nil
	}

	if locked {
		// Return lockout time (approximate based on lockout duration)
		lockUntil := time.Now().Add(r.lockoutDuration)
		return true, lockUntil, nil
	}

	return false, time.Time{}, nil
}

// ResetTOTPAttempts resets the failed TOTP attempts counter
func (r *TOTPRateLimiter) ResetTOTPAttempts(ctx context.Context, tenantID, userID uuid.UUID) error {
	key := fmt.Sprintf("totp:failed:%s:%s", tenantID.String(), userID.String())
	lockoutKey := fmt.Sprintf("totp:locked:%s:%s", tenantID.String(), userID.String())
	
	if err := r.cacheRepo.Delete(ctx, key, lockoutKey); err != nil {
		return fmt.Errorf("failed to reset TOTP attempts: %w", err)
	}

	return nil
}

// GetRemainingAttempts returns the number of remaining TOTP attempts before lockout
func (r *TOTPRateLimiter) GetRemainingAttempts(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	key := fmt.Sprintf("totp:failed:%s:%s", tenantID.String(), userID.String())
	
	var attempts int64
	err := r.cacheRepo.Get(ctx, key, &attempts)
	if err != nil {
		// If key doesn't exist, all attempts remain
		return r.maxAttempts, nil
	}

	remaining := r.maxAttempts - int(attempts)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}
