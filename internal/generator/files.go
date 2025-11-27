package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateMakefile() error {
	data := MakefileTemplateData{
		ProjectName:   g.config.ProjectName,
		GoVersion:     g.config.GoVersion,
		IncludeDocker: g.config.IncludeDocker,
	}
	return g.writeEmbeddedTemplate("Makefile", "Makefile.tmpl", data)
}

func (g *Generator) generateEnvFile() error {
	if !g.config.EnvSample {
		// Generate minimal .env.example
		return g.generateMinimalEnvFile()
	}

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf(`# %s Environment Configuration
# ============================================
# This file contains all environment variables for the application.
# Copy this file to .env and modify values as needed.
#
# Usage:
#   cp .env.example .env
#   # Edit .env with your values
#
# ============================================

`, g.config.ProjectName))

	// Application settings
	sb.WriteString(`# ============================================
# Application Settings
# ============================================

# Environment: development, staging, production
ENVIRONMENT=development

# HTTP server port
PORT=8080

# Log level: debug, info, warn, error
LOG_LEVEL=info

`)

	// Database settings
	if g.config.HasDatabase("postgres") || g.config.HasDatabase("mysql") || g.config.HasDatabase("mongodb") {
		sb.WriteString(`# ============================================
# Database Configuration
# ============================================

`)
		if g.config.HasDatabase("postgres") {
			sb.WriteString(`# PostgreSQL connection string
# Format: postgres://user:password@host:port/database?sslmode=disable
POSTGRES_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable

# PostgreSQL connection pool settings (optional)
# POSTGRES_MAX_CONNECTIONS=25
# POSTGRES_MAX_IDLE_TIME=5m

`)
		}
		if g.config.HasDatabase("mysql") {
			sb.WriteString(`# MySQL connection string
# Format: user:password@tcp(host:port)/database
MYSQL_URL=user:password@tcp(localhost:3306)/dbname

# MySQL connection pool settings (optional)
# MYSQL_MAX_CONNECTIONS=25
# MYSQL_MAX_IDLE_TIME=5m

`)
		}
		if g.config.HasDatabase("mongodb") {
			sb.WriteString(`# MongoDB connection string
# Format: mongodb://[user:password@]host:port[/database][?options]
MONGO_URL=mongodb://localhost:27017
MONGO_DATABASE=dbname

# MongoDB connection pool settings (optional)
# MONGO_MAX_POOL_SIZE=100
# MONGO_MIN_POOL_SIZE=10

`)
		}
	}

	// Cache settings
	if g.config.HasDatabase("redis") {
		sb.WriteString(`# ============================================
# Cache Configuration
# ============================================

# Redis connection string
# Format: redis://[user:password@]host:port[/db]
REDIS_URL=redis://localhost:6379

# Redis connection pool settings (optional)
# REDIS_POOL_SIZE=10
# REDIS_MIN_IDLE_CONNS=5

`)
	}

	// Observability settings
	if g.config.EnableTracing || g.config.EnableMetrics {
		sb.WriteString(`# ============================================
# Observability Configuration
# ============================================

`)
		if g.config.EnableTracing {
			sb.WriteString(fmt.Sprintf(`# OpenTelemetry Configuration
# OTLP endpoint for trace export (gRPC)
OTLP_ENDPOINT=localhost:4317

# Service name for tracing
SERVICE_NAME=%s

# Trace sampling rate: 0.0 to 1.0 (1.0 = 100%% sampling)
# TRACE_SAMPLE_RATE=1.0

# Enable/disable tracing
# TRACING_ENABLED=true

`, g.config.ProjectName))
		}
		if g.config.EnableMetrics {
			sb.WriteString(`# Prometheus Metrics
# Enable/disable metrics endpoint
METRICS_ENABLED=true

# Metrics endpoint path (default: /metrics)
# METRICS_PATH=/metrics

`)
		}
	}

	// Security settings
	sb.WriteString(`# ============================================
# Security Configuration (Optional)
# ============================================

# JWT secret key for token signing (generate a strong random string)
# JWT_SECRET=your-super-secret-key-change-in-production

# JWT token expiration (e.g., 24h, 168h for 1 week)
# JWT_EXPIRATION=24h

# CORS allowed origins (comma-separated)
# CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Rate limiting (requests per minute)
# RATE_LIMIT=100

`)

	// API Keys section
	sb.WriteString(`# ============================================
# External Service API Keys (Optional)
# ============================================

# Example: Third-party service API key
# EXTERNAL_API_KEY=your-api-key

# Example: Webhook secret
# WEBHOOK_SECRET=your-webhook-secret

`)

	return g.writeFile(".env.example", sb.String())
}

func (g *Generator) generateMinimalEnvFile() error {
	envVars := []string{
		"ENVIRONMENT=development",
		"PORT=8080",
		"",
	}

	if g.config.HasDatabase("postgres") {
		envVars = append(envVars, "POSTGRES_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable")
	}
	if g.config.HasDatabase("mysql") {
		envVars = append(envVars, "MYSQL_URL=user:password@tcp(localhost:3306)/dbname")
	}
	if g.config.HasDatabase("mongodb") {
		envVars = append(envVars, "MONGO_URL=mongodb://localhost:27017")
	}
	if g.config.HasDatabase("redis") {
		envVars = append(envVars, "REDIS_URL=redis://localhost:6379")
	}

	if len(envVars) > 3 {
		envVars = append(envVars, "")
	}

	if g.config.EnableTracing {
		envVars = append(envVars,
			"OTLP_ENDPOINT=localhost:4317",
			fmt.Sprintf("SERVICE_NAME=%s", g.config.ProjectName),
			"",
		)
	}

	if g.config.EnableMetrics {
		envVars = append(envVars, "METRICS_ENABLED=true", "")
	}

	return g.writeFile(".env.example", strings.Join(envVars, "\n"))
}

