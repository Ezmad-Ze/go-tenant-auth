package service

import (
	"context"
	"time"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/google/uuid"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Registration and login
	Register(ctx context.Context, req *RegisterRequest) (*domain.TokenPair, error)
	Login(ctx context.Context, req *LoginRequest) (*domain.TokenPair, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, tenantID, userID uuid.UUID) error

	// Token operations
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	VerifyAccessToken(ctx context.Context, accessToken string) (*domain.User, error)

	// Magic link authentication
	SendMagicLink(ctx context.Context, tenantID uuid.UUID, email string) error
	VerifyMagicLink(ctx context.Context, token string) (*domain.TokenPair, error)

	// Password operations
	ChangePassword(ctx context.Context, tenantID, userID uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, tenantID uuid.UUID, email string) error
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error

	// Email verification
	SendVerificationEmail(ctx context.Context, tenantID, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
}

// AuthorizationService defines the interface for authorization operations
type AuthorizationService interface {
	// Permission checking (RBAC + ABAC)
	CheckPermission(ctx context.Context, check *domain.PermissionCheck) (bool, error)
	GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Permission, error)

	// Role management
	CreateRole(ctx context.Context, tenantID uuid.UUID, name, description string, isSystem bool) (*domain.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)

	// Permission management
	CreatePermission(ctx context.Context, tenantID uuid.UUID, resourceName, scopeName, description string, abacRule *domain.ABACRule) (*domain.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	AssignPermissionToUser(ctx context.Context, userID, permissionID, assignedBy uuid.UUID, expiresAt *time.Time) error
	RemovePermissionFromUser(ctx context.Context, userID, permissionID uuid.UUID) error

	// Resource and Scope management
	CreateResource(ctx context.Context, tenantID uuid.UUID, name, description string) (*domain.Resource, error)
	CreateScope(ctx context.Context, tenantID uuid.UUID, name, description string) (*domain.Scope, error)
	ListResources(ctx context.Context, tenantID uuid.UUID) ([]*domain.Resource, error)
	ListScopes(ctx context.Context, tenantID uuid.UUID) ([]*domain.Scope, error)
}

// TOTPService defines the interface for 2FA/TOTP operations
type TOTPService interface {
	Setup(ctx context.Context, tenantID, userID uuid.UUID) (*TOTPSetupResponse, error)
	Verify(ctx context.Context, tenantID, userID uuid.UUID, code string) error
	Enable(ctx context.Context, tenantID, userID uuid.UUID, code string) error
	Disable(ctx context.Context, tenantID, userID uuid.UUID, password string) error
	GenerateBackupCodes(ctx context.Context, tenantID, userID uuid.UUID) ([]string, error)
	ValidateBackupCode(ctx context.Context, tenantID, userID uuid.UUID, code string) error
}

// DeviceService defines the interface for device management
type DeviceService interface {
	RegisterDevice(ctx context.Context, tenantID, userID uuid.UUID, req *DeviceRequest) (*domain.Device, error)
	ListDevices(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Device, error)
	TrustDevice(ctx context.Context, tenantID, userID, deviceID uuid.UUID) error
	RevokeDevice(ctx context.Context, tenantID, userID, deviceID uuid.UUID) error
}

// EmailService defines the interface for email operations
type EmailService interface {
	SendMagicLink(ctx context.Context, toEmail, magicLink string) error
	SendVerificationEmail(ctx context.Context, toEmail, token string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, token string) error
}

// TenantService defines the interface for tenant management
type TenantService interface {
	CreateTenant(ctx context.Context, name, slug, adminEmail, adminPassword string, settings map[string]interface{}) (*TenantCreationResult, error)
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, update *TenantUpdateRequest) (*domain.Tenant, error)
	ListTenants(ctx context.Context) ([]*domain.Tenant, error)
}

// Request and response types
type RegisterRequest struct {
	TenantID  uuid.UUID
	Email     string
	Password  string
	Username  *string
	FirstName *string
	LastName  *string
	Phone     *string
	Metadata  map[string]interface{}
}

type LoginRequest struct {
	TenantID   uuid.UUID
	Email      string
	Password   string
	DeviceInfo *DeviceInfo
	TOTP       *string // Optional TOTP code if 2FA enabled
	BackupCode *string // Optional backup code if 2FA enabled
}

type DeviceInfo struct {
	Fingerprint string
	Name        *string
	Type        *string
	UserAgent   *string
	IPAddress   *string
}

type DeviceRequest struct {
	DeviceName  *string
	DeviceType  *string
	Fingerprint string
	UserAgent   *string
	IPAddress   *string
}

type TOTPSetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

type TenantUpdateRequest struct {
	Name     *string
	Slug     *string
	Domain   *string
	IsActive *bool
	Settings map[string]interface{}
}

type TenantCreationResult struct {
	Tenant    *domain.Tenant `json:"tenant"`
	AdminUser *AdminUserInfo `json:"admin_user"`
}

type AdminUserInfo struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Username *string   `json:"username,omitempty"`
}
