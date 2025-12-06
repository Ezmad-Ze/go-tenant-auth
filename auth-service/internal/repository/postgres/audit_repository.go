package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ezmad/auth-service/gen/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuditRepository interface {
	CreateAuditLog(ctx context.Context, params *CreateAuditLogParams) error
	GetAuditLogsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*AuditLog, error)
	GetAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int32) ([]*AuditLog, error)
	GetAuditLogsByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int32) ([]*AuditLog, error)
	GetAuditLogsByResource(ctx context.Context, tenantID uuid.UUID, resourceType string, resourceID uuid.UUID) ([]*AuditLog, error)
	GetAuditLogsByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, limit, offset int32) ([]*AuditLog, error)
	CountAuditLogsByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	CountAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID) (int64, error)
}

type auditRepository struct {
	querier db.Querier
}

func NewAuditRepository(querier db.Querier) AuditRepository {
	return &auditRepository{
		querier: querier,
	}
}

type CreateAuditLogParams struct {
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       string
	ResourceType *string
	ResourceID   *uuid.UUID
	Metadata     map[string]interface{}
	IPAddress    *string
	UserAgent    *string
	Status       string
	ErrorMessage *string
}

type AuditLog struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       string
	ResourceType *string
	ResourceID   *uuid.UUID
	Metadata     map[string]interface{}
	IPAddress    *string
	UserAgent    *string
	Status       string
	ErrorMessage *string
	CreatedAt    time.Time
}

func (r *auditRepository) CreateAuditLog(ctx context.Context, params *CreateAuditLogParams) error {
	var metadataJSON []byte
	var err error
	if params.Metadata != nil {
		metadataJSON, err = json.Marshal(params.Metadata)
		if err != nil {
			return err
		}
	}

	dbParams := db.CreateAuditLogParams{
		TenantID: params.TenantID,
		UserID: pgtype.UUID{
			Bytes: uuidToBytes(params.UserID),
			Valid: params.UserID != nil,
		},
		Action:       params.Action,
		ResourceType: params.ResourceType,
		ResourceID: pgtype.UUID{
			Bytes: uuidToBytes(params.ResourceID),
			Valid: params.ResourceID != nil,
		},
		Metadata:     metadataJSON,
		IpAddress:    params.IPAddress,
		UserAgent:    params.UserAgent,
		Status:       &params.Status,
		ErrorMessage: params.ErrorMessage,
	}

	_, err = r.querier.CreateAuditLog(ctx, dbParams)
	return err
}

func (r *auditRepository) GetAuditLogsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*AuditLog, error) {
	rows, err := r.querier.GetAuditLogsByTenant(ctx, db.GetAuditLogsByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	return convertAuditLogs(rows), nil
}

func (r *auditRepository) GetAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int32) ([]*AuditLog, error) {
	rows, err := r.querier.GetAuditLogsByUser(ctx, db.GetAuditLogsByUserParams{
		TenantID: tenantID,
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return convertAuditLogs(rows), nil
}

func (r *auditRepository) GetAuditLogsByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int32) ([]*AuditLog, error) {
	rows, err := r.querier.GetAuditLogsByAction(ctx, db.GetAuditLogsByActionParams{
		TenantID: tenantID,
		Action:   action,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	return convertAuditLogs(rows), nil
}

func (r *auditRepository) GetAuditLogsByResource(ctx context.Context, tenantID uuid.UUID, resourceType string, resourceID uuid.UUID) ([]*AuditLog, error) {
	rows, err := r.querier.GetAuditLogsByResource(ctx, db.GetAuditLogsByResourceParams{
		TenantID:     tenantID,
		ResourceType: &resourceType,
		ResourceID: pgtype.UUID{
			Bytes: resourceID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return convertAuditLogs(rows), nil
}

func (r *auditRepository) GetAuditLogsByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, limit, offset int32) ([]*AuditLog, error) {
	rows, err := r.querier.GetAuditLogsByDateRange(ctx, db.GetAuditLogsByDateRangeParams{
		TenantID: tenantID,
		CreatedAt: pgtype.Timestamptz{
			Time:  startDate,
			Valid: true,
		},
		CreatedAt_2: pgtype.Timestamptz{
			Time:  endDate,
			Valid: true,
		},
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return convertAuditLogs(rows), nil
}

func (r *auditRepository) CountAuditLogsByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return r.querier.CountAuditLogsByTenant(ctx, tenantID)
}

func (r *auditRepository) CountAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return r.querier.CountAuditLogsByUser(ctx, db.CountAuditLogsByUserParams{
		TenantID: tenantID,
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
	})
}

func convertAuditLogs(rows []db.AuditLog) []*AuditLog {
	logs := make([]*AuditLog, len(rows))
	for i, row := range rows {
		logs[i] = convertAuditLog(&row)
	}
	return logs
}

func convertAuditLog(row *db.AuditLog) *AuditLog {
	log := &AuditLog{
		ID:        row.ID,
		TenantID:  row.TenantID,
		Action:    row.Action,
		CreatedAt: row.CreatedAt.Time,
	}

	if row.Status != nil {
		log.Status = *row.Status
	}

	if row.UserID.Valid {
		userID := uuid.UUID(row.UserID.Bytes)
		log.UserID = &userID
	}

	if row.ResourceType != nil {
		log.ResourceType = row.ResourceType
	}

	if row.ResourceID.Valid {
		resourceID := uuid.UUID(row.ResourceID.Bytes)
		log.ResourceID = &resourceID
	}

	if len(row.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(row.Metadata, &metadata); err == nil {
			log.Metadata = metadata
		}
	}

	if row.IpAddress != nil {
		log.IPAddress = row.IpAddress
	}

	if row.UserAgent != nil {
		log.UserAgent = row.UserAgent
	}

	if row.ErrorMessage != nil {
		log.ErrorMessage = row.ErrorMessage
	}

	return log
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func uuidToBytes(id *uuid.UUID) [16]byte {
	if id == nil {
		return [16]byte{}
	}
	return *id
}
