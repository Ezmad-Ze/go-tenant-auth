# Deployment Guide

This guide describes how to deploy the `auth-service` to a production environment.

## Deployment Options

### 1. Docker Compose (Recommended for small deployments)

The provided `deploy/docker/docker-compose.yaml` is production-ready, provided you supply a secure `.env` file.

1.  **Prepare the server**: Ensure Docker and Docker Compose are installed.
2.  **Copy files**: Copy `docker-compose.yaml` and `.env` to the server.
3.  **Configure `.env`**:
    - Set `APP_ENV=production`.
    - Generate strong secrets for `ACCESS_TOKEN_SECRET`, `REFRESH_TOKEN_SECRET`, and `PASETO_SYMMETRIC_KEY`.
    - Configure a real SMTP server (e.g., AWS SES, SendGrid) instead of MailHog.
    - Set `CORS_ALLOWED_ORIGINS` to your frontend domain(s).
4.  **Start the service**:
    ```bash
    docker-compose up -d
    ```

### 2. Kubernetes (K8s)

For larger deployments, use Kubernetes.

1.  **Build and Push Image**:
    ```bash
    docker build -t your-registry/auth-service:latest -f deploy/docker/Dockerfile .
    docker push your-registry/auth-service:latest
    ```
2.  **Create Secrets**: Store database credentials and token keys in K8s Secrets.
3.  **Deploy**: Create Deployment and Service manifests. Ensure you define readiness and liveness probes pointing to `/health`.

## Environment Variables

Ensure these variables are set in your production environment:

| Variable | Description | Production Value |
|----------|-------------|------------------|
| `APP_ENV` | Environment mode | `production` |
| `LOG_LEVEL` | Logging verbosity | `info` or `warn` |
| `DATABASE_URL` | DB Connection | Point to managed DB (RDS, Cloud SQL) |
| `REDIS_URL` | Redis Connection | Point to managed Redis (ElastiCache) |
| `ACCESS_TOKEN_SECRET` | JWT signing key | **Must be 32+ chars, random** |
| `REFRESH_TOKEN_SECRET` | Refresh token key | **Must be 32+ chars, random** |
| `PASETO_SYMMETRIC_KEY` | PASETO encryption key | **Must be exactly 32 chars** |
| `SMTP_HOST` | Email server host | Real SMTP host |
| `SMTP_PASSWORD` | Email server password | Secure password |
| `CORS_ALLOWED_ORIGINS` | Allowed frontends | `https://your-app.com` |

## Security Checklist

- [ ] **HTTPS**: Ensure the service is behind a load balancer or reverse proxy (Nginx, Traefik) handling TLS termination.
- [ ] **Secrets**: Never commit `.env` files to version control. Use a secrets manager.
- [ ] **Database**: Run migrations (`make migrate-up`) before starting the new version of the application.
- [ ] **Network**: Restrict access to the database and Redis ports (5432, 6379) to only the application container.

## Health Checks

The service exposes a health check endpoint at:
- `GET /health`: Returns 200 OK if the service and Redis are reachable.

Configure your load balancer or orchestrator to monitor this endpoint.
