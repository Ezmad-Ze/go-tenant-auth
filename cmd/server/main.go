package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ezmad/auth-service/internal/config"
	_ "github.com/ezmad/auth-service/internal/domain"
	"github.com/ezmad/auth-service/internal/http/handlers"
	"github.com/ezmad/auth-service/internal/http/middleware"
	"github.com/ezmad/auth-service/internal/infrastructure/db"
	"github.com/ezmad/auth-service/internal/infrastructure/email"
	"github.com/ezmad/auth-service/internal/infrastructure/redis"
	"github.com/ezmad/auth-service/internal/observability"
	"github.com/ezmad/auth-service/internal/repository/postgres"
	redisrepo "github.com/ezmad/auth-service/internal/repository/redis"
	"github.com/ezmad/auth-service/internal/security"
	_ "github.com/ezmad/auth-service/internal/service"
	"github.com/ezmad/auth-service/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/rs/zerolog"
	zerologlog "github.com/rs/zerolog/log"

	_ "github.com/ezmad/auth-service/docs/swagger" // Swagger docs
)

// @title Auth Service API
// @version 1.0
// @description Multi-tenant authentication and authorization service with TOTP 2FA, RBAC/ABAC, and comprehensive security features
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@authservice.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:6969
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Setup logger
	setupLogger()

	zerologlog.Info().Msg("Starting auth-service...")

	// Initialize observability
	observability.InitLogger(slog.LevelInfo)
	zerologlog.Info().Msg("Observability initialized")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zerologlog.Fatal().Err(err).Msg("Failed to load configuration")
	}

	ctx := context.Background()

	// Initialize database connection
	dbStore, err := db.NewStore(ctx, cfg.Database)
	if err != nil {
		zerologlog.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer dbStore.Close()

	// Initialize Redis client
	redisClient, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		zerologlog.Fatal().Err(err).Msg("Failed to initialize Redis")
	}
	defer redisClient.Close()

	// Initialize cache repository
	cacheRepo := redisrepo.NewCacheRepository(redisClient.GetClient())

	// Initialize security components
	argon2Params := &security.Argon2Params{
		Memory:      cfg.Security.Argon2Memory,
		Iterations:  cfg.Security.Argon2Iterations,
		Parallelism: cfg.Security.Argon2Parallelism,
		SaltLength:  cfg.Security.Argon2SaltLength,
		KeyLength:   cfg.Security.Argon2KeyLength,
	}

	tokenManager, err := security.NewPasetoToken(cfg.Token.PasetoSymmetricKey, cfg.Token.Issuer)
	if err != nil {
		zerologlog.Fatal().Err(err).Msg("Failed to initialize token manager")
	}

	// Initialize TOTP manager
	totpManager := security.NewTOTPManager(
		cfg.TOTP.Issuer,
		uint(cfg.TOTP.Period),
		cfg.TOTP.Digits,
	)

	// Initialize repositories with caching
	queries := dbStore.Querier()
	userRepo := postgres.NewUserRepository(queries, cacheRepo, cfg.Cache.UserTTL)
	sessionRepo := postgres.NewSessionRepository(queries, cacheRepo, cfg.Cache.SessionTTL)
	tenantRepo := postgres.NewTenantRepository(queries)
	_ = tenantRepo // Will be used later
	deviceRepo := postgres.NewDeviceRepository(queries)
	magicLinkRepo := postgres.NewMagicLinkRepository(queries)
	totpRepo := postgres.NewTOTPRepository(queries)
	roleRepo := postgres.NewRoleRepository(queries)
	permissionRepo := postgres.NewPermissionRepository(queries)
	auditRepo := postgres.NewAuditRepository(queries)

	// Initialize rate limiter
	rateLimiter := service.NewRateLimiter(cacheRepo)

	// Initialize failed login tracker
	failedLoginTracker := service.NewFailedLoginTracker(
		cacheRepo,
		cfg.Security.MaxFailedAttempts,
		cfg.Security.FailedLoginBlockDuration,
		cfg.Cache.FailedLoginTTL,
	)

	// Initialize TOTP rate limiter
	totpRateLimiter := service.NewTOTPRateLimiter(
		cacheRepo,
		5,                                       // max 5 attempts
		15*time.Minute,                         // 15-minute lockout
		cfg.Cache.FailedLoginTTL,               // attempt window TTL
	)

	// Initialize email service (mock for now)
	emailService := email.NewMockEmailService()

	// Initialize audit service with async logging
	auditService := observability.NewAuditService(auditRepo, 1000)
	defer auditService.Close()

	// Initialize authorization service
	authzService := service.NewAuthorizationService(permissionRepo, roleRepo, userRepo)
	_ = authzService // Will be used by middleware/handlers

	// Initialize tenant service
	tenantService := service.NewTenantService(tenantRepo, userRepo, roleRepo, argon2Params)

	// Initialize TOTP service
	totpService := service.NewTOTPService(totpRepo, userRepo, totpManager, argon2Params)

	// Initialize auth service
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		deviceRepo,
		magicLinkRepo,
		totpRepo,
		cacheRepo,
		emailService,
		failedLoginTracker,
		totpRateLimiter,
		argon2Params,
		tokenManager,
		totpManager,
		service.AuthConfig{
			AccessTokenDuration:      cfg.Token.AccessTokenDuration,
			RefreshTokenDuration:     cfg.Token.RefreshTokenDuration,
			MagicLinkDuration:        cfg.Token.MagicLinkDuration,
			MaxFailedAttempts:        cfg.Security.MaxFailedAttempts,
			FailedLoginBlockDuration: cfg.Security.FailedLoginBlockDuration,
			SessionMaxDevices:        cfg.Session.MaxDevices,
		},
	)

	// Initialize rate limit middleware
	rateLimitMW := middleware.NewRateLimitMiddleware(rateLimiter, middleware.RateLimitConfig{
		Enabled:        cfg.RateLimit.Enabled,
		RequestsPerMin: cfg.RateLimit.APIPerMin,
		Window:         time.Minute,
		KeyPrefix:      "ratelimit",
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, auditService)
	authzHandler := handlers.NewAuthorizationHandler(authzService, auditService)
	tenantHandler := handlers.NewTenantHandler(tenantService, auditService)
	totpHandler := handlers.NewTOTPHandler(totpService, auditService)

	// Setup router
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)

	// Public endpoints (no auth/tenant required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check Redis connectivity
		if err := redisClient.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Redis unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.Handle("/metrics", promhttp.Handler())
	
	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Apply middleware for protected routes
	r.Group(func(r chi.Router) {
		r.Use(observability.MetricsMiddleware)  // Prometheus metrics
		r.Use(middleware.TenantMiddleware)       // Extract tenant ID
		r.Use(middleware.AuthMiddleware(tokenManager)) // Validate JWT and extract user ID
		r.Use(rateLimitMW.Limit)                 // Apply rate limiting globally

		// Register routes
		r.Route("/auth", func(r chi.Router) {
			authHandler.RegisterRoutes(r)
			totpHandler.RegisterRoutes(r)
		})

		r.Route("/authz", func(r chi.Router) {
			authzHandler.RegisterRoutes(r)
		})

		tenantHandler.RegisterRoutes(r)
	})

	// Log successful initialization
	zerologlog.Info().
		Str("redis_url", cfg.Redis.URL).
		Bool("rate_limit_enabled", cfg.RateLimit.Enabled).
		Int("max_failed_attempts", cfg.Security.MaxFailedAttempts).
		Msg("Services initialized successfully")

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		zerologlog.Info().Str("port", cfg.App.Port).Msg("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zerologlog.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zerologlog.Info().Msg("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zerologlog.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	zerologlog.Info().Msg("Server exited")
}

func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	zerologlog.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// Placeholder initialization functions (would be implemented with actual DB/Redis clients)
// func initDatabase(cfg config.DatabaseConfig) (*pgxpool.Pool, error) { ... }
// func initRedis(cfg config.RedisConfig) (*redis.Client, error) { ... }
