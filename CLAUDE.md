# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Reefline is a container image security and hygiene scanning platform for Kubernetes infrastructure. It uses a server-worker architecture to analyze container images for vulnerabilities, CIS Docker Benchmark compliance, layer efficiency, and generates optimization recommendations.

## Technology Stack

**Backend (Go 1.25.2):**
- HTTP Framework: Fiber v2
- Database: PostgreSQL + GORM
- Object Storage: MinIO
- Job Queue: Redis (Asynq) with in-memory fallback
- Telemetry: OpenTelemetry
- Security Tools:
  - Grype (Anchore) - vulnerability scanning
  - Dockle - CIS Docker Benchmark checks
  - Dive (Wagoodman) - layer efficiency analysis
  - containers/image - Skopeo-like image inspection

**Frontend:**
- Two separate React + TypeScript + Vite applications
- UI: shadcn/ui components with Tailwind CSS
- `frontend/dashboard/` - Main dashboard application
- `frontend/web/` - Public web application

## Architecture

### Command Entry Points (`cmd/`)

**Server (`cmd/server/`):**
- Runs the HTTP API server (Fiber)
- Handles job submission (enqueues to Redis/memory queue)
- Does NOT process analysis jobs (worker does this)
- Initializes: database, MinIO, encryption, telemetry, image inspector (optional)

**Worker (`cmd/worker/`):**
- Processes async image analysis jobs from the queue
- Initializes all security scanning tools (Grype, Dockle, Dive, Skopeo)
- Registers handler: `analyze_image` → `worker.ProcessAnalyzeJob`
- Stores results to MinIO and updates job status in PostgreSQL

**Debug (`cmd/debug/`):**
- `queue_stats.go` - Display Redis queue statistics (uses Asynq Inspector)

### Internal Packages (`internal/`)

