# Deployment Specification

> **Version:** 1.0.0
> **Last Updated:** 2026-03-31
> **Status:** Approved

---

## 1. Environments `[DEPLOY-001]`

| Environment | Purpose | URL | Auto-deploy |
|-------------|---------|-----|-------------|
| Local | Development | `http://localhost:3000` (frontend), `:8080` (API) | Manual |
| Staging | Integration testing | Configurable | On merge to `main` |
| Production | Live users | Configurable | Manual approval |

---

## 2. Environment Configuration `[DEPLOY-002]`

### 2.1 Required Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET` | JWT signing key (min 32 chars recommended) | `a3f8c2...` | **Yes** |

### 2.2 Optional Environment Variables (with defaults)

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8080` | API server HTTP port |
| `GRPC_PORT` | `9090` | Reserved gRPC port |
| `HOST_ADDR` | `0.0.0.0` | Bind address |
| `ENABLE_TLS` | `false` | Enable HTTPS |
| `TLS_CERT_PATH` | — | TLS certificate file |
| `TLS_KEY_PATH` | — | TLS private key file |
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon socket |
| `WORKSPACE_IMAGE` | `cloudide-workspace:latest` | Container image for workspaces |
| `CONTAINER_MEMORY_MB` | `4096` | Memory limit per container (MB) |
| `CONTAINER_CPU_SHARES` | `2048` | CPU shares per container |
| `CONTAINER_TIMEOUT_MIN` | `480` | Max container lifetime (minutes) |
| `DOCKER_NETWORK` | `cloudide-net` | Docker bridge network name |
| `ALLOWED_ORIGINS` | `https://ide.cloudcode.dev` | CORS origins (comma-separated) |
| `SESSION_TIMEOUT_HOURS` | `24` | Session TTL (hours) |
| `RATE_LIMIT_RPS` | `100` | Rate limit requests/second per IP |
| `RATE_LIMIT_BURST` | `200` | Rate limit burst size per IP |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `SHUTDOWN_TIMEOUT_SEC` | `30` | Graceful shutdown timeout |
| `WS_PING_INTERVAL_SEC` | `30` | WebSocket ping interval |
| `WS_PONG_TIMEOUT_SEC` | `40` | WebSocket pong timeout |
| `WS_WRITE_TIMEOUT_SEC` | `10` | WebSocket write timeout |
| `WS_MAX_MESSAGE_SIZE` | `65536` | Max WebSocket message (bytes) |
| `WS_READ_BUFFER_SIZE` | `8192` | WebSocket read buffer (bytes) |
| `WS_WRITE_BUFFER_SIZE` | `8192` | WebSocket write buffer (bytes) |

### 2.3 Environment-Specific Overrides

**Local (.env file):**
```bash
JWT_SECRET=local-dev-secret-32-characters-min
ALLOWED_ORIGINS=http://localhost:3000
LOG_LEVEL=debug
CONTAINER_MEMORY_MB=2048
```

**Staging:**
```bash
JWT_SECRET=${STAGING_JWT_SECRET}  # from secrets manager
ALLOWED_ORIGINS=https://staging.cloudcode.dev
LOG_LEVEL=info
ENABLE_TLS=true
TLS_CERT_PATH=/etc/ssl/certs/staging.pem
TLS_KEY_PATH=/etc/ssl/private/staging.key
```

**Production:**
```bash
JWT_SECRET=${PROD_JWT_SECRET}  # from secrets manager
ALLOWED_ORIGINS=https://ide.cloudcode.dev
LOG_LEVEL=warn
ENABLE_TLS=true
RATE_LIMIT_RPS=50
RATE_LIMIT_BURST=100
CONTAINER_MEMORY_MB=4096
SHUTDOWN_TIMEOUT_SEC=60
```

---

## 3. Docker Images `[DEPLOY-003]`

### 3.1 API Server Image (`Dockerfile.api`)

```
Stage 1: golang:1.22-alpine (builder)
  - Copies go.mod, go.sum, downloads modules
  - Copies source, builds with -ldflags="-w -s"
  - Output: /app/cloudide-api binary

Stage 2: alpine:3.19 (runtime)
  - Installs ca-certificates
  - Creates non-root user (appuser, UID 1000)
  - Copies binary from builder
  - Exposes 8080, 9090
  - Entrypoint: /app/cloudide-api
```

