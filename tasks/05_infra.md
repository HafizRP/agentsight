# ROLE SPECIFICATION: SRE & INFRASTRUCTURE ENGINEER (AGENT 5)
# TASK ID: TASK_5
# TARGET BRANCH: infra/docker-deploy
# DEPENDENCIES: [TASK_4]

## 1. OBJECTIVE
Author multi-stage container configurations with headless Chromium, namespaced Docker Compose, port guards, and health probes.

## 2. SCOPE OF IMPLEMENTATION

### Multi-Stage Dockerfile
```dockerfile
# Stage 1: Builder
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

# Stage 2: Runner with Chromium
FROM alpine:3.19
RUN apk add --no-cache chromium chromium-chromedriver ttf-freefont ca-certificates tzdata
ENV CHROME_BIN=/usr/bin/chromium-browser
COPY --from=builder /bin/server /bin/server
COPY migrations/ /app/migrations/
COPY internal/templates/ /app/templates/
COPY static/ /app/static/
EXPOSE 8080
ENTRYPOINT ["/bin/server"]
```

### docker-compose.yml
```yaml
name: agentsight
services:
  postgres:
    container_name: agentsight_postgres
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${DB_USER:-agentsight}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-agentsight}
      POSTGRES_DB: ${DB_NAME:-agentsight}
    ports:
      - "127.0.0.1:${DB_PORT:-5437}:5432"
    volumes:
      - agentsight_pgdata:/var/lib/postgresql/data
    networks:
      - agentsight_net
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-agentsight} -d ${DB_NAME:-agentsight}"]
      interval: 5s
      timeout: 5s
      retries: 10
  app:
    container_name: agentsight_app
    build: .
    ports:
      - "${APP_PORT:-8084}:8080"
    environment:
      DATABASE_URL: postgres://${DB_USER:-agentsight}:${DB_PASSWORD:-agentsight}@postgres:5432/${DB_NAME:-agentsight}?sslmode=disable
      CHROME_BIN: /usr/bin/chromium-browser
      GITHUB_TOKEN: ${GITHUB_TOKEN:-}
      GITHUB_CLIENT_ID: ${GITHUB_CLIENT_ID:-}
      GITHUB_CLIENT_SECRET: ${GITHUB_CLIENT_SECRET:-}
      SESSION_SECRET: ${SESSION_SECRET:-agentsight-dev-secret}
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - agentsight_net
volumes:
  agentsight_pgdata:
    name: agentsight_pgdata
networks:
  agentsight_net:
    name: agentsight_net
    driver: bridge
```

### `.env.example`
```env
DB_USER=agentsight
DB_PASSWORD=agentsight
DB_NAME=agentsight
DB_PORT=5437
APP_PORT=8084
SESSION_SECRET=change-me-in-production
GITHUB_TOKEN=              # Optional: GitHub PAT for higher API rate limits
GITHUB_CLIENT_ID=          # For GitHub OAuth login
GITHUB_CLIENT_SECRET=      # For GitHub OAuth login
```

### Health Probes
- Ensure `/healthz` (liveness) and `/readyz` (readiness: DB pool + scraper scheduler) are registered and functional.

### Pre-flight Port Check (`scripts/preflight.sh`)
- Check if `DB_PORT` and `APP_PORT` are available. If occupied, auto-increment and update `.env`.

## 3. VERIFICATION CRITERIA
1. `docker compose -p agentsight build` completes without errors.
2. `docker compose -p agentsight up -d` brings all services to healthy state.
3. `curl http://localhost:${APP_PORT:-8084}/healthz` returns HTTP 200.

## 4. EXIT PROTOCOL
1. Commit configuration changes to `infra/docker-deploy`.
2. Push branch and open PR, or stage for local merge.
