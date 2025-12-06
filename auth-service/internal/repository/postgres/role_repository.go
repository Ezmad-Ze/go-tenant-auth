package postgres

import (
	"context"
	"time"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
)

type roleRepository struct {
	q *db.Queries
}

func NewRoleRepository(q *db.Queries) repository.RoleRepository {
	return &roleRepository{
		q: q,
	}
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	ro, err := r.q.CreateRole(ctx, db.CreateRoleParams{
		TenantID:    role.TenantID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    &role.IsSystem,
	})
	if err != nil {
		return err
	}

	role.ID = ro.ID
	role.CreatedAt = ro.CreatedAt.Time
	role.UpdatedAt = ro.UpdatedAt.Time
	if ro.IsActive != nil {
		role.IsActive = *ro.IsActive
	} else {
		role.IsActive = false
	}

	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, tenantID, roleID uuid.UUID) (*domain.Role, error) {
	ro, err := r.q.GetRoleByID(ctx, db.GetRoleByIDParams{
		ID:       roleID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainRole(ro), nil
}

func (r *roleRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.Role, error) {
	ro, err := r.q.GetRoleByName(ctx, db.GetRoleByNameParams{
		Name:     name,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainRole(ro), nil
}

func (r *roleRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	roles, err := r.q.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Role, len(roles))
	for i, ro := range roles {
		result[i] = toDomainRole(ro)
	}
	return result, nil
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error {
	_, err := r.q.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: uuidPtrToPgUUID(&assignedBy),
		ExpiresAt:  timePtrToPgTimestamptz(expiresAt),
	})
	return err
}

func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.q.RemoveRoleFromUser(ctx, db.RemoveRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	roles, err := r.q.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Role, len(roles))
	for i, ro := range roles {
		result[i] = toDomainRole(ro)
	}
	return result, nil
}

func (r *roleRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	_, err := r.q.AssignPermissionToRole(ctx, db.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	return err
}

func (r *roleRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.q.RemovePermissionFromRole(ctx, db.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	perms, err := r.q.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Permission, len(perms))
	for i, p := range perms {
		result[i] = toDomainPermission(p)
	}
	return result, nil
}

func (r *roleRepository) AssignPermissionToUser(ctx context.Context, userID, permissionID, assignedBy uuid.UUID, expiresAt *time.Time) error {
	_, err := r.q.AssignPermissionToUser(ctx, db.AssignPermissionToUserParams{
		UserID:       userID,
		PermissionID: permissionID,
		AssignedBy:   uuidPtrToPgUUID(&assignedBy),
		ExpiresAt:    timePtrToPgTimestamptz(expiresAt),
	})
	return err
}

func (r *roleRepository) RemovePermissionFromUser(ctx context.Context, userID, permissionID uuid.UUID) error {
	return r.q.RemovePermissionFromUser(ctx, db.RemovePermissionFromUserParams{
		UserID:       userID,
		PermissionID: permissionID,
	})
}

func toDomainRole(r db.Role) *domain.Role {
	return &domain.Role{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem != nil && *r.IsSystem,
		IsActive:    r.IsActive != nil && *r.IsActive,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}
}

func toDomainPermission(p db.Permission) *domain.Permission {
	return &domain.Permission{
		ID:          p.ID,
		TenantID:    p.TenantID,
		ResourceID:  p.ResourceID,
		ScopeID:     p.ScopeID,
		Name:        p.Name,
		Description: p.Description,
		// ABACRule:    nil, // TODO: Handle JSON unmarshalling if needed
		IsActive:    p.IsActive != nil && *p.IsActive,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}
