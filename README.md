# Auth Service

A production-grade authentication and authorization microservice built with Go, featuring multi-tenancy, RBAC+ABAC permissions, 2FA, and comprehensive session management.

## 🚀 Features

### Authentication
- ✅ User registration and login with argon2id password hashing
- ✅ JWT/PASETO token-based authentication
- ✅ Access token + refresh token with rotation
- ✅ Magic link passwordless authentication
- ✅ Email verification
- ✅ Password reset
- ✅ Session management with device tracking
- ✅ Failed login protection with Redis-based rate limiting
- ✅ 2FA support (TOTP with QR codes and backup codes)

### Authorization
- ✅ Dynamic RBAC (Role-Based Access Control)
- ✅ ABAC (Attribute-Based Access Control) with contextual rules
- ✅ Dynamic resources (project, invoice, user, etc.)
- ✅ Dynamic scopes (create, read, update, delete, etc.)
- ✅ Permissions = resource + scope (e.g., `project.read`, `invoice.update`)
- ✅ Many-to-many relationships:
  - Users ↔ Roles
  - Roles ↔ Permissions
  - Users ↔ Direct Permissions
- ✅ Permission check endpoint with ABAC rule evaluation
- ✅ Expirable role and permission assignments

### Multi-Tenancy
- ✅ Full tenant isolation at database level
- ✅ Tenant context extraction from header or subdomain
- ✅ Per-tenant resources, roles, and permissions

## 📁 Project Structure

```
auth-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── domain/
│   │   └── models.go            # Domain models
│   ├── repository/
│   │   └── interfaces.go        # Repository interfaces
│   ├── service/
│   │   ├── interfaces.go        # Service interfaces
│   │   ├── auth_service.go      # Auth service implementation
│   │   └── authorization_service.go  # Authorization service
│   ├── http/
│   │   ├── handlers/
│   │   │   ├── auth_handler.go       # Auth HTTP handlers
│   │   │   └── authorization_handler.go  # Authz HTTP handlers
│   │   ├── middleware/
│   │   │   ├── auth.go          # Authentication middleware
│   │   │   ├── tenant.go        # Tenant extraction middleware
│   │   │   └── logger.go        # Logging middleware
│   │   └── routes/
│   │       └── router.go        # Route registration
│   ├── security/
│   │   ├── argon2.go            # Argon2id password hashing
│   │   ├── token.go             # PASETO token management
│   │   └── totp.go              # TOTP 2FA implementation
│   └── util/
│       └── http.go              # HTTP utilities
├── sql/
│   ├── schema/                   # Goose migrations
│   │   ├── 00001_create_tenants.sql
│   │   ├── 00002_create_users.sql
│   │   ├── 00003_create_sessions.sql
│   │   ├── 00004_create_devices.sql
│   │   ├── 00005_create_magic_links.sql
│   │   ├── 00006_create_totp_2fa.sql
│   │   ├── 00007_create_authorization_basics.sql
│   │   ├── 00008_create_roles.sql
│   │   └── 00009_create_triggers.sql
│   └── queries/                  # SQLC query files
│       ├── users.sql
│       ├── sessions.sql
│       ├── permissions.sql
│       ├── roles.sql
│       ├── tenants.sql
│       └── auth_features.sql
├── gen/                          # SQLC generated code (created on build)
├── deploy/
│   └── docker/
│       ├── Dockerfile
│       └── docker-compose.yaml
├── go.mod
├── go.sum
├── Makefile
├── sqlc.yaml
├── .env.example
└── README.md
```

## 🛠 Tech Stack

- **Language**: Go 1.22+
- **Router**: chi v5
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **ORM**: SQLC for type-safe SQL
- **Migrations**: Goose
- **Password Hashing**: argon2id
- **Tokens**: PASETO v2 (symmetric encryption)
- **2FA**: TOTP with QR code generation
- **Logging**: zerolog
- **Docker**: Multi-stage builds with Alpine

## 📚 Documentation

- **[Development Guide](docs/DEVELOPMENT.md)**: Setup, local development, and contribution workflow.
- **[Deployment Guide](docs/DEPLOYMENT.md)**: Production deployment via Docker Compose or Kubernetes.
- **[API Documentation](docs/API.md)**: (Coming soon) Detailed API reference.

## ⚙️ Quick Start

