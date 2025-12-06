package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Request Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Authentication Metrics
	LoginAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"status", "tenant_id"},
	)

	RegistrationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_registrations_total",
			Help: "Total number of user registrations",
		},
		[]string{"tenant_id"},
	)

	TwoFactorUsageTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_2fa_usage_total",
			Help: "Total number of 2FA operations",
		},
		[]string{"operation", "status"},
	)

	PasswordResetsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_password_resets_total",
			Help: "Total number of password reset requests",
		},
		[]string{"tenant_id"},
	)

	// Session Metrics
	ActiveSessionsGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_active_sessions",
			Help: "Number of currently active sessions",
		},
		[]string{"tenant_id"},
	)

	// Authorization Metrics
	PermissionChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_permission_checks_total",
			Help: "Total number of permission checks",
		},
		[]string{"result", "tenant_id"},
	)

	// Cache Metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	// Database Metrics
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Business Metrics
	ActiveUsersGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_active_users",
			Help: "Number of active users",
		},
		[]string{"tenant_id"},
	)

	TenantsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_tenants_total",
			Help: "Total number of tenants",
		},
	)
)

// RecordHTTPRequest records an HTTP request with duration
func RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordLoginAttempt records a login attempt
func RecordLoginAttempt(status, tenantID string) {
	LoginAttemptsTotal.WithLabelValues(status, tenantID).Inc()
}

// RecordRegistration records a user registration
func RecordRegistration(tenantID string) {
	RegistrationsTotal.WithLabelValues(tenantID).Inc()
}

// Record2FAOperation records a 2FA operation
func Record2FAOperation(operation, status string) {
	TwoFactorUsageTotal.WithLabelValues(operation, status).Inc()
}

// RecordPasswordReset records a password reset request
func RecordPasswordReset(tenantID string) {
	PasswordResetsTotal.WithLabelValues(tenantID).Inc()
}

// SetActiveSessions sets the number of active sessions
func SetActiveSessions(tenantID string, count float64) {
	ActiveSessionsGauge.WithLabelValues(tenantID).Set(count)
}

// RecordPermissionCheck records a permission check
func RecordPermissionCheck(result, tenantID string) {
	PermissionChecksTotal.WithLabelValues(result, tenantID).Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cacheType string) {
	CacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cacheType string) {
	CacheMissesTotal.WithLabelValues(cacheType).Inc()
}

// RecordDatabaseQuery records a database query duration
func RecordDatabaseQuery(operation string, duration time.Duration) {
	DatabaseQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// SetActiveUsers sets the number of active users
func SetActiveUsers(tenantID string, count float64) {
	ActiveUsersGauge.WithLabelValues(tenantID).Set(count)
}

// SetTenantsTotal sets the total number of tenants
func SetTenantsTotal(count float64) {
	TenantsTotal.Set(count)
}
