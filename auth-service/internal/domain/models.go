package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	Email               string
	Username            *string
	PasswordHash        string
	FirstName           *string
	LastName            *string
	Phone               *string
	IsActive            bool
	IsVerified          bool
	EmailVerifiedAt     *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	PasswordChangedAt   time.Time
	Metadata            map[string]interface{}
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Session represents an active user session
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	TenantID         uuid.UUID
	DeviceID         *uuid.UUID
	TokenHash        string
	RefreshTokenHash *string
	IPAddress        *string
	UserAgent        *string
	ExpiresAt        time.Time
	RefreshExpiresAt *time.Time
	LastActivityAt   time.Time
	IsActive         bool
	CreatedAt        time.Time
}

// Device represents a user device
type Device struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TenantID    uuid.UUID
	DeviceName  *string
	DeviceType  *string
	Fingerprint string
	UserAgent   *string
	IPAddress   *string
	IsTrusted   bool
	LastUsedAt  time.Time
	CreatedAt   time.Time
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a permission (resource + scope)
type Permission struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	ResourceID  uuid.UUID `json:"resource_id"`
	ScopeID     uuid.UUID `json:"scope_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ABACRule    *ABACRule `json:"abac_rule,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Resource represents an entity (project, invoice, etc.)
type Resource struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Scope represents an action (create, read, update, delete)
type Scope struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ABACRule represents an attribute-based access control rule
type ABACRule struct {
	Conditions []ABACCondition `json:"conditions"`
}

// ABACCondition represents a single ABAC condition
type ABACCondition struct {
	Attribute string      `json:"attribute"` // e.g., "project.owner_id"
	Operator  string      `json:"operator"`  // e.g., "equals", "not_equals", "in"
	Value     interface{} `json:"value"`     // e.g., "user.id"
}

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Domain    *string                `json:"domain,omitempty"`
	IsActive  bool                   `json:"is_active"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// TOTP2FA represents two-factor authentication settings
type TOTP2FA struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TenantID    uuid.UUID
	Secret      string
	BackupCodes []string
	IsEnabled   bool
	VerifiedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MagicLink represents a passwordless login link
type MagicLink struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	TenantID  uuid.UUID
	Email     string
	TokenHash string
	IPAddress *string
	UserAgent *string
	ExpiresAt time.Time
	UsedAt    *time.Time
	IsUsed    bool
	CreatedAt time.Time
}

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	UserID       uuid.UUID `json:"user_id"`
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	UserID     uuid.UUID
	TenantID   uuid.UUID
	Resource   string
	Scope      string
	Attributes map[string]interface{} // For ABAC contextual checks
}
