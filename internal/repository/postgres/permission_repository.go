package postgres

import (
	"context"
	"encoding/json"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
)

type permissionRepository struct {
	q *db.Queries
}

func NewPermissionRepository(q *db.Queries) repository.PermissionRepository {
	return &permissionRepository{
		q: q,
	}
}

func (r *permissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	// Serialize ABAC rule if present
	var abacRule []byte
	if permission.ABACRule != nil {
		var err error
		abacRule, err = json.Marshal(permission.ABACRule)
		if err != nil {
			return err
		}
	}

	p, err := r.q.CreatePermission(ctx, db.CreatePermissionParams{
		TenantID:    permission.TenantID,
		ResourceID:  permission.ResourceID,
		ScopeID:     permission.ScopeID,
		Name:        permission.Name,
		Description: permission.Description,
		AbacRule:    abacRule,
	})
	if err != nil {
		return err
	}

	permission.ID = p.ID
	permission.CreatedAt = p.CreatedAt.Time
	permission.UpdatedAt = p.UpdatedAt.Time
	if p.IsActive != nil {
		permission.IsActive = *p.IsActive
	} else {
		permission.IsActive = false
	}

	return nil
}

func (r *permissionRepository) GetByID(ctx context.Context, tenantID, permissionID uuid.UUID) (*domain.Permission, error) {
	p, err := r.q.GetPermissionByID(ctx, db.GetPermissionByIDParams{
		ID:       permissionID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	
	// Map row to domain permission
	perm := &domain.Permission{
		ID:          p.ID,
		TenantID:    p.TenantID,
		ResourceID:  p.ResourceID,
		ScopeID:     p.ScopeID,
		Name:        p.Name,
		Description: p.Description,
		IsActive:    p.IsActive != nil && *p.IsActive,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
	
	if len(p.AbacRule) > 0 {
		var rule domain.ABACRule
		if err := json.Unmarshal(p.AbacRule, &rule); err == nil {
			perm.ABACRule = &rule
		}
	}
	
	return perm, nil
}

func (r *permissionRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.Permission, error) {
	p, err := r.q.GetPermissionByName(ctx, db.GetPermissionByNameParams{
		Name:     name,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}
	
	perm := &domain.Permission{
		ID:          p.ID,
		TenantID:    p.TenantID,
		ResourceID:  p.ResourceID,
		ScopeID:     p.ScopeID,
		Name:        p.Name,
		Description: p.Description,
		IsActive:    p.IsActive != nil && *p.IsActive,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
	
	if len(p.AbacRule) > 0 {
		var rule domain.ABACRule
		if err := json.Unmarshal(p.AbacRule, &rule); err == nil {
			perm.ABACRule = &rule
		}
	}
	
	return perm, nil
}

func (r *permissionRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Permission, error) {
	perms, err := r.q.ListPermissions(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Permission, len(perms))
	for i, p := range perms {
		perm := &domain.Permission{
			ID:          p.ID,
			TenantID:    p.TenantID,
			ResourceID:  p.ResourceID,
			ScopeID:     p.ScopeID,
			Name:        p.Name,
			Description: p.Description,
			IsActive:    p.IsActive != nil && *p.IsActive,
			CreatedAt:   p.CreatedAt.Time,
			UpdatedAt:   p.UpdatedAt.Time,
		}
		
		if len(p.AbacRule) > 0 {
			var rule domain.ABACRule
			if err := json.Unmarshal(p.AbacRule, &rule); err == nil {
				perm.ABACRule = &rule
			}
		}
		result[i] = perm
	}
	return result, nil
}

func (r *permissionRepository) GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Permission, error) {
	perms, err := r.q.GetUserPermissions(ctx, db.GetUserPermissionsParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Permission, len(perms))
	for i, p := range perms {
		perm := &domain.Permission{
			ID:          p.ID,
			TenantID:    p.TenantID,
			ResourceID:  p.ResourceID,
			ScopeID:     p.ScopeID,
			Name:        p.Name,
			Description: p.Description,
			// Note: CreatedAt/UpdatedAt/IsActive might not be in the projection or need to be handled if crucial
			// The query returns specific fields, we map what we have
		}
		
		if len(p.AbacRule) > 0 {
			var rule domain.ABACRule
			if err := json.Unmarshal(p.AbacRule, &rule); err == nil {
				perm.ABACRule = &rule
			}
		}
		result[i] = perm
	}
	return result, nil
}

func (r *permissionRepository) CheckUserPermission(ctx context.Context, tenantID, userID uuid.UUID, permissionName string) (bool, error) {
	return r.q.CheckUserPermission(ctx, db.CheckUserPermissionParams{
		UserID:   userID,
		Name:     permissionName,
		TenantID: tenantID,
	})
}

func (r *permissionRepository) CreateResource(ctx context.Context, resource *domain.Resource) error {
	res, err := r.q.CreateResource(ctx, db.CreateResourceParams{
		TenantID:    resource.TenantID,
		Name:        resource.Name,
		Description: resource.Description,
	})
	if err != nil {
		return err
	}

	resource.ID = res.ID
	resource.CreatedAt = res.CreatedAt.Time
	resource.UpdatedAt = res.UpdatedAt.Time
	if res.IsActive != nil {
		resource.IsActive = *res.IsActive
	} else {
		resource.IsActive = false
	}

	return nil
}

func (r *permissionRepository) CreateScope(ctx context.Context, scope *domain.Scope) error {
	s, err := r.q.CreateScope(ctx, db.CreateScopeParams{
		TenantID:    scope.TenantID,
		Name:        scope.Name,
		Description: scope.Description,
	})
	if err != nil {
		return err
	}

	scope.ID = s.ID
	scope.CreatedAt = s.CreatedAt.Time
	scope.UpdatedAt = s.UpdatedAt.Time
	if s.IsActive != nil {
		scope.IsActive = *s.IsActive
	} else {
		scope.IsActive = false
	}

	return nil
}

func (r *permissionRepository) ListResources(ctx context.Context, tenantID uuid.UUID) ([]*domain.Resource, error) {
	resources, err := r.q.ListResources(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Resource, len(resources))
	for i, res := range resources {
		result[i] = &domain.Resource{
			ID:          res.ID,
			TenantID:    res.TenantID,
			Name:        res.Name,
			Description: res.Description,
			IsActive:    res.IsActive != nil && *res.IsActive,
			CreatedAt:   res.CreatedAt.Time,
			UpdatedAt:   res.UpdatedAt.Time,
		}
	}
	return result, nil
}

func (r *permissionRepository) ListScopes(ctx context.Context, tenantID uuid.UUID) ([]*domain.Scope, error) {
	scopes, err := r.q.ListScopes(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Scope, len(scopes))
	for i, s := range scopes {
		result[i] = &domain.Scope{
			ID:          s.ID,
			TenantID:    s.TenantID,
			Name:        s.Name,
			Description: s.Description,
			IsActive:    s.IsActive != nil && *s.IsActive,
			CreatedAt:   s.CreatedAt.Time,
			UpdatedAt:   s.UpdatedAt.Time,
		}
	}
	return result, nil
}
