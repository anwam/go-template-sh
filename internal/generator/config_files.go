package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateConfigFiles() error {
	switch g.config.ConfigFormat {
	case "yaml":
		if err := g.generateYAMLConfig(); err != nil {
			return err
		}
		if err := g.generateYAMLConfigLoader(); err != nil {
			return err
		}
	case "json":
		if err := g.generateJSONConfig(); err != nil {
			return err
		}
		if err := g.generateJSONConfigLoader(); err != nil {
			return err
		}
	case "toml":
		if err := g.generateTOMLConfig(); err != nil {
			return err
		}
		if err := g.generateTOMLConfigLoader(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateYAMLConfig() error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# %s Configuration
# ============================================

# Application settings
app:
  name: %s
  environment: development  # development, staging, production
  port: 8080
  log_level: info  # debug, info, warn, error

`, g.config.ProjectName, g.config.ProjectName))

	// Database configuration
	if g.config.HasDatabase("postgres") || g.config.HasDatabase("mysql") || g.config.HasDatabase("mongodb") {
		sb.WriteString("# Database configuration\ndatabase:\n")

		if g.config.HasDatabase("postgres") {
			sb.WriteString(`  postgres:
    url: postgres://user:password@localhost:5432/dbname?sslmode=disable
    max_connections: 25
    max_idle_time: 5m

`)
		}
		if g.config.HasDatabase("mysql") {
			sb.WriteString(`  mysql:
    url: user:password@tcp(localhost:3306)/dbname
    max_connections: 25
    max_idle_time: 5m

`)
		}
		if g.config.HasDatabase("mongodb") {
			sb.WriteString(`  mongodb:
    url: mongodb://localhost:27017
    database: dbname
    max_pool_size: 100
    min_pool_size: 10

`)
		}
	}

	// Cache configuration
	if g.config.HasDatabase("redis") {
		sb.WriteString(`# Cache configuration
cache:
  redis:
    url: redis://localhost:6379
    pool_size: 10
    min_idle_conns: 5

`)
	}

	// Observability configuration
	if g.config.EnableTracing || g.config.EnableMetrics {
		sb.WriteString("# Observability configuration\nobservability:\n")

		if g.config.EnableTracing {
			sb.WriteString(fmt.Sprintf(`  tracing:
    enabled: true
    otlp_endpoint: localhost:4317
    service_name: %s
    sample_rate: 1.0  # 0.0 to 1.0

`, g.config.ProjectName))
		}
		if g.config.EnableMetrics {
			sb.WriteString(`  metrics:
    enabled: true
    path: /metrics

`)
		}
	}

	// Security configuration
	sb.WriteString(`# Security configuration (optional)
# security:
#   jwt:
#     secret: your-super-secret-key-change-in-production
#     expiration: 24h
#   cors:
#     allowed_origins:
#       - http://localhost:3000
#       - https://yourdomain.com
#   rate_limit: 100  # requests per minute
`)

	return g.writeFile("config.yaml.example", sb.String())
}

func (g *Generator) generateJSONConfig() error {
	var sb strings.Builder

	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf(`  "app": {
    "name": "%s",
    "environment": "development",
    "port": 8080,
    "log_level": "info"
  }`, g.config.ProjectName))

	// Database configuration
	if g.config.HasDatabase("postgres") || g.config.HasDatabase("mysql") || g.config.HasDatabase("mongodb") {
		sb.WriteString(",\n  \"database\": {")
		dbParts := []string{}

		if g.config.HasDatabase("postgres") {
			dbParts = append(dbParts, `
    "postgres": {
      "url": "postgres://user:password@localhost:5432/dbname?sslmode=disable",
      "max_connections": 25,
      "max_idle_time": "5m"
    }`)
		}
		if g.config.HasDatabase("mysql") {
			dbParts = append(dbParts, `
    "mysql": {
      "url": "user:password@tcp(localhost:3306)/dbname",
      "max_connections": 25,
      "max_idle_time": "5m"
    }`)
		}
		if g.config.HasDatabase("mongodb") {
			dbParts = append(dbParts, `
    "mongodb": {
      "url": "mongodb://localhost:27017",
      "database": "dbname",
      "max_pool_size": 100,
      "min_pool_size": 10
    }`)
		}
		sb.WriteString(strings.Join(dbParts, ","))
		sb.WriteString("\n  }")
	}

	// Cache configuration
	if g.config.HasDatabase("redis") {
		sb.WriteString(`,
  "cache": {
    "redis": {
      "url": "redis://localhost:6379",
      "pool_size": 10,
      "min_idle_conns": 5
    }
  }`)
	}

	// Observability configuration
	if g.config.EnableTracing || g.config.EnableMetrics {
		sb.WriteString(",\n  \"observability\": {")
		obsParts := []string{}

		if g.config.EnableTracing {
			obsParts = append(obsParts, fmt.Sprintf(`
    "tracing": {
      "enabled": true,
      "otlp_endpoint": "localhost:4317",
      "service_name": "%s",
      "sample_rate": 1.0
    }`, g.config.ProjectName))
		}
		if g.config.EnableMetrics {
			obsParts = append(obsParts, `
    "metrics": {
      "enabled": true,
      "path": "/metrics"
    }`)
		}
		sb.WriteString(strings.Join(obsParts, ","))
		sb.WriteString("\n  }")
	}

	sb.WriteString("\n}\n")

	return g.writeFile("config.json.example", sb.String())
}

func (g *Generator) generateTOMLConfig() error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`# %s Configuration
# ============================================

[app]
name = "%s"
environment = "development"  # development, staging, production
port = 8080
log_level = "info"  # debug, info, warn, error

`, g.config.ProjectName, g.config.ProjectName))

	// Database configuration
	if g.config.HasDatabase("postgres") {
		sb.WriteString(`[database.postgres]
url = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
max_connections = 25
max_idle_time = "5m"

`)
	}
	if g.config.HasDatabase("mysql") {
		sb.WriteString(`[database.mysql]
url = "user:password@tcp(localhost:3306)/dbname"
max_connections = 25
max_idle_time = "5m"

`)
	}
	if g.config.HasDatabase("mongodb") {
		sb.WriteString(`[database.mongodb]
url = "mongodb://localhost:27017"
database = "dbname"
max_pool_size = 100
min_pool_size = 10

`)
	}

	// Cache configuration
	if g.config.HasDatabase("redis") {
		sb.WriteString(`[cache.redis]
url = "redis://localhost:6379"
pool_size = 10
min_idle_conns = 5

`)
	}

	// Observability configuration
	if g.config.EnableTracing {
		sb.WriteString(fmt.Sprintf(`[observability.tracing]
enabled = true
otlp_endpoint = "localhost:4317"
service_name = "%s"
sample_rate = 1.0

`, g.config.ProjectName))
	}
	if g.config.EnableMetrics {
		sb.WriteString(`[observability.metrics]
enabled = true
path = "/metrics"

`)
	}

	// Security configuration
	sb.WriteString(`# Security configuration (optional)
# [security.jwt]
# secret = "your-super-secret-key-change-in-production"
# expiration = "24h"

# [security.cors]
# allowed_origins = ["http://localhost:3000", "https://yourdomain.com"]

# [security]
# rate_limit = 100  # requests per minute
`)

	return g.writeFile("config.toml.example", sb.String())
}

func (g *Generator) generateYAMLConfigLoader() error {
	content := fmt.Sprintf(`package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App           AppConfig           `+"`yaml:\"app\"`"+`
%s%s%s
}

type AppConfig struct {
	Name        string `+"`yaml:\"name\"`"+`
	Environment string `+"`yaml:\"environment\"`"+`
	Port        int    `+"`yaml:\"port\"`"+`
	LogLevel    string `+"`yaml:\"log_level\"`"+`
}

%s%s%s

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %%w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %%w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return cfg, cfg.validate()
}

func (c *Config) applyEnvOverrides() {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		c.App.Environment = env
	}
	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%%d", &c.App.Port)
	}
}

func (c *Config) validate() error {
	if c.App.Port == 0 {
		return fmt.Errorf("app.port is required")
	}
	return nil
}
`, g.getYAMLDatabaseConfigField(), g.getYAMLCacheConfigField(), g.getYAMLObservabilityConfigField(),
		g.getYAMLDatabaseConfigTypes(), g.getYAMLCacheConfigTypes(), g.getYAMLObservabilityConfigTypes())

	return g.writeFile("internal/config/config.go", content)
}

func (g *Generator) getYAMLDatabaseConfigField() string {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return ""
	}
	return "\tDatabase      DatabaseConfig      `yaml:\"database\"`\n"
}

