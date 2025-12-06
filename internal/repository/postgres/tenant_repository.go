package postgres

import (
	"context"
	"encoding/json"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type tenantRepository struct {
	q *db.Queries
}

func NewTenantRepository(q *db.Queries) repository.TenantRepository {
	return &tenantRepository{
		q: q,
	}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	// Serialize settings
	var settings []byte
	if tenant.Settings != nil {
		var err error
		settings, err = json.Marshal(tenant.Settings)
		if err != nil {
			return err
		}
	}

	t, err := r.q.CreateTenant(ctx, db.CreateTenantParams{
		Name:     tenant.Name,
		Slug:     tenant.Slug,
		Domain:   tenant.Domain,
		Settings: settings,
	})
	if err != nil {
		return err
	}
	
	// Update ID and timestamps
	tenant.ID = t.ID
	tenant.CreatedAt = t.CreatedAt.Time
	tenant.UpdatedAt = t.UpdatedAt.Time
	if t.IsActive != nil {
		tenant.IsActive = *t.IsActive
	} else {
		tenant.IsActive = false
	}
	
	return nil
}

func (r *tenantRepository) GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	t, err := r.q.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return toDomainTenant(t), nil
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	t, err := r.q.GetTenantBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return toDomainTenant(t), nil
}

func (r *tenantRepository) GetByDomain(ctx context.Context, domain string) (*domain.Tenant, error) {
	t, err := r.q.GetTenantByDomain(ctx, &domain)
	if err != nil {
		return nil, err
	}
	return toDomainTenant(t), nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	// Serialize settings
	var settings []byte
	if tenant.Settings != nil {
		var err error
		settings, err = json.Marshal(tenant.Settings)
		if err != nil {
			return err
		}
	}

	t, err := r.q.UpdateTenant(ctx, db.UpdateTenantParams{
		Name:     &tenant.Name,
		Domain:   tenant.Domain,
		Settings: settings,
		ID:       tenant.ID,
	})
	if err != nil {
		return err
	}
	
	tenant.UpdatedAt = t.UpdatedAt.Time
	return nil
}

func (r *tenantRepository) List(ctx context.Context) ([]*domain.Tenant, error) {
	tenants, err := r.q.ListTenants(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Tenant, len(tenants))
	for i, t := range tenants {
		result[i] = toDomainTenant(t)
	}
	return result, nil
}

func toDomainTenant(t db.Tenant) *domain.Tenant {
	tenant := &domain.Tenant{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		Domain:    t.Domain,
		IsActive:  t.IsActive != nil && *t.IsActive,
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
	}
	
	if len(t.Settings) > 0 {
		var settings map[string]interface{}
		if err := json.Unmarshal(t.Settings, &settings); err == nil {
			tenant.Settings = settings
		}
	}
	
	return tenant
}

func stringPtrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
