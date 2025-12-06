# Development Guide

This guide covers how to set up the development environment, run the service locally, and contribute to the project.

## Prerequisites

- **Go**: Version 1.22 or higher
- **Docker & Docker Compose**: For running dependencies (PostgreSQL, Redis, MailHog)
- **Make**: For running automation scripts
- **Goose**: For database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- **SQLC**: For generating type-safe database code (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

## Local Setup

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/yourusername/auth-service.git
    cd auth-service
    ```

2.  **Install dependencies**:
    ```bash
    go mod download
    ```

3.  **Set up environment variables**:
    Copy the example configuration:
    ```bash
    cp .env.example .env
    ```
    Review `.env` and adjust any settings if needed. The defaults are configured to work with the Docker Compose setup.

4.  **Start infrastructure**:
    Start PostgreSQL, Redis, and MailHog:
    ```bash
    make docker-up
    ```

5.  **Run migrations**:
    Apply database schema changes:
    ```bash
    make migrate-up
    ```

6.  **Generate code (optional)**:
    If you modify SQL queries, regenerate the Go code:
    ```bash
    make sqlc
    ```

7.  **Run the service**:
    ```bash
    make run
    ```
    The server will start at `http://localhost:8083`.

## Project Structure

- `cmd/server`: Entry point of the application.
- `internal/config`: Configuration loading and validation.
- `internal/domain`: Core business entities and models.
- `internal/repository`: Database access layer (interfaces and implementations).
- `internal/service`: Business logic layer.
- `internal/http`: HTTP handlers, middleware, and routing.
- `sql/schema`: Database migration files.
- `sql/queries`: SQL queries used by SQLC.

## Workflow

### Adding a New Feature

1.  **Database Changes**:
    - Create a new migration: `make migrate-create NAME=add_feature_table`
    - Add SQL queries to `sql/queries/feature.sql`.
    - Run `make sqlc` to generate the repository code.

2.  **Business Logic**:
    - Define new models in `internal/domain`.
    - Create/Update service interfaces in `internal/service`.
    - Implement the logic in `internal/service`.

3.  **API Endpoint**:
    - Create a new handler in `internal/http/handlers`.
    - Register the route in `internal/http/routes/router.go`.
    - Add request/response structs and validation.

### Testing

- Run all tests:
    ```bash
    make test
    ```
- Run API verification script (integration test):
    ```bash
    ./scripts/local_dev/verify_api.sh
    ```

## Debugging

- **Logs**: The application uses `zerolog`. Set `LOG_LEVEL=debug` in `.env` for verbose output.
- **MailHog**: View sent emails at `http://localhost:8025`.
- **Database**: Connect to the DB using `psql` or a GUI tool:
    ```bash
    psql postgres://authuser:authpass@localhost:5436/authdb
    ```
    *(Note: Port 5436 is exposed by Docker Compose)*

## Code Style

- Follow standard Go conventions.
- Run `make fmt` to format code.
- Run `make lint` to check for issues (requires `golangci-lint`).
