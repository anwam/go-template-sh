package generator

import (
"path/filepath"
"strings"
"testing"

"github.com/anwam/go-template-sh/internal/config"
)

// TestGenerator_WithMemoryFS tests generation using in-memory file system.
// This is faster and more reliable than using the real file system.
func TestGenerator_WithMemoryFS(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "mem-test",
		ModulePath:    "github.com/test/mem-test",
		GoVersion:     "1.23",
		Framework:     "stdlib",
		Logger:        "slog",
		EnableTracing: true,
		EnableMetrics: true,
		IncludeDocker: true,
		ConfigFormat:  "env",
		Databases:     []string{"postgres", "redis"},
		CI:            "github",
	}

	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that expected files were created
	expectedFiles := []string{
		"go.mod",
		"cmd/mem-test/main.go",
		"internal/config/config.go",
		"internal/server/server.go",
		"internal/handlers/handlers.go",
		"internal/middleware/middleware.go",
		"internal/observability/observability.go",
		"internal/observability/logger.go",
		"internal/database/postgres.go",
		"internal/cache/redis.go",
		"Dockerfile",
		"docker-compose.yml",
		".dockerignore",
		".github/workflows/ci.yml",
		"Makefile",
		"README.md",
		".gitignore",
		".env.example",
	}

	for _, file := range expectedFiles {
		path := filepath.Join("/output/mem-test", file)
		if !mfs.HasFile(path) {
			t.Errorf("Expected file %s to exist", file)
		}
	}
}

func TestGenerator_GoModContent(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "gomod-test",
		ModulePath:    "github.com/test/gomod-test",
		GoVersion:     "1.23",
		Framework:     "chi",
		Logger:        "zap",
		Databases:     []string{"postgres", "mongodb"},
		EnableTracing: true,
		EnableMetrics: true,
		ConfigFormat:  "env",
	}

	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/gomod-test/go.mod")

	expectedDeps := []string{
		"github.com/go-chi/chi/v5",
		"go.uber.org/zap",
		"github.com/jackc/pgx/v5",
		"go.mongodb.org/mongo-driver",
		"go.opentelemetry.io/otel",
		"github.com/prometheus/client_golang",
		"github.com/google/uuid",
	}

	for _, dep := range expectedDeps {
		if !strings.Contains(content, dep) {
			t.Errorf("go.mod should contain %s", dep)
		}
	}

	if !strings.Contains(content, "module github.com/test/gomod-test") {
		t.Error("go.mod should contain correct module path")
	}

	if !strings.Contains(content, "go 1.23") {
		t.Error("go.mod should contain correct Go version")
	}
}

func TestGenerator_MainFileContent(t *testing.T) {
	cfg := createTestConfig()
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/cmd/test-project/main.go")

	checks := []string{
		"package main",
		"config.Load()",
		"observability.New",
		"server.New",
		"srv.Start()",
		"srv.Shutdown",
		"signal.Notify",
		"syscall.SIGINT",
		"syscall.SIGTERM",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("main.go should contain %q", check)
		}
	}
}

func TestGenerator_ServerFileContent_Chi(t *testing.T) {
	cfg := createTestConfig()
	cfg.Framework = "chi"
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/server/server.go")

	checks := []string{
		"package server",
		"github.com/go-chi/chi/v5",
		"chi.NewRouter()",
		"r.Get(\"/health\"",
		"r.Get(\"/ready\"",
		"r.Get(\"/\"",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("server.go should contain %q", check)
		}
	}
}

func TestGenerator_ServerFileContent_Gin(t *testing.T) {
	cfg := createTestConfig()
	cfg.Framework = "gin"
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/server/server.go")

	checks := []string{
		"package server",
		"github.com/gin-gonic/gin",
		"gin.New()",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("server.go should contain %q", check)
		}
	}
}

func TestGenerator_HandlerFileContent(t *testing.T) {
	cfg := createTestConfig()
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/handlers/handlers.go")

	checks := []string{
		"package handlers",
		"type Handler struct",
		"func NewHandler",
		"func (h *Handler) Health",
		"func (h *Handler) Ready",
		"func (h *Handler) Index",
		`json:"status"`,
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("handlers.go should contain %q", check)
		}
	}
}

func TestGenerator_MiddlewareFileContent(t *testing.T) {
	cfg := createTestConfig()
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/middleware/middleware.go")

	checks := []string{
		"package middleware",
		"func RequestID",
		"func Logger",
		"func Recoverer",
		"X-Request-ID",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("middleware.go should contain %q", check)
		}
	}
}

