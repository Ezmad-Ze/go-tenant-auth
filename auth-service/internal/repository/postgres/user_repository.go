package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	UserCacheKeyPattern      = "user:%s:%s"       // user:{tenant_id}:{user_id}
	UserEmailCacheKeyPattern = "user:email:%s:%s" // user:email:{tenant_id}:{email}
)

type userRepository struct {
	q     *db.Queries
	cache repository.CacheRepository
	ttl   time.Duration
}

func NewUserRepository(q *db.Queries, cache repository.CacheRepository, ttl time.Duration) repository.UserRepository {
	return &userRepository{
		q:     q,
		cache: cache,
		ttl:   ttl,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	createdUser, err := r.q.CreateUser(ctx, db.CreateUserParams{
		TenantID:     user.TenantID,
		Email:        user.Email,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Phone:        user.Phone,
	})
	if err != nil {
		return err
	}
	
	// Update the user object with the database-generated ID and timestamps
	user.ID = createdUser.ID
	user.CreatedAt = createdUser.CreatedAt.Time
	user.UpdatedAt = createdUser.UpdatedAt.Time
	
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	// Try cache first
	if r.cache != nil {
		cacheKey := r.userCacheKey(tenantID, userID)
		var cachedUser domain.User
		if err := r.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
			return &cachedUser, nil
		}
	}

	// Cache miss - query database
	u, err := r.q.GetUserByID(ctx, db.GetUserByIDParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	user := toDomainUser(u)

	// Cache the user
	if r.cache != nil {
		_ = r.cache.Set(ctx, r.userCacheKey(tenantID, userID), user, r.ttl)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	// Try cache first
	if r.cache != nil {
		cacheKey := r.userEmailCacheKey(tenantID, email)
		var cachedUser domain.User
		if err := r.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
			return &cachedUser, nil
		}
	}

	// Cache miss - query database
	u, err := r.q.GetUserByEmail(ctx, db.GetUserByEmailParams{
		Email:    email,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	user := toDomainUser(u)

	// Cache the user
	if r.cache != nil {
		_ = r.cache.Set(ctx, r.userEmailCacheKey(tenantID, email), user, r.ttl)
		_ = r.cache.Set(ctx, r.userCacheKey(tenantID, user.ID), user, r.ttl)
	}

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, tenantID uuid.UUID, username string) (*domain.User, error) {
	u, err := r.q.GetUserByUsername(ctx, db.GetUserByUsernameParams{
		Username: &username,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainUser(u), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		ID:        user.ID,
		TenantID:  user.TenantID,
	})
	if err != nil {
		return err
	}
	
	// Invalidate cache
	r.invalidateUserCache(ctx, user.TenantID, user.ID, user.Email)
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, email, passwordHash string) error {
	err := r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		PasswordHash: passwordHash,
		ID:           userID,
		TenantID:     tenantID,
	})
	if err != nil {
		return err
	}
	
	// Invalidate cache
	r.invalidateUserCache(ctx, tenantID, userID, email)
	return nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, tenantID, userID uuid.UUID) error {
	err := r.q.VerifyUserEmail(ctx, db.VerifyUserEmailParams{
		ID:       userID,
		TenantID: tenantID,
	})
	if err != nil {
		return err
	}
	
	// Invalidate cache
	r.invalidateUserCache(ctx, tenantID, userID, "")
	return nil
}

func (r *userRepository) IncrementFailedLoginAttempts(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.q.IncrementFailedLoginAttempts(ctx, db.IncrementFailedLoginAttemptsParams{
		ID:       userID,
		TenantID: tenantID,
	})
}

func (r *userRepository) ResetFailedLoginAttempts(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.q.ResetFailedLoginAttempts(ctx, db.ResetFailedLoginAttemptsParams{
		ID:       userID,
		TenantID: tenantID,
	})
}

func (r *userRepository) LockUser(ctx context.Context, tenantID, userID uuid.UUID, until *time.Time) error {
	var lockedUntil pgtype.Timestamptz
	if until != nil {
		lockedUntil = pgtype.Timestamptz{Time: *until, Valid: true}
	} else {
		lockedUntil.Valid = false
	}
	return r.q.LockUser(ctx, db.LockUserParams{
		LockedUntil: lockedUntil,
		ID:          userID,
		TenantID:    tenantID,
	})
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.q.UpdateLastLogin(ctx, db.UpdateLastLoginParams{
		ID:       userID,
		TenantID: tenantID,
	})
}

func (r *userRepository) Deactivate(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.q.DeactivateUser(ctx, db.DeactivateUserParams{
		ID:       userID,
		TenantID: tenantID,
	})
}

func (r *userRepository) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.q.DeleteUser(ctx, db.DeleteUserParams{
		ID:       userID,
		TenantID: tenantID,
	})
}

func (r *userRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.User, error) {
	users, err := r.q.ListUsers(ctx, db.ListUsersParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	domainUsers := make([]*domain.User, len(users))
	for i, u := range users {
		domainUsers[i] = toDomainUser(u)
	}
	return domainUsers, nil
}

func toDomainUser(u db.User) *domain.User {
	isActive := false
	if u.IsActive != nil {
		isActive = *u.IsActive
	}
	
	isVerified := false
	if u.IsVerified != nil {
		isVerified = *u.IsVerified
	}
	
	failedAttempts := 0
	if u.FailedLoginAttempts != nil {
		failedAttempts = int(*u.FailedLoginAttempts)
	}
	
	return &domain.User{
		ID:                  u.ID,
		TenantID:            u.TenantID,
		Email:               u.Email,
		Username:            u.Username,
		PasswordHash:        u.PasswordHash,
		FirstName:           u.FirstName,
		LastName:            u.LastName,
		Phone:               u.Phone,
		IsActive:            isActive,
		IsVerified:          isVerified,
		EmailVerifiedAt:     ptrTime(u.EmailVerifiedAt),
		FailedLoginAttempts: failedAttempts,
		LockedUntil:         ptrTime(u.LockedUntil),
		LastLoginAt:         ptrTime(u.LastLoginAt),
		PasswordChangedAt:   u.PasswordChangedAt.Time,
		Metadata:            nil, // decode JSON if needed later
		CreatedAt:           u.CreatedAt.Time,
		UpdatedAt:           u.UpdatedAt.Time,
	}
}

func ptrTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// userCacheKey generates a cache key for a user by ID
func (r *userRepository) userCacheKey(tenantID, userID uuid.UUID) string {
	return fmt.Sprintf(UserCacheKeyPattern, tenantID.String(), userID.String())
}

// userEmailCacheKey generates a cache key for a user by email
func (r *userRepository) userEmailCacheKey(tenantID uuid.UUID, email string) string {
	return fmt.Sprintf(UserEmailCacheKeyPattern, tenantID.String(), email)
}

// invalidateUserCache removes all cache entries for a user
func (r *userRepository) invalidateUserCache(ctx context.Context, tenantID, userID uuid.UUID, email string) {
	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.userCacheKey(tenantID, userID))
		if email != "" {
			_ = r.cache.Delete(ctx, r.userEmailCacheKey(tenantID, email))
		}
	}
}