package observability

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TenantIDKey  contextKey = "tenant_id"
	UserIDKey    contextKey = "user_id"
)

var logger *slog.Logger

// InitLogger initializes the structured logger
func InitLogger(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// Logger returns the default logger
func Logger() *slog.Logger {
	if logger == nil {
		InitLogger(slog.LevelInfo)
	}
	return logger
}

// WithContext creates a logger with context values
func WithContext(ctx context.Context) *slog.Logger {
	l := Logger()
	
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		l = l.With("request_id", requestID)
	}
	
	if tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		l = l.With("tenant_id", tenantID.String())
	}
	
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		l = l.With("user_id", userID.String())
	}
	
	return l
}

// SetRequestID adds request ID to context
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// SetTenantID adds tenant ID to context
func SetTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// SetUserID adds user ID to context
func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// LogAuth logs authentication events
func LogAuth(ctx context.Context, event string, attrs ...any) {
	WithContext(ctx).Info(event, append([]any{"event_type", "auth"}, attrs...)...)
}

// LogAuthz logs authorization events
func LogAuthz(ctx context.Context, event string, attrs ...any) {
	WithContext(ctx).Info(event, append([]any{"event_type", "authz"}, attrs...)...)
}

// LogError logs errors with context
func LogError(ctx context.Context, msg string, err error, attrs ...any) {
	WithContext(ctx).Error(msg, append([]any{"error", err.Error()}, attrs...)...)
}

// LogSecurity logs security events
func LogSecurity(ctx context.Context, event string, attrs ...any) {
	WithContext(ctx).Warn(event, append([]any{"event_type", "security"}, attrs...)...)
}
