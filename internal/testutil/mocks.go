package testutil

import (
	"context"
	"time"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	CreateFunc                     func(ctx context.Context, user *domain.User) error
	GetByIDFunc                    func(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error)
	GetByEmailFunc                 func(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error)
	UpdatePasswordFunc             func(ctx context.Context, tenantID, userID uuid.UUID, email, passwordHash string) error
	VerifyEmailFunc                func(ctx context.Context, tenantID, userID uuid.UUID) error
	IncrementFailedLoginAttemptsFunc func(ctx context.Context, tenantID, userID uuid.UUID) error
	ResetFailedLoginAttemptsFunc   func(ctx context.Context, tenantID, userID uuid.UUID) error
	UpdateLastLoginFunc            func(ctx context.Context, tenantID, userID uuid.UUID) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, tenantID, userID)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, tenantID, email)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, tenantID uuid.UUID, username string) (*domain.User, error) {
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, email, passwordHash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, tenantID, userID, email, passwordHash)
	}
	return nil
}

func (m *MockUserRepository) VerifyEmail(ctx context.Context, tenantID, userID uuid.UUID) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(ctx, tenantID, userID)
	}
	return nil
}

func (m *MockUserRepository) IncrementFailedLoginAttempts(ctx context.Context, tenantID, userID uuid.UUID) error {
	if m.IncrementFailedLoginAttemptsFunc != nil {
		return m.IncrementFailedLoginAttemptsFunc(ctx, tenantID, userID)
	}
	return nil
}

func (m *MockUserRepository) ResetFailedLoginAttempts(ctx context.Context, tenantID, userID uuid.UUID) error {
	if m.ResetFailedLoginAttemptsFunc != nil {
		return m.ResetFailedLoginAttemptsFunc(ctx, tenantID, userID)
	}
	return nil
}

func (m *MockUserRepository) LockUser(ctx context.Context, tenantID, userID uuid.UUID, until *time.Time) error {
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, tenantID, userID uuid.UUID) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(ctx, tenantID, userID)
	}
	return nil
}

func (m *MockUserRepository) Deactivate(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	CreateFunc func(ctx context.Context, session *domain.Session) error
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, session)
	}
	return nil
}

func (m *MockSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	return nil, nil
}

func (m *MockSessionRepository) GetByID(ctx context.Context, tenantID, sessionID uuid.UUID) (*domain.Session, error) {
	return nil, nil
}

func (m *MockSessionRepository) ListUserSessions(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Session, error) {
	return nil, nil
}

func (m *MockSessionRepository) UpdateActivity(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (m *MockSessionRepository) Invalidate(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (m *MockSessionRepository) InvalidateAllUserSessions(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}

func (m *MockSessionRepository) InvalidateExpired(ctx context.Context) error {
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (m *MockSessionRepository) CountUserActiveSessions(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return 0, nil
}

// MockTOTPRepository is a mock implementation of TOTPRepository
type MockTOTPRepository struct {
	GetByUserIDFunc      func(ctx context.Context, userID uuid.UUID) (*domain.TOTP2FA, error)
	CreateFunc           func(ctx context.Context, totp *domain.TOTP2FA) error
	EnableFunc           func(ctx context.Context, userID uuid.UUID) error
	UpdateBackupCodesFunc func(ctx context.Context, userID uuid.UUID, codes []string) error
}

func (m *MockTOTPRepository) Create(ctx context.Context, totp *domain.TOTP2FA) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, totp)
	}
	return nil
}

func (m *MockTOTPRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TOTP2FA, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockTOTPRepository) Enable(ctx context.Context, userID uuid.UUID) error {
	if m.EnableFunc != nil {
		return m.EnableFunc(ctx, userID)
	}
	return nil
}

func (m *MockTOTPRepository) Disable(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *MockTOTPRepository) UpdateBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error {
	if m.UpdateBackupCodesFunc != nil {
		return m.UpdateBackupCodesFunc(ctx, userID, codes)
	}
	return nil
}

// Ensure mocks implement interfaces
var (
	_ repository.UserRepository    = (*MockUserRepository)(nil)
	_ repository.SessionRepository = (*MockSessionRepository)(nil)
	_ repository.TOTPRepository    = (*MockTOTPRepository)(nil)
)
