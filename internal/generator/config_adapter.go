package generator

import (
	"strings"
)

// Config accessor methods generation
// This file generates helper methods for the Config struct to provide
// a consistent API regardless of the config format (env, yaml, json, toml)

func (g *Generator) generateConfigAccessors() string {
	if g.config.ConfigFormat == "" || g.config.ConfigFormat == "env" {
		// Env-based config has flat structure, no accessors needed
		return ""
	}

	var sb strings.Builder

	// Port accessor - handles int to string conversion
	sb.WriteString(`
// GetPort returns the port as a string
func (c *Config) GetPort() string {
	return fmt.Sprintf("%d", c.App.Port)
}

// GetEnvironment returns the application environment
func (c *Config) GetEnvironment() string {
	return c.App.Environment
}

// GetLogLevel returns the log level
func (c *Config) GetLogLevel() string {
	return c.App.LogLevel
}
`)

	// Database accessors
	if g.config.HasDatabase("postgres") {
		sb.WriteString(`
// GetPostgresURL returns the PostgreSQL connection URL
func (c *Config) GetPostgresURL() string {
	return c.Database.Postgres.URL
}
`)
	}

	if g.config.HasDatabase("mysql") {
		sb.WriteString(`
// GetMySQLURL returns the MySQL connection URL
func (c *Config) GetMySQLURL() string {
	return c.Database.MySQL.URL
}
`)
	}

	if g.config.HasDatabase("mongodb") {
		sb.WriteString(`
// GetMongoURL returns the MongoDB connection URL
func (c *Config) GetMongoURL() string {
	return c.Database.MongoDB.URL
}

// GetMongoDatabase returns the MongoDB database name
func (c *Config) GetMongoDatabase() string {
	return c.Database.MongoDB.Database
}
`)
	}

	// Cache accessors
	if g.config.NeedsCache() {
		sb.WriteString(`
// GetRedisURL returns the Redis connection URL
func (c *Config) GetRedisURL() string {
	return c.Cache.Redis.URL
}
`)
	}

	// Observability accessors
	if g.config.EnableTracing {
		sb.WriteString(`
// GetOTLPEndpoint returns the OTLP endpoint for tracing
func (c *Config) GetOTLPEndpoint() string {
	return c.Observability.Tracing.OTLPEndpoint
}

// GetServiceName returns the service name for tracing
func (c *Config) GetServiceName() string {
	return c.Observability.Tracing.ServiceName
}

// IsTracingEnabled returns whether tracing is enabled
func (c *Config) IsTracingEnabled() bool {
	return c.Observability.Tracing.Enabled
}
`)
	}

	if g.config.EnableMetrics {
		sb.WriteString(`
// IsMetricsEnabled returns whether metrics are enabled
func (c *Config) IsMetricsEnabled() bool {
	return c.Observability.Metrics.Enabled
}

// GetMetricsPath returns the metrics endpoint path
func (c *Config) GetMetricsPath() string {
	return c.Observability.Metrics.Path
}
`)
	}

	return sb.String()
}

// getConfigFieldReference returns the correct way to reference a config field
// based on the config format being used
func (g *Generator) getConfigFieldReference(field string) string {
	if g.config.ConfigFormat == "" || g.config.ConfigFormat == "env" {
		return "cfg." + field
	}

	// For structured configs (YAML, JSON, TOML), use accessor methods
	switch field {
	case "Port":
		return "cfg.GetPort()"
	case "Environment":
		return "cfg.GetEnvironment()"
	case "LogLevel":
		return "cfg.GetLogLevel()"
	case "PostgresURL":
		return "cfg.GetPostgresURL()"
	case "MySQLURL":
		return "cfg.GetMySQLURL()"
	case "MongoURL":
		return "cfg.GetMongoURL()"
	case "RedisURL":
		return "cfg.GetRedisURL()"
	case "OTLPEndpoint":
		return "cfg.GetOTLPEndpoint()"
	case "ServiceName":
		return "cfg.GetServiceName()"
	default:
		return "cfg." + field
	}
}
