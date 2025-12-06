package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
)

// authorizationService implements the AuthorizationService interface
type authorizationService struct {
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
	userRepo       repository.UserRepository
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
) AuthorizationService {
	return &authorizationService{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		userRepo:       userRepo,
	}
}

// CheckPermission implements RBAC + ABAC permission checking
func (s *authorizationService) CheckPermission(ctx context.Context, check *domain.PermissionCheck) (bool, error) {
	// First check RBAC - does user have the permission through roles or direct assignment?
	permissionName := fmt.Sprintf("%s.%s", check.Resource, check.Scope)
	hasPermission, err := s.permissionRepo.CheckUserPermission(ctx, check.TenantID, check.UserID, permissionName)
	if err != nil {
		return false, err
	}

	if !hasPermission {
		return false, nil
	}

	// If RBAC passed, check ABAC rules if any
	permission, err := s.permissionRepo.GetByName(ctx, check.TenantID, permissionName)
	if err != nil {
		return false, err
	}

	// If no ABAC rule, RBAC permission is sufficient
	if permission.ABACRule == nil || len(permission.ABACRule.Conditions) == 0 {
		return true, nil
	}

	// Evaluate ABAC conditions
	return s.evaluateABACRule(permission.ABACRule, check.Attributes), nil
}

// evaluateABACRule evaluates ABAC conditions
// Example: user can update project only if project.owner_id == user.id
func (s *authorizationService) evaluateABACRule(rule *domain.ABACRule, attributes map[string]interface{}) bool {
	for _, condition := range rule.Conditions {
		// Get the attribute value from the context
		attrValue, exists := attributes[condition.Attribute]
		if !exists {
			return false // Required attribute not provided
		}

		// Evaluate based on operator
		switch condition.Operator {
		case "equals":
			if attrValue != condition.Value {
				return false
			}
		case "not_equals":
			if attrValue == condition.Value {
				return false
			}
		case "in":
			// Check if attrValue is in a slice of values
			valueSlice, ok := condition.Value.([]interface{})
			if !ok {
				return false
			}
			found := false
			for _, v := range valueSlice {
				if attrValue == v {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		default:
			return false // Unknown operator
		}
	}

	// All conditions passed
	return true
}

// GetUserPermissions returns all permissions for a user
func (s *authorizationService) GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Permission, error) {
	return s.permissionRepo.GetUserPermissions(ctx, tenantID, userID)
}

// CreateRole creates a new role
func (s *authorizationService) CreateRole(ctx context.Context, tenantID uuid.UUID, name, description string, isSystem bool) (*domain.Role, error) {
	role := &domain.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: &description,
		IsSystem:    isSystem,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// AssignRoleToUser assigns a role to a user
func (s *authorizationService) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error {
	return s.roleRepo.AssignRoleToUser(ctx, userID, roleID, assignedBy, expiresAt)
}

// RemoveRoleFromUser removes a role from a user
func (s *authorizationService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.roleRepo.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetUserRoles returns all roles for a user
func (s *authorizationService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	return s.roleRepo.GetUserRoles(ctx, userID)
}

// CreatePermission creates a new permission
func (s *authorizationService) CreatePermission(ctx context.Context, tenantID uuid.UUID, resourceName, scopeName, description string, abacRule *domain.ABACRule) (*domain.Permission, error) {
	// Get or create resource
	resources, err := s.permissionRepo.ListResources(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var resource *domain.Resource
	for _, r := range resources {
		if r.Name == resourceName {
			resource = r
			break
		}
	}
	if resource == nil {
		resource = &domain.Resource{
			ID:       uuid.New(),
			TenantID: tenantID,
			Name:     resourceName,
			IsActive: true,
		}
		if err := s.permissionRepo.CreateResource(ctx, resource); err != nil {
			return nil, err
		}
	}

	// Get or create scope
	scopes, err := s.permissionRepo.ListScopes(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var scope *domain.Scope
	for _, sc := range scopes {
		if sc.Name == scopeName {
			scope = sc
			break
		}
	}
	if scope == nil {
		scope = &domain.Scope{
			ID:       uuid.New(),
			TenantID: tenantID,
			Name:     scopeName,
			IsActive: true,
		}
		if err := s.permissionRepo.CreateScope(ctx, scope); err != nil {
			return nil, err
		}
	}

	// Create permission
	permission := &domain.Permission{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ResourceID:  resource.ID,
		ScopeID:     scope.ID,
		Name:        fmt.Sprintf("%s.%s", resourceName, scopeName),
		Description: &description,
		ABACRule:    abacRule,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, err
	}

	return permission, nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *authorizationService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.roleRepo.AssignPermissionToRole(ctx, roleID, permissionID)
}

// RemovePermissionFromRole removes a permission from a role
func (s *authorizationService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.roleRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

// AssignPermissionToUser assigns a permission directly to a user
func (s *authorizationService) AssignPermissionToUser(ctx context.Context, userID, permissionID, assignedBy uuid.UUID, expiresAt *time.Time) error {
	return s.roleRepo.AssignPermissionToUser(ctx, userID, permissionID, assignedBy, expiresAt)
}

// RemovePermissionFromUser removes a direct permission from a user
func (s *authorizationService) RemovePermissionFromUser(ctx context.Context, userID, permissionID uuid.UUID) error {
	return s.roleRepo.RemovePermissionFromUser(ctx, userID, permissionID)
}

// CreateResource creates a new resource
func (s *authorizationService) CreateResource(ctx context.Context, tenantID uuid.UUID, name, description string) (*domain.Resource, error) {
	resource := &domain.Resource{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: &description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.permissionRepo.CreateResource(ctx, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

// CreateScope creates a new scope
func (s *authorizationService) CreateScope(ctx context.Context, tenantID uuid.UUID, name, description string) (*domain.Scope, error) {
	scope := &domain.Scope{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: &description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.permissionRepo.CreateScope(ctx, scope); err != nil {
		return nil, err
	}
	return scope, nil
}

// ListResources lists all resources for a tenant
func (s *authorizationService) ListResources(ctx context.Context, tenantID uuid.UUID) ([]*domain.Resource, error) {
	return s.permissionRepo.ListResources(ctx, tenantID)
}

// ListScopes lists all scopes for a tenant
func (s *authorizationService) ListScopes(ctx context.Context, tenantID uuid.UUID) ([]*domain.Scope, error) {
	return s.permissionRepo.ListScopes(ctx, tenantID)
}