**Build:**
```bash
docker build -t cloudide-api:latest -f Dockerfile.api .
```

**Size:** ~20MB (Alpine + static Go binary)

### 3.2 Workspace Image (`workspace/Dockerfile`)

```
Base: ubuntu:22.04

Installed:
  - System: bash, curl, wget, git, sudo, vim, nano, tmux, htop, jq
  - Build: gcc, g++, make, cmake, pkg-config
  - Node.js 20 LTS + npm, yarn, pnpm, typescript, ts-node
  - Python 3.11 + pip, venv
  - Go 1.22 + gopls, dlv
  - Docker CLI

User: developer (non-root, with sudo)
Workdir: /workspace
Healthcheck: bash -c "echo ok"
```

**Build:**
```bash
docker build -t cloudide-workspace:latest ./workspace
```

**Size:** ~2.5GB

---

## 4. Docker Compose (Local Development) `[DEPLOY-004]`

**File:** `docker-compose.yml`

### Services

| Service | Image | Ports | Resources |
|---------|-------|-------|-----------|
| `api` | `cloudide-api:latest` | 8080, 9090 | — |
| `workspace-test` | `cloudide-workspace:latest` | — | 4GB RAM, 2 CPU |

### Volumes

| Volume | Mount | Purpose |
|--------|-------|---------|
| `/var/run/docker.sock` | API container | Container management |
| `workspace-data` | `/workspace` | Persistent workspace storage |

### Network

| Network | Driver | Purpose |
|---------|--------|---------|
| `cloudide-net` | bridge | Container-to-container communication |

### Health Checks

| Service | Check | Interval | Timeout | Retries |
|---------|-------|----------|---------|---------|
| `api` | `wget --spider http://localhost:8080/health` | 10s | 5s | 3 (start: 10s) |
| `workspace-test` | `bash -c "echo ok"` | 30s | 5s | 3 |

### Commands

```bash
make docker-up      # Build images + start services
make docker-down    # Stop + remove services + volumes
```

---

## 5. CI/CD Pipeline `[DEPLOY-005]`

### 5.1 Pipeline Configuration

**File:** `.github/workflows/ci.yml`

**Trigger:** Push to `main` and pull requests to `main`

### 5.2 Pipeline Jobs

```
ci.yml
  |
  +-- backend (Go)
  |     +-- Checkout
  |     +-- Setup Go 1.22
  |     +-- go build ./...
  |     +-- go vet ./...
  |     +-- go test -v -race -count=1 ./internal/...
  |
  +-- frontend (React)
        +-- Checkout
        +-- Setup Node 20
        +-- npm ci
        +-- npx tsc --noEmit
        +-- npm test
        +-- npx vite build
```

### 5.3 Pipeline Requirements

| Check | Blocking | Description |
|-------|----------|-------------|
| Go build | Yes | Code must compile |
| Go vet | Yes | Static analysis must pass |
| Go test (race) | Yes | All 109 tests must pass |
| TypeScript check | Yes | No type errors |
| Vitest | Yes | All 86 tests must pass |
| Vite build | Yes | Production build must succeed |

### 5.4 Deployment Pipeline (Planned)

```
PR merged to main
  |
  v
CI passes (build + test)
  |
  v
Docker images built and pushed to registry
  |
  v
Staging deployment (automatic)
  |
  v
Smoke tests pass
  |
  v
Production deployment (manual approval)
```

---

## 6. Build Targets `[DEPLOY-006]`

**File:** `Makefile`

| Target | Command | Description |
|--------|---------|-------------|
| `build` | `go build -ldflags="-w -s" -o cloudide-api ./cmd/server` | Compile Go binary |
| `test` | `go test -v -race ./...` | Run Go tests |
| `run` | `JWT_SECRET=dev-secret ./cloudide-api` | Run locally |
| `clean` | `rm -f cloudide-api` | Remove binary |
| `docker-build` | Build API + workspace images | Build all Docker images |
| `docker-up` | `docker compose up -d` | Start services |
| `docker-down` | `docker compose down -v` | Stop services |
| `lint` | `golangci-lint run ./...` | Go linting |
| `fmt` | `go fmt ./...` | Go formatting |
| `vet` | `go vet ./...` | Go static analysis |
| `coverage` | `go test -coverprofile=coverage.out ./internal/...` | Coverage report |
| `test-frontend` | `cd frontend && npm test` | Frontend tests |
| `test-all` | `make test && make test-frontend` | All tests |