func TestGenerator_DockerfileContent(t *testing.T) {
	cfg := createTestConfig()
	cfg.IncludeDocker = true
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/Dockerfile")

	checks := []string{
		"FROM golang:1.23-alpine AS builder",
		"WORKDIR /app",
		"COPY go.mod",
		"go mod download",
		"CGO_ENABLED=0",
		"FROM alpine:latest",
		"EXPOSE 8080",
		"HEALTHCHECK",
		"CMD",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Dockerfile should contain %q", check)
		}
	}
}

func TestGenerator_DockerComposeWithDatabases(t *testing.T) {
	cfg := createTestConfig()
	cfg.IncludeDocker = true
	cfg.Databases = []string{"postgres", "redis", "mongodb"}
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/docker-compose.yml")

	checks := []string{
		"postgres:",
		"image: postgres:",
		"redis:",
		"image: redis:",
		"mongodb:",
		"image: mongo:",
		"depends_on:",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("docker-compose.yml should contain %q", check)
		}
	}
}

func TestGenerator_GitHubActionsContent(t *testing.T) {
	cfg := createTestConfig()
	cfg.CI = "github"
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/.github/workflows/ci.yml")

	checks := []string{
		"name: CI",
		"on:",
		"push:",
		"pull_request:",
		"go-version:",
		"go test",
		"golangci-lint",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("ci.yml should contain %q", check)
		}
	}
}

func TestGenerator_MakefileContent(t *testing.T) {
	cfg := createTestConfig()
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/Makefile")

	checks := []string{
		".PHONY:",
		"build:",
		"run:",
		"test:",
		"lint:",
		"clean:",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("Makefile should contain %q", check)
		}
	}
}

func TestGenerator_DatabasePostgres(t *testing.T) {
	cfg := createTestConfig()
	cfg.Databases = []string{"postgres"}
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/database/postgres.go")

	checks := []string{
		"package database",
		"type PostgresDB struct",
		"func NewPostgresDB",
		"pgxpool.New",
		"pool.Ping",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("postgres.go should contain %q", check)
		}
	}
}

func TestGenerator_CacheRedis(t *testing.T) {
	cfg := createTestConfig()
	cfg.Databases = []string{"redis"}
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/cache/redis.go")

	checks := []string{
		"package cache",
		"type RedisCache struct",
		"func NewRedisCache",
		"redis.NewClient",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("redis.go should contain %q", check)
		}
	}
}

func TestGenerator_ObservabilityWithTracing(t *testing.T) {
	cfg := createTestConfig()
	cfg.EnableTracing = true
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/observability/observability.go")

	checks := []string{
		"package observability",
		"go.opentelemetry.io/otel",
		"TracerProvider",
		"initTracer",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("observability.go should contain %q for tracing", check)
		}
	}
}

func TestGenerator_ObservabilityWithMetrics(t *testing.T) {
	cfg := createTestConfig()
	cfg.EnableMetrics = true
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := mfs.FileContent("/output/test-project/internal/observability/observability.go")

	checks := []string{
		"prometheus/client_golang",
		"httpRequestsTotal",
		"httpRequestDuration",
		"MetricsHandler",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("observability.go should contain %q for metrics", check)
		}
	}
}

func TestGenerator_ConfigFormats(t *testing.T) {
	formats := []struct {
		format       string
		expectedFile string
		checks       []string
	}{
		{
			format:       "yaml",
			expectedFile: "config.yaml.example",
			checks:       []string{"app:", "port:"},
		},
		{
			format:       "json",
			expectedFile: "config.json.example",
			checks:       []string{`"app":`, `"port":`},
		},
		{
			format:       "toml",
			expectedFile: "config.toml.example",
			checks:       []string{"[app]", "port ="},
		},
	}

	for _, tc := range formats {
		t.Run(tc.format, func(t *testing.T) {
cfg := createTestConfig()
			cfg.ConfigFormat = tc.format
			gen, mfs := createTestGenerator(cfg)

			if err := gen.Generate(); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			path := filepath.Join("/output/test-project", tc.expectedFile)
			if !mfs.HasFile(path) {
				t.Errorf("Expected %s to exist", tc.expectedFile)
				return
			}

			content := mfs.FileContent(path)
			for _, check := range tc.checks {
				if !strings.Contains(content, check) {
					t.Errorf("%s should contain %q", tc.expectedFile, check)
				}
			}
		})
	}
}

func TestGenerator_NoFilesOutsideProjectDir(t *testing.T) {
	cfg := createTestConfig()
	cfg.IncludeDocker = true
	cfg.CI = "github"
	cfg.Databases = []string{"postgres", "redis"}
	gen, mfs := createTestGenerator(cfg)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	projectDir := "/output/test-project"
	for path := range mfs.Files() {
		if !strings.HasPrefix(path, projectDir) {
			t.Errorf("File %s is outside project directory %s", path, projectDir)
		}
	}
}