func (g *Generator) getYAMLCacheConfigField() string {
	if !g.config.NeedsCache() {
		return ""
	}
	return "\tCache         CacheConfig         `yaml:\"cache\"`\n"
}

func (g *Generator) getYAMLObservabilityConfigField() string {
	if !g.config.EnableTracing && !g.config.EnableMetrics {
		return ""
	}
	return "\tObservability ObservabilityConfig `yaml:\"observability\"`\n"
}

func (g *Generator) getYAMLDatabaseConfigTypes() string {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("type DatabaseConfig struct {\n")
	if g.config.HasDatabase("postgres") {
		sb.WriteString("\tPostgres PostgresConfig `yaml:\"postgres\"`\n")
	}
	if g.config.HasDatabase("mysql") {
		sb.WriteString("\tMySQL    MySQLConfig    `yaml:\"mysql\"`\n")
	}
	if g.config.HasDatabase("mongodb") {
		sb.WriteString("\tMongoDB  MongoDBConfig  `yaml:\"mongodb\"`\n")
	}
	sb.WriteString("}\n\n")

	if g.config.HasDatabase("postgres") {
		sb.WriteString(`type PostgresConfig struct {
	URL            string        ` + "`yaml:\"url\"`" + `
	MaxConnections int           ` + "`yaml:\"max_connections\"`" + `
	MaxIdleTime    time.Duration ` + "`yaml:\"max_idle_time\"`" + `
}

`)
	}
	if g.config.HasDatabase("mysql") {
		sb.WriteString(`type MySQLConfig struct {
	URL            string        ` + "`yaml:\"url\"`" + `
	MaxConnections int           ` + "`yaml:\"max_connections\"`" + `
	MaxIdleTime    time.Duration ` + "`yaml:\"max_idle_time\"`" + `
}

`)
	}
	if g.config.HasDatabase("mongodb") {
		sb.WriteString(`type MongoDBConfig struct {
	URL         string ` + "`yaml:\"url\"`" + `
	Database    string ` + "`yaml:\"database\"`" + `
	MaxPoolSize int    ` + "`yaml:\"max_pool_size\"`" + `
	MinPoolSize int    ` + "`yaml:\"min_pool_size\"`" + `
}

`)
	}

	return sb.String()
}

