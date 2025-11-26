package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateDockerFiles() error {
	if err := g.generateDockerfile(); err != nil {
		return err
	}

	if err := g.generateDockerCompose(); err != nil {
		return err
	}

	if err := g.generateDockerignore(); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateDockerfile() error {
	content := fmt.Sprintf(`# Build stage
FROM golang:%s-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/%s

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
`, g.config.GoVersion, g.config.ProjectName)

	return g.writeFile("Dockerfile", content)
}

func (g *Generator) generateDockerCompose() error {
	services := []string{
		`  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=development
      - PORT=8080`,
	}

	envVars := []string{}
	depends := []string{}

	if g.config.HasDatabase("postgres") {
		services = append(services, `  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: dbname
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 10s
      timeout: 5s
      retries: 5`)
		envVars = append(envVars, "      - POSTGRES_URL=postgres://user:password@postgres:5432/dbname?sslmode=disable")
		depends = append(depends, "postgres")
	}

	if g.config.HasDatabase("mysql") {
		services = append(services, `  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: dbname
      MYSQL_USER: user
      MYSQL_PASSWORD: password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5`)
		envVars = append(envVars, "      - MYSQL_URL=user:password@tcp(mysql:3306)/dbname")
		depends = append(depends, "mysql")
	}

	if g.config.HasDatabase("mongodb") {
		services = append(services, `  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/test --quiet
      interval: 10s
      timeout: 5s
      retries: 5`)
		envVars = append(envVars, "      - MONGO_URL=mongodb://mongodb:27017")
		depends = append(depends, "mongodb")
	}

	if g.config.HasDatabase("redis") {
		services = append(services, `  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5`)
		envVars = append(envVars, "      - REDIS_URL=redis://redis:6379")
		depends = append(depends, "redis")
	}

	if g.config.EnableTracing {
		envVars = append(envVars,
			"      - OTLP_ENDPOINT=jaeger:4317",
			fmt.Sprintf("      - SERVICE_NAME=%s", g.config.ProjectName),
		)
	}

	if g.config.EnableMetrics {
		envVars = append(envVars, "      - METRICS_ENABLED=true")
	}

	if len(envVars) > 0 {
		services[0] += "\n" + strings.Join(envVars, "\n")
	}

	if len(depends) > 0 {
		services[0] += "\n    depends_on:\n"
		for _, dep := range depends {
			services[0] += fmt.Sprintf("      - %s\n", dep)
		}
	}

	volumes := []string{}
	if g.config.HasDatabase("postgres") {
		volumes = append(volumes, "  postgres_data:")
	}
	if g.config.HasDatabase("mysql") {
		volumes = append(volumes, "  mysql_data:")
	}
	if g.config.HasDatabase("mongodb") {
		volumes = append(volumes, "  mongodb_data:")
	}
	if g.config.HasDatabase("redis") {
		volumes = append(volumes, "  redis_data:")
	}

	volumesSection := ""
	if len(volumes) > 0 {
		volumesSection = "\nvolumes:\n" + strings.Join(volumes, "\n")
	}

	content := fmt.Sprintf(`version: '3.8'

services:
%s
%s
`, strings.Join(services, "\n\n"), volumesSection)

	return g.writeFile("docker-compose.yml", content)
}

func (g *Generator) generateDockerignore() error {
	content := `# Git
.git
.gitignore

# Binaries
bin/
*.exe

# Test coverage
*.out

# Environment
.env

# IDE
.vscode/
.idea/

# OS
.DS_Store

# Build cache
vendor/
`
	return g.writeFile(".dockerignore", content)
}