**handlers/** - HTTP request handlers:
- `analyze.go` - Submit Dockerfile/image for analysis (POST /api/v1/analyze)
- `jobs.go` - Job CRUD operations
- `report.go` - Download analysis artifacts (report, SBOM, Dockerfile, graph)
- `compare.go` - Compare two analysis jobs
- `integration.go` - Manage integrations (GitHub, Docker Hub, Harbor)
- `sse.go` - Server-sent events for real-time job progress
- `health.go` - Health/readiness/liveness checks

**queue/** - Async job queue abstraction:
- `queue.go` - Queue interface
- `redis.go` - Redis implementation using Asynq
- `memory.go` - In-memory implementation for development

**routes/** - API routing:
- `routes.go` - All routes mounted under `/api/v1`

**worker/** - Job processing logic:
- Processes `analyze_image` jobs
- Orchestrates Grype, Dockle, Dive analysis
- Stores results to MinIO

**integration/** - External service integrations:
- `github/` - GitHub API and GHCR
- `dockerhub/` - Docker Hub API
- `harbor/` - Harbor registry API
- `ai/` - AI-powered optimization suggestions

### Public Packages (`pkg/`)

**database/** - PostgreSQL connection and migrations using GORM

**storage/** - MinIO object storage client

**crypto/** - AES-256-GCM encryption for sensitive data (credentials)

**models/** - Database models:
- `integration.go` - Integration credentials
- `job.go` - Analysis job tracking
- `vulnerability.go`, `alert.go`, `dockerfile.go` - Analysis results

**tools/** - Security scanning tool wrappers:
- `grype.go` - Vulnerability scanner
- `dockle.go` - CIS Docker Benchmark
- `dive.go` - Layer efficiency analyzer
- `skopeo.go` - Image inspector

**telemetry/** - OpenTelemetry configuration

## Development Commands

### Backend

**Run the server:**
```bash
go run cmd/server/main.go
```

**Run the worker:**
```bash
go run cmd/worker/main.go
```

**Check queue statistics:**
```bash
go run cmd/debug/queue_stats.go
```

**Run tests:**
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/tools/...

# Run a specific test
go test ./pkg/tools/ -run TestDockleScanner
```

**Build binaries:**
```bash
# Build all binaries
go build -o bin/server cmd/server/main.go
go build -o bin/worker cmd/worker/main.go
go build -o bin/queue-stats cmd/debug/queue_stats.go
```

### Frontend

**Dashboard:**
```bash
cd frontend/dashboard
npm run dev      # Start dev server
npm run build    # Build for production
npm run lint     # Run ESLint
npm run preview  # Preview production build
```

**Web:**
```bash
cd frontend/web
npm run dev      # Start dev server
npm run build    # Build for production
npm run lint     # Run ESLint
npm run preview  # Preview production build
```

### Infrastructure

**Start dependencies with Docker Compose:**
```bash
cd deploy
docker-compose up -d       # Start PostgreSQL, MinIO, Redis
docker-compose down        # Stop all services
docker-compose logs -f     # View logs
```

**Start with Harbor registry:**
```bash
cd deploy
docker-compose -f docker-compose.harbor.yml up -d
```

## Environment Configuration

Required environment variables (see [.env.example](.env.example)):

**Core:**
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - development/production

**Database:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`

**MinIO:**
- `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_USE_SSL`, `MINIO_DEFAULT_BUCKET`

**Redis (optional):**
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- If not set, falls back to in-memory queue

**Security Tools (worker only):**
- `VULNERABILITY_SCANNER_ENABLED=true` - Enable Grype
- `DOCKLE_SCANNER_ENABLED=true` - Enable Dockle
- `DIVE_ANALYZER_ENABLED=true` - Enable Dive
- `IMAGE_INSPECTOR_ENABLED=true` - Enable image inspector

**Telemetry:**
- `OTEL_ENABLED`, `OTEL_SERVICE_NAME`, `OTEL_SERVICE_VERSION`

## Key Architecture Decisions

**Server-Worker Separation:**
The server only handles HTTP requests and enqueues jobs. The worker runs the CPU/memory-intensive security scans. This allows horizontal scaling of workers independently.

**Queue Abstraction:**
The `internal/queue` package provides an interface with Redis (Asynq) and in-memory implementations. Server uses queue for enqueueing only; worker uses it to process jobs.

**Tool Initialization:**
Security scanning tools (Grype, Dockle, Dive) are initialized in the worker process based on environment flags. The server initializes only the image inspector for metadata operations.

**Encrypted Credentials:**
Integration credentials (GitHub tokens, Docker Hub passwords, Harbor tokens) are encrypted using AES-256-GCM before storing in PostgreSQL. The `pkg/crypto` package handles encryption/decryption.

**MinIO Storage Structure:**
Analysis results are stored in MinIO with the following structure:
- `{bucket}/{job_id}/report.json` - Full analysis report
- `{bucket}/{job_id}/dockerfile` - Optimized Dockerfile
- `{bucket}/{job_id}/sbom.json` - Software Bill of Materials
- `{bucket}/{job_id}/graph.svg` - Build graph visualization

## API Structure

All routes are under `/api/v1`:

**Health:**
- `GET /health`, `/health/ready`, `/health/live`

**Analysis:**
- `POST /analyze` - Submit image/Dockerfile for analysis

**Jobs:**
- `GET /jobs` - List jobs
- `GET /jobs/:id` - Get job status
- `DELETE /jobs/:id` - Delete job
- `GET /jobs/:id/stream` - SSE real-time progress
- `GET /jobs/:id/report` - Download JSON report
- `GET /jobs/:id/dockerfile` - Download optimized Dockerfile
- `GET /jobs/:id/sbom` - Download SBOM
- `GET /jobs/:id/graph` - Download build graph

**Compare:**
- `POST /compare` - Compare two analysis results

**Integrations:**
- `GET /integrations` - List integrations
- `POST /integrations/:id/connect` - Connect integration
- `POST /integrations/:id/disconnect` - Disconnect
- `POST /integrations/:id/test` - Test connection
- Provider-specific endpoints for GitHub, Docker Hub, Harbor

## Testing

Tests are located alongside source files using Go's `_test.go` convention:
- `pkg/tools/grype_test.go` - Grype integration tests
- `pkg/tools/dockle_test.go` - Dockle integration tests
- `pkg/tools/dive_test.go` - Dive integration tests
- `pkg/tools/skopeo_test.go` - Skopeo integration tests

These tests require the respective tools to be installed on the system.

## Database Migrations

Migrations run automatically on server/worker startup using GORM AutoMigrate. Models are registered in `cmd/server/main.go` and `cmd/worker/main.go`:

```go
database.AutoMigrate(db, &models.Integration{}, &models.Job{})
```