func (g *Generator) getYAMLCacheConfigTypes() string {
	if !g.config.NeedsCache() {
		return ""
	}

	return `type CacheConfig struct {
	Redis RedisConfig ` + "`yaml:\"redis\"`" + `
}

type RedisConfig struct {
	URL          string ` + "`yaml:\"url\"`" + `
	PoolSize     int    ` + "`yaml:\"pool_size\"`" + `
	MinIdleConns int    ` + "`yaml:\"min_idle_conns\"`" + `
}

`
}

func (g *Generator) getYAMLObservabilityConfigTypes() string {
	if !g.config.EnableTracing && !g.config.EnableMetrics {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("type ObservabilityConfig struct {\n")
	if g.config.EnableTracing {
		sb.WriteString("\tTracing TracingConfig `yaml:\"tracing\"`\n")
	}
	if g.config.EnableMetrics {
		sb.WriteString("\tMetrics MetricsConfig `yaml:\"metrics\"`\n")
	}
	sb.WriteString("}\n\n")

	if g.config.EnableTracing {
		sb.WriteString(`type TracingConfig struct {
	Enabled      bool    ` + "`yaml:\"enabled\"`" + `
	OTLPEndpoint string  ` + "`yaml:\"otlp_endpoint\"`" + `
	ServiceName  string  ` + "`yaml:\"service_name\"`" + `
	SampleRate   float64 ` + "`yaml:\"sample_rate\"`" + `
}

`)
	}
	if g.config.EnableMetrics {
		sb.WriteString(`type MetricsConfig struct {
	Enabled bool   ` + "`yaml:\"enabled\"`" + `
	Path    string ` + "`yaml:\"path\"`" + `
}

`)
	}

	return sb.String()
}

func (g *Generator) generateJSONConfigLoader() error {
	content := fmt.Sprintf(`package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	App           AppConfig           `+"`json:\"app\"`"+`
%s%s%s
}

type AppConfig struct {
	Name        string `+"`json:\"name\"`"+`
	Environment string `+"`json:\"environment\"`"+`
	Port        int    `+"`json:\"port\"`"+`
	LogLevel    string `+"`json:\"log_level\"`"+`
}

%s%s%s

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %%w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %%w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return cfg, cfg.validate()
}

func (c *Config) applyEnvOverrides() {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		c.App.Environment = env
	}
	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%%d", &c.App.Port)
	}
}

func (c *Config) validate() error {
	if c.App.Port == 0 {
		return fmt.Errorf("app.port is required")
	}
	return nil
}
`, g.getJSONDatabaseConfigField(), g.getJSONCacheConfigField(), g.getJSONObservabilityConfigField(),
		g.getJSONDatabaseConfigTypes(), g.getJSONCacheConfigTypes(), g.getJSONObservabilityConfigTypes())

	return g.writeFile("internal/config/config.go", content)
}

func (g *Generator) getJSONDatabaseConfigField() string {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return ""
	}
	return "\tDatabase      DatabaseConfig      `json:\"database\"`\n"
}

