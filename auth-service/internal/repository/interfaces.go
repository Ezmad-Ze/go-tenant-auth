package repository

import (
	"context"
	"time"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, tenantID uuid.UUID, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, email, passwordHash string) error
	VerifyEmail(ctx context.Context, tenantID, userID uuid.UUID) error
	IncrementFailedLoginAttempts(ctx context.Context, tenantID, userID uuid.UUID) error
	ResetFailedLoginAttempts(ctx context.Context, tenantID, userID uuid.UUID) error
	LockUser(ctx context.Context, tenantID, userID uuid.UUID, until *time.Time) error
	UpdateLastLogin(ctx context.Context, tenantID, userID uuid.UUID) error
	Deactivate(ctx context.Context, tenantID, userID uuid.UUID) error
	Delete(ctx context.Context, tenantID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.User, error)
}

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	GetByID(ctx context.Context, tenantID, sessionID uuid.UUID) (*domain.Session, error)
	ListUserSessions(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Session, error)
	UpdateActivity(ctx context.Context, sessionID uuid.UUID) error
	Invalidate(ctx context.Context, sessionID uuid.UUID) error
	InvalidateAllUserSessions(ctx context.Context, tenantID, userID uuid.UUID) error
	InvalidateExpired(ctx context.Context) error
	Delete(ctx context.Context, sessionID uuid.UUID) error
	CountUserActiveSessions(ctx context.Context, tenantID, userID uuid.UUID) (int64, error)
}

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	Create(ctx context.Context, permission *domain.Permission) error
	GetByID(ctx context.Context, tenantID, permissionID uuid.UUID) (*domain.Permission, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.Permission, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Permission, error)
	GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Permission, error)
	CheckUserPermission(ctx context.Context, tenantID, userID uuid.UUID, permissionName string) (bool, error)
	CreateResource(ctx context.Context, resource *domain.Resource) error
	CreateScope(ctx context.Context, scope *domain.Scope) error
	ListResources(ctx context.Context, tenantID uuid.UUID) ([]*domain.Resource, error)
	ListScopes(ctx context.Context, tenantID uuid.UUID) ([]*domain.Scope, error)
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, tenantID, roleID uuid.UUID) (*domain.Role, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.Role, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error)
	AssignPermissionToUser(ctx context.Context, userID, permissionID, assignedBy uuid.UUID, expiresAt *time.Time) error
	RemovePermissionFromUser(ctx context.Context, userID, permissionID uuid.UUID) error
}

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	List(ctx context.Context) ([]*domain.Tenant, error)
}

// DeviceRepository defines the interface for device data operations
type DeviceRepository interface {
	Create(ctx context.Context, device *domain.Device) error
	GetByID(ctx context.Context, tenantID, deviceID uuid.UUID) (*domain.Device, error)
	GetByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error)
	ListUserDevices(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Device, error)
	UpdateLastUsed(ctx context.Context, deviceID uuid.UUID) error
	TrustDevice(ctx context.Context, deviceID uuid.UUID) error
	Delete(ctx context.Context, deviceID, userID uuid.UUID) error
}

// MagicLinkRepository defines the interface for magic link data operations
type MagicLinkRepository interface {
	Create(ctx context.Context, link *domain.MagicLink) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.MagicLink, error)
	MarkUsed(ctx context.Context, linkID uuid.UUID) error
}

// TOTPRepository defines the interface for TOTP 2FA data operations
type TOTPRepository interface {
	Create(ctx context.Context, totp *domain.TOTP2FA) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TOTP2FA, error)
	Enable(ctx context.Context, userID uuid.UUID) error
	Disable(ctx context.Context, userID uuid.UUID) error
	UpdateBackupCodes(ctx context.Context, userID uuid.UUID, codes []string) error
}

// CacheRepository defines the interface for Redis cache operations
type CacheRepository interface {
	// Basic operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// Atomic operations
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
}
