package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ezmad/auth-service/internal/repository/postgres"
	"github.com/google/uuid"
	zerologlog "github.com/rs/zerolog/log"
)

// AuditService handles audit logging for compliance and security monitoring
type AuditService interface {
	LogEvent(ctx context.Context, event *AuditEvent) error
	Close() error
}

// AuditEvent represents an auditable action in the system
type AuditEvent struct {
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       string
	ResourceType *string
	ResourceID   *uuid.UUID
	Metadata     map[string]interface{}
	IPAddress    *string
	UserAgent    *string
	Status       AuditStatus
	Error        error
}

// AuditStatus represents the outcome of an audited action
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailure AuditStatus = "failure"
	AuditStatusError   AuditStatus = "error"
)

// Audit action constants
const (
	// Authentication actions
	ActionLogin          = "login"
	ActionLoginFailed    = "login_failed"
	ActionLogout         = "logout"
	ActionRegister       = "register"
	ActionPasswordChange = "password_change"
	ActionPasswordReset  = "password_reset"
	ActionMagicLinkSent  = "magic_link_sent"
	ActionMagicLinkUsed  = "magic_link_used"

	// Email verification
	ActionEmailVerified      = "email_verified"
	ActionVerificationSent   = "verification_sent"

	// 2FA actions
	ActionTOTPEnabled       = "totp_enabled"
	ActionTOTPDisabled      = "totp_disabled"
	ActionTOTPVerified      = "totp_verified"
	ActionTOTPFailed        = "totp_failed"
	ActionBackupCodeUsed    = "backup_code_used"
	ActionBackupCodeGenerated = "backup_code_generated"

	// Authorization actions
	ActionRoleAssigned      = "role_assigned"
	ActionRoleRevoked       = "role_revoked"
	ActionPermissionGranted = "permission_granted"
	ActionPermissionRevoked = "permission_revoked"
	ActionPermissionChecked = "permission_checked"

	// Administrative actions
	ActionTenantCreated = "tenant_created"
	ActionTenantUpdated = "tenant_updated"
	ActionUserCreated   = "user_created"
	ActionUserUpdated   = "user_updated"
	ActionUserDeleted   = "user_deleted"
)

type auditService struct {
	repo       postgres.AuditRepository
	eventChan  chan *AuditEvent
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	bufferSize int
}

// NewAuditService creates a new audit service with async logging
func NewAuditService(repo postgres.AuditRepository, bufferSize int) AuditService {
	ctx, cancel := context.WithCancel(context.Background())
	
	svc := &auditService{
		repo:       repo,
		eventChan:  make(chan *AuditEvent, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
		bufferSize: bufferSize,
	}

	// Start background worker
	svc.wg.Add(1)
	go svc.processEvents()

	return svc
}

// LogEvent queues an audit event for async processing
func (s *auditService) LogEvent(ctx context.Context, event *AuditEvent) error {
	select {
	case s.eventChan <- event:
		return nil
	case <-time.After(100 * time.Millisecond):
		// If buffer is full, log synchronously to avoid data loss
		zerologlog.Warn().Msg("Audit buffer full, logging synchronously")
		return s.logEventSync(ctx, event)
	}
}

// Close gracefully shuts down the audit service
func (s *auditService) Close() error {
	s.cancel()
	close(s.eventChan)
	s.wg.Wait()
	return nil
}

// processEvents processes audit events from the channel
func (s *auditService) processEvents() {
	defer s.wg.Done()

	for {
		select {
		case event, ok := <-s.eventChan:
			if !ok {
				// Channel closed, drain remaining events
				return
			}
			if err := s.logEventSync(s.ctx, event); err != nil {
				zerologlog.Error().
					Err(err).
					Str("action", event.Action).
					Str("tenant_id", event.TenantID.String()).
					Msg("Failed to log audit event")
			}
		case <-s.ctx.Done():
			// Drain remaining events before shutdown
			for event := range s.eventChan {
				if err := s.logEventSync(context.Background(), event); err != nil {
					zerologlog.Error().
						Err(err).
						Str("action", event.Action).
						Msg("Failed to log audit event during shutdown")
				}
			}
			return
		}
	}
}

// logEventSync synchronously writes an audit event to the database
func (s *auditService) logEventSync(ctx context.Context, event *AuditEvent) error {
	status := string(event.Status)
	if status == "" {
		if event.Error != nil {
			status = string(AuditStatusError)
		} else {
			status = string(AuditStatusSuccess)
		}
	}

	var errorMessage *string
	if event.Error != nil {
		msg := event.Error.Error()
		errorMessage = &msg
	}

	params := &postgres.CreateAuditLogParams{
		TenantID:     event.TenantID,
		UserID:       event.UserID,
		Action:       event.Action,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		Metadata:     event.Metadata,
		IPAddress:    event.IPAddress,
		UserAgent:    event.UserAgent,
		Status:       status,
		ErrorMessage: errorMessage,
	}

	return s.repo.CreateAuditLog(ctx, params)
}

// Helper functions for common audit scenarios

// LogAuthEvent logs an authentication-related event
func LogAuthEvent(svc AuditService, ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action string, ipAddr, userAgent *string, err error) {
	status := AuditStatusSuccess
	if err != nil {
		status = AuditStatusFailure
	}

	event := &AuditEvent{
		TenantID:  tenantID,
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddr,
		UserAgent: userAgent,
		Status:    status,
		Error:     err,
	}

	if logErr := svc.LogEvent(ctx, event); logErr != nil {
		zerologlog.Error().Err(logErr).Msg("Failed to log auth event")
	}
}

// LogPermissionChange logs a permission or role change
func LogPermissionChange(svc AuditService, ctx context.Context, tenantID, userID uuid.UUID, action string, resourceType string, resourceID uuid.UUID, metadata map[string]interface{}, ipAddr, userAgent *string) {
	event := &AuditEvent{
		TenantID:     tenantID,
		UserID:       &userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Metadata:     metadata,
		IPAddress:    ipAddr,
		UserAgent:    userAgent,
		Status:       AuditStatusSuccess,
	}

	if err := svc.LogEvent(ctx, event); err != nil {
		zerologlog.Error().Err(err).Msg("Failed to log permission change")
	}
}

// LogAdminAction logs an administrative action
func LogAdminAction(svc AuditService, ctx context.Context, tenantID, userID uuid.UUID, action string, resourceType string, resourceID *uuid.UUID, metadata map[string]interface{}, ipAddr, userAgent *string) {
	event := &AuditEvent{
		TenantID:     tenantID,
		UserID:       &userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		IPAddress:    ipAddr,
		UserAgent:    userAgent,
		Status:       AuditStatusSuccess,
	}

	if err := svc.LogEvent(ctx, event); err != nil {
		zerologlog.Error().Err(err).Msg("Failed to log admin action")
	}
}

// String returns the string representation of AuditStatus
func (s AuditStatus) String() string {
	return string(s)
}

// Validate checks if the audit event has required fields
func (e *AuditEvent) Validate() error {
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if e.Action == "" {
		return fmt.Errorf("action is required")
	}
	return nil
}