func (g *Generator) getJSONCacheConfigField() string {
	if !g.config.NeedsCache() {
		return ""
	}
	return "\tCache         CacheConfig         `json:\"cache\"`\n"
}

func (g *Generator) getJSONObservabilityConfigField() string {
	if !g.config.EnableTracing && !g.config.EnableMetrics {
		return ""
	}
	return "\tObservability ObservabilityConfig `json:\"observability\"`\n"
}

func (g *Generator) getJSONDatabaseConfigTypes() string {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("type DatabaseConfig struct {\n")
	if g.config.HasDatabase("postgres") {
		sb.WriteString("\tPostgres PostgresConfig `json:\"postgres\"`\n")
	}
	if g.config.HasDatabase("mysql") {
		sb.WriteString("\tMySQL    MySQLConfig    `json:\"mysql\"`\n")
	}
	if g.config.HasDatabase("mongodb") {
		sb.WriteString("\tMongoDB  MongoDBConfig  `json:\"mongodb\"`\n")
	}
	sb.WriteString("}\n\n")

	if g.config.HasDatabase("postgres") {
		sb.WriteString(`type PostgresConfig struct {
	URL            string ` + "`json:\"url\"`" + `
	MaxConnections int    ` + "`json:\"max_connections\"`" + `
	MaxIdleTime    string ` + "`json:\"max_idle_time\"`" + `
}

func (p *PostgresConfig) GetMaxIdleTime() time.Duration {
	d, _ := time.ParseDuration(p.MaxIdleTime)
	return d
}

`)
	}
	if g.config.HasDatabase("mysql") {
		sb.WriteString(`type MySQLConfig struct {
	URL            string ` + "`json:\"url\"`" + `
	MaxConnections int    ` + "`json:\"max_connections\"`" + `
	MaxIdleTime    string ` + "`json:\"max_idle_time\"`" + `
}

func (m *MySQLConfig) GetMaxIdleTime() time.Duration {
	d, _ := time.ParseDuration(m.MaxIdleTime)
	return d
}

`)
	}
	if g.config.HasDatabase("mongodb") {
		sb.WriteString(`type MongoDBConfig struct {
	URL         string ` + "`json:\"url\"`" + `
	Database    string ` + "`json:\"database\"`" + `
	MaxPoolSize int    ` + "`json:\"max_pool_size\"`" + `
	MinPoolSize int    ` + "`json:\"min_pool_size\"`" + `
}

`)
	}

	return sb.String()
}

