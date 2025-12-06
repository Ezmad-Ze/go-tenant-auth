package handlers

// LoginRequest represents the login request body
type LoginRequest struct {
	Email      string  `json:"email" example:"user@example.com"`
	Password   string  `json:"password" example:"secret123"`
	TOTP       *string `json:"totp,omitempty" example:"123456"`
	BackupCode *string `json:"backup_code,omitempty" example:"12345678"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email     string                 `json:"email" example:"user@example.com"`
	Password  string                 `json:"password" example:"secret123"`
	Username  *string                `json:"username,omitempty" example:"jdoe"`
	FirstName *string                `json:"first_name,omitempty" example:"John"`
	LastName  *string                `json:"last_name,omitempty" example:"Doe"`
	Phone     *string                `json:"phone,omitempty" example:"+1234567890"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type" example:"Bearer"`
	UserID       string `json:"user_id"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"refresh_token_string"`
}

// EmailRequest represents a request with just an email
type EmailRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

// TokenRequest represents a request with just a token
type TokenRequest struct {
	Token string `json:"token" example:"verification_token_string"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" example:"old_secret"`
	NewPassword string `json:"new_password" example:"new_secret"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid credentials"`
}

// TOTPVerifyRequest represents the TOTP verification request
type TOTPVerifyRequest struct {
	Code string `json:"code" example:"123456"`
}

// TOTPDisableRequest represents the TOTP disable request
type TOTPDisableRequest struct {
	Password string `json:"password" example:"secret123"`
}

// TOTPSetupResponse represents the TOTP setup response
type TOTPSetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// BackupCodesResponse represents the backup codes response
type BackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// CheckPermissionRequest represents the permission check request
type CheckPermissionRequest struct {
	UserID     string                 `json:"user_id" example:"uuid-string"`
	Resource   string                 `json:"resource" example:"document"`
	Scope      string                 `json:"scope" example:"read"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// CheckPermissionResponse represents the permission check response
type CheckPermissionResponse struct {
	HasPermission bool   `json:"has_permission"`
	Permission    string `json:"permission"`
}

// UserPermissionsResponse represents the user permissions response
type UserPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}

// CreateRoleRequest represents the create role request
type CreateRoleRequest struct {
	Name        string `json:"name" example:"editor"`
	Description string `json:"description" example:"Editor role"`
	IsSystem    bool   `json:"is_system"`
}

// AssignRoleRequest represents the assign role request
type AssignRoleRequest struct {
	UserID string `json:"user_id" example:"uuid-string"`
	RoleID string `json:"role_id" example:"uuid-string"`
}

// RevokeRoleRequest represents the revoke role request
type RevokeRoleRequest struct {
	UserID string `json:"user_id" example:"uuid-string"`
	RoleID string `json:"role_id" example:"uuid-string"`
}

// UserRolesResponse represents the user roles response
type UserRolesResponse struct {
	Roles []interface{} `json:"roles"`
}

// CreatePermissionRequest represents the create permission request
type CreatePermissionRequest struct {
	ResourceName string `json:"resource_name" example:"document"`
	ScopeName    string `json:"scope_name" example:"read"`
	Description  string `json:"description" example:"Read documents"`
}

// AssignPermissionRoleRequest represents the assign permission to role request
type AssignPermissionRoleRequest struct {
	RoleID       string `json:"role_id" example:"uuid-string"`
	PermissionID string `json:"permission_id" example:"uuid-string"`
}

// AssignPermissionUserRequest represents the assign permission to user request
type AssignPermissionUserRequest struct {
	UserID       string `json:"user_id" example:"uuid-string"`
	PermissionID string `json:"permission_id" example:"uuid-string"`
}

// ListResourcesResponse represents the list resources response
type ListResourcesResponse struct {
	Resources []string `json:"resources"`
}

// ListScopesResponse represents the list scopes response
type ListScopesResponse struct {
	Scopes []string `json:"scopes"`
}

// CreateTenantRequest represents the create tenant request
type CreateTenantRequest struct {
	Name          string                 `json:"name" example:"Acme Corp"`
	Slug          string                 `json:"slug" example:"acme"`
	Domain        *string                `json:"domain,omitempty" example:"acme.com"`
	Settings      map[string]interface{} `json:"settings,omitempty"`
	AdminEmail    string                 `json:"admin_email" example:"admin@acme.com"`
	AdminPassword string                 `json:"admin_password" example:"secret123"`
}

// TenantListResponse represents the tenant list response
type TenantListResponse struct {
	Tenants []interface{} `json:"tenants"`
}

// TenantResponse represents the tenant response
type TenantResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Domain    *string                `json:"domain,omitempty"`
	IsActive  bool                   `json:"is_active"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}
