package routes

import (
	"net/http"

	"github.com/ezmad/auth-service/internal/http/handlers"
	"github.com/ezmad/auth-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Router holds all route dependencies
type Router struct {
	authHandler  *handlers.AuthHandler
	authzHandler *handlers.AuthorizationHandler
	authMW       *middleware.AuthMiddleware
	tenantMW     *middleware.TenantMiddleware
	corsOptions  cors.Options
}

// NewRouter creates a new router
func NewRouter(
	authHandler *handlers.AuthHandler,
	authzHandler *handlers.AuthorizationHandler,
	authMW *middleware.AuthMiddleware,
	tenantMW *middleware.TenantMiddleware,
	corsOptions cors.Options,
) *Router {
	return &Router{
		authHandler:  authHandler,
		authzHandler: authzHandler,
		authMW:       authMW,
		tenantMW:     tenantMW,
		corsOptions:  corsOptions,
	}
}

// Setup configures all routes
func (router *Router) Setup() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(router.corsOptions))
	r.Use(router.tenantMW.ExtractTenant)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Group(func(r chi.Router) {
			r.Route("/auth", func(r chi.Router) {
				router.authHandler.RegisterRoutes(r)
			})
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(router.authMW.Authenticate)

			// Authorization management
			r.Route("/authz", func(r chi.Router) {
				router.authzHandler.RegisterRoutes(r)
			})

			// User profile routes would go here
			r.Route("/users", func(r chi.Router) {
				// User management endpoints
			})
		})
	})

	return r
}