func (g *Generator) getJSONCacheConfigTypes() string {
	if !g.config.NeedsCache() {
		return ""
	}

	return `type CacheConfig struct {
	Redis RedisConfig ` + "`json:\"redis\"`" + `
}

type RedisConfig struct {
	URL          string ` + "`json:\"url\"`" + `
	PoolSize     int    ` + "`json:\"pool_size\"`" + `
	MinIdleConns int    ` + "`json:\"min_idle_conns\"`" + `
}

`
}

func (g *Generator) getJSONObservabilityConfigTypes() string {
	if !g.config.EnableTracing && !g.config.EnableMetrics {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("type ObservabilityConfig struct {\n")
	if g.config.EnableTracing {
		sb.WriteString("\tTracing TracingConfig `json:\"tracing\"`\n")
	}
	if g.config.EnableMetrics {
		sb.WriteString("\tMetrics MetricsConfig `json:\"metrics\"`\n")
	}
	sb.WriteString("}\n\n")

	if g.config.EnableTracing {
		sb.WriteString(`type TracingConfig struct {
	Enabled      bool    ` + "`json:\"enabled\"`" + `
	OTLPEndpoint string  ` + "`json:\"otlp_endpoint\"`" + `
	ServiceName  string  ` + "`json:\"service_name\"`" + `
	SampleRate   float64 ` + "`json:\"sample_rate\"`" + `
}

`)
	}
	if g.config.EnableMetrics {
		sb.WriteString(`type MetricsConfig struct {
	Enabled bool   ` + "`json:\"enabled\"`" + `
	Path    string ` + "`json:\"path\"`" + `
}

`)
	}

	return sb.String()
}

func (g *Generator) generateTOMLConfigLoader() error {
	content := fmt.Sprintf(`package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	App           AppConfig           `+"`toml:\"app\"`"+`
%s%s%s
}

type AppConfig struct {
	Name        string `+"`toml:\"name\"`"+`
	Environment string `+"`toml:\"environment\"`"+`
	Port        int    `+"`toml:\"port\"`"+`
	LogLevel    string `+"`toml:\"log_level\"`"+`
}

%s%s%s

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.toml"
	}

	cfg := &Config{}
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %%w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return cfg, cfg.validate()
}

func (c *Config) applyEnvOverrides() {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		c.App.Environment = env
	}
	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%%d", &c.App.Port)
	}
}

func (c *Config) validate() error {
	if c.App.Port == 0 {
		return fmt.Errorf("app.port is required")
	}
	return nil
}
`, g.getTOMLDatabaseConfigField(), g.getTOMLCacheConfigField(), g.getTOMLObservabilityConfigField(),
		g.getTOMLDatabaseConfigTypes(), g.getTOMLCacheConfigTypes(), g.getTOMLObservabilityConfigTypes())

	return g.writeFile("internal/config/config.go", content)
}

func (g *Generator) getTOMLDatabaseConfigField() string {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return ""
	}
	return "\tDatabase      DatabaseConfig      `toml:\"database\"`\n"
}