For detailed instructions on how to get started, please refer to the [Development Guide](docs/DEVELOPMENT.md).

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application locally
make test              # Run tests
make docker-build      # Build Docker image
make docker-up         # Start all services with docker-compose
make docker-down       # Stop all services
make migrate-up        # Run all up migrations
make migrate-down      # Rollback last migration
make migrate-create    # Create a new migration (NAME=migration_name)
make sqlc              # Generate SQLC code
make fmt               # Format Go code
make lint              # Run linter
make clean             # Clean build artifacts
```

## 🔐 API Endpoints

### Authentication

```
POST   /api/v1/auth/register         # Register new user
POST   /api/v1/auth/login            # Login user
POST   /api/v1/auth/logout           # Logout current session
POST   /api/v1/auth/refresh          # Refresh access token
POST   /api/v1/auth/magic-link/send  # Send magic link email
POST   /api/v1/auth/magic-link/verify # Verify magic link token
POST   /api/v1/auth/password/change  # Change password
POST   /api/v1/auth/password/reset   # Request password reset
POST   /api/v1/auth/email/verify     # Verify email address
```

### Authorization (Protected Routes)

```
POST   /api/v1/authz/check                    # Check user permission
GET    /api/v1/authz/permissions              # Get user permissions
POST   /api/v1/authz/roles                    # Create role
POST   /api/v1/authz/roles/assign             # Assign role to user
DELETE /api/v1/authz/roles/revoke             # Revoke role from user
GET    /api/v1/authz/users/{userID}/roles     # Get user's roles
POST   /api/v1/authz/permissions              # Create permission
POST   /api/v1/authz/permissions/assign-to-role   # Assign permission to role
POST   /api/v1/authz/permissions/assign-to-user   # Assign permission to user
POST   /api/v1/authz/resources                # Create resource
POST   /api/v1/authz/scopes                   # Create scope
GET    /api/v1/authz/resources                # List resources
GET    /api/v1/authz/scopes                   # List scopes
```

## 🎯 Usage Examples

### Register a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

### Check Permission (RBAC + ABAC)

```bash
curl -X POST http://localhost:8080/api/v1/authz/check \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "user_id": "user-uuid-here",
    "resource": "project",
    "scope": "update",
    "attributes": {
      "project.owner_id": "user-uuid-here"
    }
  }'
```

## 🔧 Configuration

Key environment variables (see `.env.example` for full list):

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379/0` |
| `PASETO_SYMMETRIC_KEY` | 32-char key for PASETO tokens | Required |
| `ACCESS_TOKEN_DURATION` | Access token validity | `15m` |
| `REFRESH_TOKEN_DURATION` | Refresh token validity | `168h` |
| `MAX_FAILED_LOGIN_ATTEMPTS` | Failed login threshold | `5` |
| `TENANT_HEADER_NAME` | Header name for tenant ID | `X-Tenant-ID` |

## 🏗 Architecture

### Clean Architecture Layers

1. **Domain Layer** (`internal/domain`) - Core business entities
2. **Repository Layer** (`internal/repository`) - Data access interfaces
3. **Service Layer** (`internal/service`) - Business logic
4. **HTTP Layer** (`internal/http`) - HTTP handlers and routing
5. **Infrastructure** - External dependencies (DB, Redis, SMTP)

### Database Schema

- **tenants** - Multi-tenant organizations
- **users** - User accounts with authentication data
- **sessions** - Active user sessions and tokens
- **devices** - Registered user devices
- **magic_links** - Passwordless authentication tokens
- **totp_2fa** - TOTP 2FA secrets and backup codes
- **resources** - Authorization resources (entities)
- **scopes** - Authorization scopes (actions)
- **permissions** - Resource + Scope combinations
- **roles** - User roles
- **role_permissions** - Role-Permission assignments
- **user_roles** - User-Role assignments
- **user_permissions** - Direct user-permission assignments

## 🚦 Development Workflow

### Implementing a Repository

1. Add SQLC queries to `sql/queries/*.sql`
2. Run `make sqlc` to generate Go code
3. Implement repository interface in `internal/repository`
4. Wire up in `cmd/server/main.go`

### Adding a New Feature

1. Define domain models in `internal/domain`
2. Create repository interface and implementation
3. Implement service layer business logic
4. Add HTTP handlers
5. Register routes in `internal/http/routes`

## 📝 Roadmap

### Completed ✅
- [x] Multi-tenant architecture
- [x] User authentication (Password, Magic Link, 2FA)
- [x] RBAC & ABAC Authorization
- [x] Session Management & Device Tracking
- [x] Audit Logging
- [x] Rate Limiting

### Upcoming 🚧
- [ ] **Admin UI**: Web dashboard for managing tenants, users, and roles.
- [ ] **OIDC Provider**: Support for "Login with Auth Service" (OAuth2/OIDC).
- [ ] **API Keys**: Management for machine-to-machine authentication.
- [ ] **Webhooks**: Event notifications for user actions.


## 📄 License

MIT License - feel free to use this as a foundation for your projects!

## 🤝 Contributing

Contributions are welcome! Please follow the existing code structure and add tests for new features.