func (g *Generator) generateReadme() error {
	features := []string{
		fmt.Sprintf("- **HTTP Framework**: %s", g.getFrameworkName()),
		fmt.Sprintf("- **Logger**: %s", g.getLoggerName()),
	}

	if len(g.config.Databases) > 0 {
		dbList := []string{}
		for _, db := range g.config.Databases {
			dbList = append(dbList, g.getDatabaseName(db))
		}
		features = append(features, fmt.Sprintf("- **Databases**: %s", strings.Join(dbList, ", ")))
	}

	if g.config.EnableTracing {
		features = append(features, "- **Distributed Tracing**: OpenTelemetry")
	}

	if g.config.EnableMetrics {
		features = append(features, "- **Metrics**: Prometheus")
	}

	setupSteps := []string{
		"1. Copy environment variables:",
		"   ```bash",
		"   cp .env.example .env",
		"   ```",
		"",
		"2. Install dependencies:",
		"   ```bash",
		"   go mod download",
		"   ```",
	}

	if g.config.IncludeDocker && len(g.config.Databases) > 0 {
		setupSteps = append(setupSteps, "",
			"3. Start Docker services:",
			"   ```bash",
			"   make docker-up",
			"   ```",
		)
	}

	content := fmt.Sprintf(`# %s

A production-ready Go HTTP server following twelve-factor app methodology and Go best practices.

## Features

%s

## Architecture

This project follows clean architecture principles:

- **cmd/%s**: Application entrypoint
- **internal/config**: Configuration management (factor I, III, X)
- **internal/server**: HTTP server setup
- **internal/handlers**: HTTP request handlers
- **internal/middleware**: HTTP middleware (logging, tracing, recovery)
- **internal/observability**: Logging, tracing, and metrics
%s

## Twelve-Factor App Compliance

1. **Codebase**: Single codebase tracked in Git
2. **Dependencies**: Managed via go.mod
3. **Config**: Environment variables via .env
4. **Backing Services**: Attachable resources via connection strings
5. **Build, Release, Run**: Separate stages via Makefile
6. **Processes**: Stateless processes
7. **Port Binding**: Self-contained HTTP server
8. **Concurrency**: Horizontal scaling ready
9. **Disposability**: Graceful shutdown implemented
10. **Dev/Prod Parity**: Same codebase and dependencies
11. **Logs**: Stdout/stderr stream handling
12. **Admin Processes**: Runnable via Go commands

## Setup

%s

## Running

Development mode:
`+"```bash\n"+`make run
`+"```\n\n"+`Production build:
`+"```bash\n"+`make build
./bin/%s
`+"```\n\n"+`## Testing

`+"```bash\n"+`make test
make coverage
`+"```\n\n"+`## API Endpoints

- `+"`GET /`"+` - Welcome message
- `+"`GET /health`"+` - Health check
- `+"`GET /ready`"+` - Readiness check
%s

## Configuration

All configuration is done via environment variables. See `+"`.env.example`"+` for available options.

## Observability

### Logging

Structured logging with %s. Logs are written to stdout in JSON format for production.

### Tracing

%s

### Metrics

%s

## License

MIT
`, g.config.ProjectName, strings.Join(features, "\n"), g.config.ProjectName, g.getDatabaseDirectories(),
		strings.Join(setupSteps, "\n"), g.config.ProjectName, g.getMetricsEndpoint(), g.getLoggerName(),
		g.getTracingInfo(), g.getMetricsInfo())

	return g.writeFile("README.md", content)
}

func (g *Generator) generateGitignore() error {
	content := `# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test coverage
*.out
coverage.html

# Go workspace
*.work

# Environment variables
.env

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
	return g.writeFile(".gitignore", content)
}

func (g *Generator) getFrameworkName() string {
	switch g.config.Framework {
	case "stdlib":
		return "net/http (standard library)"
	case "chi":
		return "Chi"
	case "gin":
		return "Gin"
	case "echo":
		return "Echo"
	case "fiber":
		return "Fiber"
	default:
		return "net/http"
	}
}

func (g *Generator) getLoggerName() string {
	switch g.config.Logger {
	case "slog":
		return "slog (standard library)"
	case "zap":
		return "Zap"
	case "zerolog":
		return "Zerolog"
	default:
		return "slog"
	}
}

func (g *Generator) getDatabaseName(db string) string {
	switch db {
	case "postgres":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	case "mongodb":
		return "MongoDB"
	case "redis":
		return "Redis"
	default:
		return db
	}
}

func (g *Generator) getDatabaseDirectories() string {
	dirs := []string{}
	if g.config.NeedsSQL() || g.config.NeedsNoSQL() {
		dirs = append(dirs, "- **internal/database**: Database connections")
	}
	if g.config.NeedsCache() {
		dirs = append(dirs, "- **internal/cache**: Cache layer")
	}
	if len(dirs) > 0 {
		return "\n" + strings.Join(dirs, "\n")
	}
	return ""
}

func (g *Generator) getMetricsEndpoint() string {
	if g.config.EnableMetrics {
		return "\n- `GET /metrics` - Prometheus metrics"
	}
	return ""
}

func (g *Generator) getTracingInfo() string {
	if g.config.EnableTracing {
		return "Distributed tracing with OpenTelemetry. Traces are exported to OTLP endpoint configured via `OTLP_ENDPOINT`."
	}
	return "Not enabled."
}

func (g *Generator) getMetricsInfo() string {
	if g.config.EnableMetrics {
		return "Prometheus metrics available at `/metrics` endpoint. Includes HTTP request count and duration."
	}
	return "Not enabled."
}