func (g *Generator) getTOMLCacheConfigField() string {
	if !g.config.NeedsCache() {
		return ""
	}
	return "\tCache         CacheConfig         `toml:\"cache\"`\n"
}

func (g *Generator) getTOMLObservabilityConfigField() string {
	if !g.config.EnableTracing && !g.config.EnableMetrics {
		return ""
	}
	return "\tObservability ObservabilityConfig `toml:\"observability\"`\n"
}

func (g *Generator) getTOMLDatabaseConfigTypes() string {
	if !g.config.NeedsSQL() && !g.config.NeedsNoSQL() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("type DatabaseConfig struct {\n")
	if g.config.HasDatabase("postgres") {
		sb.WriteString("\tPostgres PostgresConfig `toml:\"postgres\"`\n")
	}
	if g.config.HasDatabase("mysql") {
		sb.WriteString("\tMySQL    MySQLConfig    `toml:\"mysql\"`\n")
	}
	if g.config.HasDatabase("mongodb") {
		sb.WriteString("\tMongoDB  MongoDBConfig  `toml:\"mongodb\"`\n")
	}
	sb.WriteString("}\n\n")

	if g.config.HasDatabase("postgres") {
		sb.WriteString(`type PostgresConfig struct {
	URL            string        ` + "`toml:\"url\"`" + `
	MaxConnections int           ` + "`toml:\"max_connections\"`" + `
	MaxIdleTime    time.Duration ` + "`toml:\"max_idle_time\"`" + `
}

`)
	}
	if g.config.HasDatabase("mysql") {
		sb.WriteString(`type MySQLConfig struct {
	URL            string        ` + "`toml:\"url\"`" + `
	MaxConnections int           ` + "`toml:\"max_connections\"`" + `
	MaxIdleTime    time.Duration ` + "`toml:\"max_idle_time\"`" + `
}

`)
	}
	if g.config.HasDatabase("mongodb") {
		sb.WriteString(`type MongoDBConfig struct {
	URL         string ` + "`toml:\"url\"`" + `
	Database    string ` + "`toml:\"database\"`" + `
	MaxPoolSize int    ` + "`toml:\"max_pool_size\"`" + `
	MinPoolSize int    ` + "`toml:\"min_pool_size\"`" + `
}

`)
	}

	return sb.String()
}

func (g *Generator) getTOMLCacheConfigTypes() string {
	if !g.config.NeedsCache() {
		return ""
	}

	return `type CacheConfig struct {
	Redis RedisConfig ` + "`toml:\"redis\"`" + `
}

type RedisConfig struct {
	URL          string ` + "`toml:\"url\"`" + `
	PoolSize     int    ` + "`toml:\"pool_size\"`" + `
	MinIdleConns int    ` + "`toml:\"min_idle_conns\"`" + `
}

`
}

func (g *Generator) getTOMLObservabilityConfigTypes() string {
	if !g.config.EnableTracing && !g.config.EnableMetrics {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("type ObservabilityConfig struct {\n")
	if g.config.EnableTracing {
		sb.WriteString("\tTracing TracingConfig `toml:\"tracing\"`\n")
	}
	if g.config.EnableMetrics {
		sb.WriteString("\tMetrics MetricsConfig `toml:\"metrics\"`\n")
	}
	sb.WriteString("}\n\n")

	if g.config.EnableTracing {
		sb.WriteString(`type TracingConfig struct {
	Enabled      bool    ` + "`toml:\"enabled\"`" + `
	OTLPEndpoint string  ` + "`toml:\"otlp_endpoint\"`" + `
	ServiceName  string  ` + "`toml:\"service_name\"`" + `
	SampleRate   float64 ` + "`toml:\"sample_rate\"`" + `
}

`)
	}
	if g.config.EnableMetrics {
		sb.WriteString(`type MetricsConfig struct {
	Enabled bool   ` + "`toml:\"enabled\"`" + `
	Path    string ` + "`toml:\"path\"`" + `
}

`)
	}

	return sb.String()
}
