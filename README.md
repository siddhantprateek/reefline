# Reefline

> Container image hygiene and runtime security for modern Kubernetes.

![](./assets/banner.png)

As more AI-generated code hits open source, managing application security is like herding cats during a power outage. Reefline gives security teams a simple, powerful way to scan container images for vulnerabilities, benchmark compliance, and layer inefficiencies — with AI-generated reports and an MCP server for agent-native access.

## Demo

<video src="https://github.com/user-attachments/assets/reef.mp4" controls autoplay loop muted width="100%"></video>

### ___Hackathon Submission___

> _For details about the tools and technologies used in this project, see [SUBMISSION.md](SUBMISSION.md)._


## Features

### Scanning & Analysis
- **Vulnerability Scanning** — Powered by [Grype (Anchore)](https://github.com/anchore/grype). Detects CVEs across OS packages and language ecosystems with severity classification (Critical, High, Medium, Low).
- **CIS Docker Benchmark** — Powered by [Dockle](https://github.com/goodwithtech/dockle). Checks your image against CIS Docker Benchmark best practices.
- **Layer Efficiency Analysis** — Powered by [Dive (Wagoodman)](https://github.com/wagoodman/dive). Breaks down each image layer, identifies wasted space and inefficient build patterns.
- **Image Inspection** — Skopeo-like metadata inspection via `containers/image`. Pulls image manifest, config, and SBOM data without a Docker daemon.

### AI-Powered Reports
- **Multi-agent report generation** — A Supervisor agent inspects SBOM, CIS data, and layer data, then drafts a report. A Critique agent reviews and improves it.
- **Actionable recommendations** — Vulnerability insights, risk assessments, layer efficiency breakdowns, and optimized Dockerfile suggestions.
- **Multiple AI providers** — OpenAI, Anthropic, Google AI, OpenRouter — configure whichever fits your stack.

### Integrations
- **GitHub / GHCR** — List repositories, fetch Dockerfiles, list container images, and auto-create GitHub issues with optimization results.
- **Docker Hub** — Browse repositories and tags.
- **Harbor** — Browse projects and artifacts.
- **Kubernetes** — In-cluster image discovery across all namespaces, no credentials required.

### MCP Server

Reefline ships an MCP server for its backend, compatible with any MCP client — Cursor, Claude Desktop, Archestra Chat Interface. Query jobs, pull reports, and interact with security data conversationally.

**Run as HTTP (recommended for remote/agent access):**

```bash
bun run start:http
```

**Run as stdio (for local Claude Desktop / Cursor):**

```bash
bun run start
```

**Configure in Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json`):**

```json
{
  "mcpServers": {
    "reefline": {
      "command": "node",
      "args": ["/path/to/reefline/mcp/build/index.js"]
    }
  }
}
```

**Configure in Cursor (`.cursor/mcp.json`):**

```json
{
  "mcpServers": {
    "reefline": {
      "url": "http://localhost:4000/mcp"
    }
  }
}
```

> For HTTP mode, expose locally via `ngrok http 4000` and use the ngrok URL as the server URL.



### Metrics & Observability
- Per-tool performance metrics (avg duration, success rate, percentiles).
- Job trend analytics (hourly / daily aggregation).
- Queue statistics (active, pending, scheduled, completed, failed, throughput).
- OpenTelemetry tracing via Fiber middleware.

### Security
- All integration credentials (GitHub tokens, Docker Hub passwords, Harbor tokens) encrypted at rest using **AES-256-GCM** before storing in PostgreSQL.

---


## Getting Started

### Prerequisites

- Go 1.22+
- Docker + Docker Compose

### 1. Start infrastructure

```bash
cd deploy
docker-compose up -d
```

This starts PostgreSQL, MinIO, and Redis.

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values (see [Environment Variables](#environment-variables) below).

### 3. Run the server

```bash
go run cmd/server/main.go
```

### 4. Run the worker

```bash
go run cmd/worker/main.go
```

### 5. Start the dashboard

```bash
cd frontend/dashboard
npm install
npm run dev
```

## Project Structure

```
reefline/
├── cmd/
│   ├── server/         ← HTTP API server entry point
│   ├── worker/         ← Job processing worker entry point
│   └── debug/          ← Queue stats debug tool
├── internal/
│   ├── handlers/       ← HTTP request handlers
│   ├── queue/          ← Queue abstraction (Redis + in-memory)
│   ├── routes/         ← API route registration
│   └── worker/         ← Job processing logic
├── pkg/
│   ├── database/       ← PostgreSQL connection + migrations
│   ├── storage/        ← MinIO client
│   ├── crypto/         ← AES-256-GCM encryption
│   ├── models/         ← Database models
│   ├── tools/          ← Grype, Dockle, Dive, Skopeo wrappers
│   └── telemetry/      ← OpenTelemetry config
├── frontend/
│   ├── dashboard/      ← Main dashboard (React + Vite)
│   └── web/            ← Public site (Next.js)
└── deploy/
    ├── docker-compose.yml
    └── docker-compose.harbor.yml
```


## Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/tools/...

# Specific test
go test ./pkg/tools/ -run TestDockleScanner
```

> Tests for scanning tools require Grype, Dockle, and Dive to be installed on the system.

---

## Built With

- [Archestra AI](https://archestra.ai) — LLM Proxy, MCP Registry, and Tool Policies used to keep agent access locked down and observable.
- [Grype](https://github.com/anchore/grype) — Vulnerability scanner
- [Dockle](https://github.com/goodwithtech/dockle) — CIS Docker Benchmark
- [Dive](https://github.com/wagoodman/dive) — Image layer analyzer
- [Fiber](https://github.com/gofiber/fiber) — Go HTTP framework

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) to get started.

This project follows a [Contributor Code of Conduct](CODE_OF_CONDUCT.md).

## Security

For vulnerability reports, see [SECURITY.md](SECURITY.md).

## License

Licensed under the terms of the [LICENSE](LICENSE) file.
