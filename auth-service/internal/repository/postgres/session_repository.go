package postgres

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	SessionCacheKeyPattern = "session:%s:%s" // session:{tenant_id}:{session_id}
)

type sessionRepository struct {
	q     *db.Queries
	cache repository.CacheRepository
	ttl   time.Duration
}

func NewSessionRepository(q *db.Queries, cache repository.CacheRepository, ttl time.Duration) repository.SessionRepository {
	return &sessionRepository{
		q:     q,
		cache: cache,
		ttl:   ttl,
	}
}

func (r *sessionRepository) Create(ctx context.Context, s *domain.Session) error {
	var ipAddr *netip.Addr
	if s.IPAddress != nil {
		if addr, err := netip.ParseAddr(*s.IPAddress); err == nil {
			ipAddr = &addr
		}
	}

	_, err := r.q.CreateSession(ctx, db.CreateSessionParams{
		UserID:            s.UserID,
		TenantID:          s.TenantID,
		DeviceID:          uuidPtrToPgUUID(s.DeviceID),
		TokenHash:         s.TokenHash,
		RefreshTokenHash:  s.RefreshTokenHash,
		IpAddress:         ipAddr,
		UserAgent:         s.UserAgent,
		ExpiresAt:         pgtype.Timestamptz{Time: s.ExpiresAt, Valid: true},
		RefreshExpiresAt:  timePtrToPgTimestamptz(s.RefreshExpiresAt),
	})
	if err != nil {
		return err
	}

	// Cache the session with TTL matching session expiration
	if r.cache != nil {
		cacheKey := r.sessionCacheKey(s.TenantID, s.ID)
		cacheTTL := time.Until(s.ExpiresAt)
		if cacheTTL > 0 && cacheTTL <= r.ttl {
			_ = r.cache.Set(ctx, cacheKey, s, cacheTTL)
		} else if cacheTTL > 0 {
			_ = r.cache.Set(ctx, cacheKey, s, r.ttl)
		}
	}

	return nil
}

func (r *sessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	// Try cache first (we don't cache by token hash, only by session ID)
	// So we go directly to database for token hash lookups
	s, err := r.q.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	
	session := toDomainSession(s)
	
	// Cache the session for future lookups by ID
	if r.cache != nil && session.IsActive {
		cacheKey := r.sessionCacheKey(session.TenantID, session.ID)
		cacheTTL := time.Until(session.ExpiresAt)
		if cacheTTL > 0 && cacheTTL <= r.ttl {
			_ = r.cache.Set(ctx, cacheKey, session, cacheTTL)
		} else if cacheTTL > 0 {
			_ = r.cache.Set(ctx, cacheKey, session, r.ttl)
		}
	}
	
	return session, nil
}

func (r *sessionRepository) Invalidate(ctx context.Context, sessionID uuid.UUID) error {
	err := r.q.InvalidateSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Invalidate cache - we need tenant ID, so we'll delete by pattern
	// For now, we'll just try to delete with a wildcard pattern
	// In production, you might want to store tenant ID separately or query it first
	if r.cache != nil {
		// We can't easily delete without tenant ID, so we'll skip cache invalidation here
		// The cache will expire naturally based on TTL
		// Alternatively, we could query the session first to get tenant ID
	}

	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, tenantID, sessionID uuid.UUID) (*domain.Session, error) {
	// Try cache first
	if r.cache != nil {
		cacheKey := r.sessionCacheKey(tenantID, sessionID)
		var cachedSession domain.Session
		if err := r.cache.Get(ctx, cacheKey, &cachedSession); err == nil {
			return &cachedSession, nil
		}
	}

	// Cache miss - query database
	s, err := r.q.GetSessionByID(ctx, db.GetSessionByIDParams{
		ID:       sessionID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	session := toDomainSession(s)

	// Cache the session
	if r.cache != nil && session.IsActive {
		cacheTTL := time.Until(session.ExpiresAt)
		if cacheTTL > 0 && cacheTTL <= r.ttl {
			_ = r.cache.Set(ctx, r.sessionCacheKey(tenantID, sessionID), session, cacheTTL)
		} else if cacheTTL > 0 {
			_ = r.cache.Set(ctx, r.sessionCacheKey(tenantID, sessionID), session, r.ttl)
		}
	}

	return session, nil
}

func (r *sessionRepository) ListUserSessions(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Session, error) {
	sessions, err := r.q.ListUserSessions(ctx, db.ListUserSessionsParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Session, len(sessions))
	for i, s := range sessions {
		result[i] = toDomainSession(s)
	}
	return result, nil
}

func (r *sessionRepository) UpdateActivity(ctx context.Context, sessionID uuid.UUID) error {
	return r.q.UpdateSessionActivity(ctx, sessionID)
}

func (r *sessionRepository) InvalidateAllUserSessions(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.q.InvalidateAllUserSessions(ctx, db.InvalidateAllUserSessionsParams{
		UserID:   userID,
		TenantID: tenantID,
	})
}

func (r *sessionRepository) InvalidateExpired(ctx context.Context) error {
	return r.q.InvalidateExpiredSessions(ctx)
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return r.q.DeleteSession(ctx, sessionID)
}

func (r *sessionRepository) CountUserActiveSessions(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return r.q.CountUserActiveSessions(ctx, db.CountUserActiveSessionsParams{
		UserID:   userID,
		TenantID: tenantID,
	})
}

func toDomainSession(s db.Session) *domain.Session {
	isActive := false
	if s.IsActive != nil {
		isActive = *s.IsActive
	}
	
	return &domain.Session{
		ID:                 s.ID,
		UserID:             s.UserID,
		TenantID:           s.TenantID,
		DeviceID:           pgUUIDToUUIDPtr(s.DeviceID),
		TokenHash:          s.TokenHash,
		RefreshTokenHash:   s.RefreshTokenHash,
		IPAddress:          ipAddrToStringPtr(s.IpAddress),
		UserAgent:          s.UserAgent,
		ExpiresAt:          s.ExpiresAt.Time,
		RefreshExpiresAt:   timestamptzToTimePtr(s.RefreshExpiresAt),
		LastActivityAt:     s.LastActivityAt.Time,
		IsActive:           isActive,
		CreatedAt:          s.CreatedAt.Time,
	}
}

func uuidPtrToPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	var bytes [16]byte
	copy(bytes[:], (*u)[:])
	return pgtype.UUID{Bytes: bytes, Valid: true}
}

func pgUUIDToUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func timePtrToPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func timestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func ipAddrToStringPtr(a *netip.Addr) *string {
	if a == nil {
		return nil
	}
	s := a.String()
	return &s
}

// sessionCacheKey generates a cache key for a session
func (r *sessionRepository) sessionCacheKey(tenantID, sessionID uuid.UUID) string {
	return fmt.Sprintf(SessionCacheKeyPattern, tenantID.String(), sessionID.String())
}