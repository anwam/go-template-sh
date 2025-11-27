package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateGoMod() error {
	deps := g.buildDependencies()
	content := fmt.Sprintf(`module %s

go %s

require (
%s
)
`, g.config.ModulePath, g.config.GoVersion, strings.Join(deps, "\n"))

	return g.writeFile("go.mod", content)
}

func (g *Generator) buildDependencies() []string {
	deps := []string{}

	switch g.config.Framework {
	case "chi":
		deps = append(deps, "\tgithub.com/go-chi/chi/v5 v5.0.11")
	case "gin":
		deps = append(deps, "\tgithub.com/gin-gonic/gin v1.10.0")
	case "echo":
		deps = append(deps, "\tgithub.com/labstack/echo/v4 v4.11.4")
	case "fiber":
		deps = append(deps, "\tgithub.com/gofiber/fiber/v2 v2.52.0")
	}

	if g.config.Logger == "zap" {
		deps = append(deps, "\tgo.uber.org/zap v1.26.0")
	} else if g.config.Logger == "zerolog" {
		deps = append(deps, "\tgithub.com/rs/zerolog v1.32.0")
	}

	if g.config.HasDatabase("postgres") {
		deps = append(deps, "\tgithub.com/jackc/pgx/v5 v5.5.1")
	}
	if g.config.HasDatabase("mysql") {
		deps = append(deps, "\tgithub.com/go-sql-driver/mysql v1.7.1")
	}
	if g.config.HasDatabase("mongodb") {
		deps = append(deps, "\tgo.mongodb.org/mongo-driver v1.13.1")
	}
	if g.config.HasDatabase("redis") {
		deps = append(deps, "\tgithub.com/redis/go-redis/v9 v9.4.0")
	}

	if g.config.EnableTracing {
		deps = append(deps,
			"\tgo.opentelemetry.io/otel v1.22.0",
			"\tgo.opentelemetry.io/otel/sdk v1.22.0",
			"\tgo.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.22.0",
		)
	}

	if g.config.EnableMetrics {
		deps = append(deps, "\tgithub.com/prometheus/client_golang v1.18.0")
	}

	// Configuration file format dependencies
	switch g.config.ConfigFormat {
	case "yaml":
		deps = append(deps, "\tgopkg.in/yaml.v3 v3.0.1")
	case "toml":
		deps = append(deps, "\tgithub.com/BurntSushi/toml v1.3.2")
	case "env":
		deps = append(deps, "\tgithub.com/joho/godotenv v1.5.1")
	default:
		deps = append(deps, "\tgithub.com/joho/godotenv v1.5.1")
	}

	// Testing dependencies
	deps = append(deps,
		"\tgithub.com/stretchr/testify v1.8.4",
		"\tgo.uber.org/mock v0.4.0",
	)

	return deps
}

func (g *Generator) generateMainFile() error {
	// Prepare code snippets for template substitution
	// loggerInit: initialization code for the logger, injected into main()
	// portRef: reference to the port field in config, used in logger.Info
	loggerInit := g.getLoggerInitCode()
	portRef := g.getConfigFieldReference("Port")

	content := fmt.Sprintf(`package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"%s/internal/config"
	"%s/internal/observability"
	"%s/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %%v\n", err)
		os.Exit(1)
	}

%s

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obs, err := observability.New(ctx, cfg)
	if err != nil {
		logger.Error("Failed to initialize observability", "error", err)
		os.Exit(1)
	}
	defer obs.Shutdown(ctx)

	srv, err := server.New(cfg, obs)
	if err != nil {
		logger.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("Starting server", "port", %s)
		if err := srv.Start(); err != nil {
			logger.Error("Server error", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutting down server...")
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	logger.Info("Server stopped gracefully")
}
`, g.config.ModulePath, g.config.ModulePath, g.config.ModulePath, loggerInit, portRef)

	return g.writeFile(fmt.Sprintf("cmd/%s/main.go", g.config.ProjectName), content)
}

func (g *Generator) getLoggerInitCode() string {
	switch g.config.Logger {
	case "slog":
		return `	logger := observability.NewLogger(cfg)
	observability.SetDefaultLogger(logger)`
	case "zap":
		return `	logger, err := observability.NewZapLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()`
	case "zerolog":
		return `	logger := observability.NewZerologLogger(cfg)`
	default:
		return `	logger := observability.NewLogger(cfg)
	observability.SetDefaultLogger(logger)`
	}
}

func (g *Generator) generateConfigPackage() error {
	content := fmt.Sprintf(`package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	
%s
%s
%s
%s
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
	}

%s

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}
`, g.getDatabaseConfigFields(), g.getCacheConfigFields(), g.getTracingConfigFields(), g.getMetricsConfigFields(), g.getConfigLoadStatements())

	return g.writeFile("internal/config/config.go", content)
}

func (g *Generator) getDatabaseConfigFields() string {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return ""
	}

	fields := []string{}
	if g.config.HasDatabase("postgres") {
		fields = append(fields, "\tPostgresURL string")
	}
	if g.config.HasDatabase("mysql") {
		fields = append(fields, "\tMySQLURL string")
	}
	if g.config.HasDatabase("mongodb") {
		fields = append(fields, "\tMongoURL string")
	}

	if len(fields) > 0 {
		return strings.Join(fields, "\n")
	}
	return ""
}

func (g *Generator) getCacheConfigFields() string {
	if g.config.HasDatabase("redis") {
		return "\tRedisURL string"
	}
	return ""
}

func (g *Generator) getTracingConfigFields() string {
	if g.config.EnableTracing {
		return `	OTLPEndpoint string
	ServiceName  string`
	}
	return ""
}

func (g *Generator) getMetricsConfigFields() string {
	if g.config.EnableMetrics {
		return "\tMetricsEnabled bool"
	}
	return ""
}

func (g *Generator) getConfigLoadStatements() string {
	statements := []string{}

	if g.config.HasDatabase("postgres") {
		statements = append(statements, `	cfg.PostgresURL = getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/dbname?sslmode=disable")`)
	}
	if g.config.HasDatabase("mysql") {
		statements = append(statements, `	cfg.MySQLURL = getEnv("MYSQL_URL", "user:password@tcp(localhost:3306)/dbname")`)
	}
	if g.config.HasDatabase("mongodb") {
		statements = append(statements, `	cfg.MongoURL = getEnv("MONGO_URL", "mongodb://localhost:27017")`)
	}
	if g.config.HasDatabase("redis") {
		statements = append(statements, `	cfg.RedisURL = getEnv("REDIS_URL", "redis://localhost:6379")`)
	}
	if g.config.EnableTracing {
		statements = append(statements,
			`	cfg.OTLPEndpoint = getEnv("OTLP_ENDPOINT", "localhost:4317")`,
			fmt.Sprintf(`	cfg.ServiceName = getEnv("SERVICE_NAME", "%s")`, g.config.ProjectName),
		)
	}
	if g.config.EnableMetrics {
		statements = append(statements, `	cfg.MetricsEnabled = getEnvBool("METRICS_ENABLED", true)`)
	}

	return strings.Join(statements, "\n")
}