---

## 7. Rollback Procedures `[DEPLOY-007]`

### 7.1 Application Rollback

```bash
# 1. Identify the previous working version
docker image ls cloudide-api --format "{{.Tag}} {{.CreatedAt}}"

# 2. Stop current deployment
docker compose down

# 3. Tag the previous image as latest
docker tag cloudide-api:<previous-tag> cloudide-api:latest

# 4. Restart services
docker compose up -d

# 5. Verify health
curl http://localhost:8080/health
```

### 7.2 Configuration Rollback

```bash
# 1. Revert .env to previous version
git checkout HEAD~1 -- .env

# 2. Restart services
docker compose restart api
```

### 7.3 Database Rollback (Future)

When a database is added:
1. All migrations must be reversible (up + down)
2. Rollback: run `migrate down` to revert the last migration
3. Data backups taken before every deployment

### 7.4 Emergency Procedures

| Situation | Action |
|-----------|--------|
| API unresponsive | Restart: `docker compose restart api` |
| Container leak (orphans) | `docker ps -a --filter name=workspace-` then `docker rm -f` |
| Memory exhaustion | Reduce `CONTAINER_MEMORY_MB`, restart |
| Auth compromise | Rotate `JWT_SECRET`, restart, all sessions invalidated |
| Rate limit too aggressive | Increase `RATE_LIMIT_RPS`/`BURST`, restart |

---

## 8. Monitoring & Observability `[DEPLOY-008]`

### 8.1 Health Check Endpoint

```bash
curl http://localhost:8080/health
```

Returns: `active_sessions`, `active_workspaces`, `timestamp`

### 8.2 Logging

| Field | Source | Description |
|-------|--------|-------------|
| Method | Request logger | HTTP method |
| Path | Request logger | URL path |
| Status | Request logger | Response status code |
| Duration | Request logger | Request processing time |
| X-Request-ID | Request ID middleware | Correlation ID |
| Component | Logger field | Source package name |

Log level controlled by `LOG_LEVEL` environment variable.

### 8.3 Key Metrics to Monitor

| Metric | Source | Alert Threshold |
|--------|--------|----------------|
| Active workspaces | `/health` | > 80% of host capacity |
| Active sessions | `/health` | > 10,000 |
| HTTP 5xx rate | Request logger | > 1% of requests |
| HTTP 429 rate | Rate limiter | > 10% of requests |
| Container creation failures | Container manager logs | Any |
| WebSocket disconnections | Client logs | > 50/minute |
| Memory usage per container | Docker stats | > 90% of limit |

### 8.4 Planned Observability (Future)

| Feature | Technology | Status |
|---------|-----------|--------|
| Metrics | Prometheus | Planned |
| Distributed tracing | OpenTelemetry + Jaeger | Planned |
| Log aggregation | ELK or Loki | Planned |
| Dashboards | Grafana | Planned |
| Alerting | PagerDuty / Grafana Alerting | Planned |

---

## 9. Security Deployment Checklist `[DEPLOY-009]`

Before deploying to production, verify:

- [ ] `JWT_SECRET` is unique, random, at least 32 characters
- [ ] `JWT_SECRET` is NOT committed to version control
- [ ] `ALLOWED_ORIGINS` set to actual production domain(s)
- [ ] `ENABLE_TLS=true` with valid certificate and key
- [ ] `RATE_LIMIT_RPS` and `RATE_LIMIT_BURST` tuned for expected load
- [ ] Docker socket is NOT mounted into workspace containers
- [ ] Workspace containers use `no-new-privileges` security option
- [ ] Container memory and CPU limits are set appropriately
- [ ] `CONTAINER_TIMEOUT_MIN` is set to prevent orphaned containers
- [ ] Health check endpoint is accessible by monitoring system
- [ ] `.env` file has restrictive permissions (`chmod 600`)
- [ ] Log level set to `warn` or `error` (not `debug`)
- [ ] Firewall restricts port 8080 to load balancer only

---

## Changelog

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-03-31 | CloudCode Team | Initial specification |
