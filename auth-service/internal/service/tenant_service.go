package service

import (
	"context"
	"fmt"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/ezmad/auth-service/internal/security"
	"github.com/google/uuid"
)

type tenantService struct {
	tenantRepo     repository.TenantRepository
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	passwordHasher *security.Argon2Params
}

func NewTenantService(
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	passwordHasher *security.Argon2Params,
) TenantService {
	return &tenantService{
		tenantRepo:     tenantRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		passwordHasher: passwordHasher,
	}
}

func (s *tenantService) CreateTenant(ctx context.Context, name, slug, adminEmail, adminPassword string, settings map[string]interface{}) (*TenantCreationResult, error) {
	// Check if slug is unique
	existing, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tenant with slug %s already exists", slug)
	}

	tenant := &domain.Tenant{
		ID:       uuid.New(),
		Name:     name,
		Slug:     slug,
		IsActive: true,
		Settings: settings,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Hash admin password
	passwordHash, err := security.HashPassword(adminPassword, s.passwordHasher)
	if err != nil {
		return nil, fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create super admin user
	adminUser := &domain.User{
		TenantID:     tenant.ID,
		Email:        adminEmail,
		PasswordHash: passwordHash,
		IsActive:     true,
		IsVerified:   true, // Auto-verify super admin
	}

	if err := s.userRepo.Create(ctx, adminUser); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create or get super_admin role
	superAdminRole, err := s.createSuperAdminRole(ctx, tenant.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create super admin role: %w", err)
	}

	// Assign super_admin role to admin user
	if err := s.roleRepo.AssignRoleToUser(ctx, adminUser.ID, superAdminRole.ID, adminUser.ID, nil); err != nil {
		return nil, fmt.Errorf("failed to assign super admin role: %w", err)
	}

	return &TenantCreationResult{
		Tenant: tenant,
		AdminUser: &AdminUserInfo{
			ID:       adminUser.ID,
			Email:    adminUser.Email,
			Username: adminUser.Username,
		},
	}, nil
}

func (s *tenantService) createSuperAdminRole(ctx context.Context, tenantID uuid.UUID) (*domain.Role, error) {
	// Try to get existing super_admin role
	role, err := s.roleRepo.GetByName(ctx, tenantID, "super_admin")
	if err == nil && role != nil {
		return role, nil
	}

	// Create new super_admin role
	description := "Super administrator with full system access"
	role = &domain.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        "super_admin",
		Description: &description,
		IsSystem:    true,
		IsActive:    true,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *tenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant, nil
}

func (s *tenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, update *TenantUpdateRequest) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Update fields if provided
	if update.Name != nil {
		tenant.Name = *update.Name
	}
	if update.Slug != nil {
		// Check if new slug is unique
		existing, err := s.tenantRepo.GetBySlug(ctx, *update.Slug)
		if err == nil && existing != nil && existing.ID != tenantID {
			return nil, fmt.Errorf("tenant with slug %s already exists", *update.Slug)
		}
		tenant.Slug = *update.Slug
	}
	if update.Domain != nil {
		tenant.Domain = update.Domain
	}
	if update.IsActive != nil {
		tenant.IsActive = *update.IsActive
	}
	if update.Settings != nil {
		tenant.Settings = update.Settings
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) ListTenants(ctx context.Context) ([]*domain.Tenant, error) {
	tenants, err := s.tenantRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}
