package generator

import (
"os"
"path/filepath"
"strings"
"testing"

"github.com/anwam/go-template-sh/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test-project",
		ModulePath:  "github.com/test/test-project",
		GoVersion:   "1.23",
	}

	gen := New(cfg, "/tmp/output")

	if gen.config != cfg {
		t.Error("Expected config to be set")
	}
	if gen.outputDir != "/tmp/output" {
		t.Errorf("Expected outputDir to be /tmp/output, got %s", gen.outputDir)
	}
	if gen.projectDir != "/tmp/output/test-project" {
		t.Errorf("Expected projectDir to be /tmp/output/test-project, got %s", gen.projectDir)
	}
}

func TestGenerator_Generate_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "",
		ModulePath:  "github.com/test/test-project",
		GoVersion:   "1.23",
	}

	gen := New(cfg, t.TempDir())
	err := gen.Generate()

	if err == nil {
		t.Error("Expected error for invalid config")
	}
	if !strings.Contains(err.Error(), "invalid configuration") {
		t.Errorf("Expected 'invalid configuration' error, got: %v", err)
	}
}

func TestGenerator_Generate_BasicProject(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "test-basic",
		ModulePath:    "github.com/test/test-basic",
		GoVersion:     "1.23",
		Framework:     "stdlib",
		Logger:        "slog",
		EnableTracing: false,
		EnableMetrics: false,
		IncludeDocker: false,
		ConfigFormat:  "env",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedFiles := []string{
		"go.mod",
		"cmd/test-basic/main.go",
		"internal/config/config.go",
		"internal/server/server.go",
		"internal/handlers/handlers.go",
		"internal/middleware/middleware.go",
		"internal/observability/observability.go",
		"Makefile",
		"README.md",
		".gitignore",
	}

	projectDir := filepath.Join(outputDir, "test-basic")
	for _, file := range expectedFiles {
		path := filepath.Join(projectDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}
}

func TestGenerator_Generate_WithDocker(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "test-docker",
		ModulePath:    "github.com/test/test-docker",
		GoVersion:     "1.23",
		Framework:     "chi",
		Logger:        "slog",
		IncludeDocker: true,
		ConfigFormat:  "env",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedFiles := []string{
		"Dockerfile",
		"docker-compose.yml",
		".dockerignore",
	}

	projectDir := filepath.Join(outputDir, "test-docker")
	for _, file := range expectedFiles {
		path := filepath.Join(projectDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected Docker file %s to exist", file)
		}
	}
}

func TestGenerator_Generate_WithDatabases(t *testing.T) {
	cfg := &config.Config{
		ProjectName:  "test-db",
		ModulePath:   "github.com/test/test-db",
		GoVersion:    "1.23",
		Framework:    "stdlib",
		Logger:       "slog",
		Databases:    []string{"postgres", "redis"},
		ConfigFormat: "env",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedFiles := []string{
		"internal/database/postgres.go",
		"internal/cache/redis.go",
	}

	projectDir := filepath.Join(outputDir, "test-db")
	for _, file := range expectedFiles {
		path := filepath.Join(projectDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected database file %s to exist", file)
		}
	}
}

func TestGenerator_buildDependencies(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected []string
	}{
		{
			name: "chi framework",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Framework:    "chi",
				ConfigFormat: "env",
			},
			expected: []string{"github.com/go-chi/chi/v5"},
		},
		{
			name: "zap logger",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Logger:       "zap",
				ConfigFormat: "env",
			},
			expected: []string{"go.uber.org/zap"},
		},
		{
			name: "postgres database",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Databases:    []string{"postgres"},
				ConfigFormat: "env",
			},
			expected: []string{"github.com/jackc/pgx/v5"},
		},
		{
			name: "uuid always included",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				ConfigFormat: "env",
			},
			expected: []string{"github.com/google/uuid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
gen := New(tt.config, "/tmp")
deps := gen.buildDependencies()
			depsStr := strings.Join(deps, "\n")

			for _, exp := range tt.expected {
				if !strings.Contains(depsStr, exp) {
					t.Errorf("Expected dependency %s not found in: %v", exp, deps)
				}
			}
		})
	}
}

func TestGenerator_getConfigFieldReference(t *testing.T) {
	tests := []struct {
		name         string
		configFormat string
		field        string
		expected     string
	}{
		{
			name:         "env format - Port",
			configFormat: "env",
			field:        "Port",
			expected:     "cfg.Port",
		},
		{
			name:         "yaml format - Port",
			configFormat: "yaml",
			field:        "Port",
			expected:     "cfg.GetPort()",
		},
		{
			name:         "empty format defaults to env",
			configFormat: "",
			field:        "Environment",
			expected:     "cfg.Environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
cfg := &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				ConfigFormat: tt.configFormat,
			}
			gen := New(cfg, "/tmp")

			result := gen.getConfigFieldReference(tt.field)
			if result != tt.expected {
				t.Errorf("getConfigFieldReference(%s) = %s, want %s", tt.field, result, tt.expected)
			}
		})
	}
}

func TestGenerator_Generate_WithGitHubActions(t *testing.T) {
	cfg := &config.Config{
		ProjectName:  "test-ci",
		ModulePath:   "github.com/test/test-ci",
		GoVersion:    "1.23",
		Framework:    "stdlib",
		Logger:       "slog",
		CI:           "github",
		ConfigFormat: "env",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	ciPath := filepath.Join(outputDir, "test-ci", ".github", "workflows", "ci.yml")
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		t.Error("Expected GitHub Actions CI file to exist")
	}
}

func TestGenerator_Generate_WithGitLabCI(t *testing.T) {
	cfg := &config.Config{
		ProjectName:  "test-gitlab",
		ModulePath:   "github.com/test/test-gitlab",
		GoVersion:    "1.23",
		Framework:    "stdlib",
		Logger:       "slog",
		CI:           "gitlab",
		ConfigFormat: "env",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	ciPath := filepath.Join(outputDir, "test-gitlab", ".gitlab-ci.yml")
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		t.Error("Expected GitLab CI file to exist")
	}
}

func TestGenerator_Generate_WithYAMLConfig(t *testing.T) {
	cfg := &config.Config{
		ProjectName:  "test-yaml",
		ModulePath:   "github.com/test/test-yaml",
		GoVersion:    "1.23",
		Framework:    "chi",
		Logger:       "slog",
		ConfigFormat: "yaml",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	configPath := filepath.Join(outputDir, "test-yaml", "config.yaml.example")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected YAML config example file to exist")
	}
}

func TestGenerator_Generate_WithJSONConfig(t *testing.T) {
	cfg := &config.Config{
		ProjectName:  "test-json",
		ModulePath:   "github.com/test/test-json",
		GoVersion:    "1.23",
		Framework:    "stdlib",
		Logger:       "slog",
		ConfigFormat: "json",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	configPath := filepath.Join(outputDir, "test-json", "config.json.example")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected JSON config example file to exist")
	}
}

func TestGenerator_Generate_WithTOMLConfig(t *testing.T) {
	cfg := &config.Config{
		ProjectName:  "test-toml",
		ModulePath:   "github.com/test/test-toml",
		GoVersion:    "1.23",
		Framework:    "stdlib",
		Logger:       "slog",
		ConfigFormat: "toml",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	configPath := filepath.Join(outputDir, "test-toml", "config.toml.example")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected TOML config example file to exist")
	}
}

func TestGenerator_Generate_WithObservability(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "test-obs",
		ModulePath:    "github.com/test/test-obs",
		GoVersion:     "1.23",
		Framework:     "stdlib",
		Logger:        "slog",
		EnableTracing: true,
		EnableMetrics: true,
		ConfigFormat:  "env",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check go.mod contains observability deps
	goModPath := filepath.Join(outputDir, "test-obs", "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	expectedDeps := []string{
		"go.opentelemetry.io/otel",
		"github.com/prometheus/client_golang",
	}

	for _, dep := range expectedDeps {
		if !strings.Contains(string(content), dep) {
			t.Errorf("Expected go.mod to contain %s", dep)
		}
	}
}

func TestGenerator_Generate_AllFrameworks(t *testing.T) {
	frameworks := []string{"stdlib", "chi", "gin", "echo", "fiber"}

	for _, framework := range frameworks {
		t.Run(framework, func(t *testing.T) {
cfg := &config.Config{
				ProjectName:  "test-" + framework,
				ModulePath:   "github.com/test/test-" + framework,
				GoVersion:    "1.23",
				Framework:    framework,
				Logger:       "slog",
				ConfigFormat: "env",
			}

			outputDir := t.TempDir()
			gen := New(cfg, outputDir)

			if err := gen.Generate(); err != nil {
				t.Fatalf("Generate failed for %s: %v", framework, err)
			}

			serverPath := filepath.Join(outputDir, "test-"+framework, "internal", "server", "server.go")
			if _, err := os.Stat(serverPath); os.IsNotExist(err) {
				t.Errorf("Expected server.go to exist for framework %s", framework)
			}
		})
	}
}

func TestGenerator_Generate_AllLoggers(t *testing.T) {
	loggers := []string{"slog", "zap", "zerolog"}

	for _, logger := range loggers {
		t.Run(logger, func(t *testing.T) {
cfg := &config.Config{
				ProjectName:  "test-" + logger,
				ModulePath:   "github.com/test/test-" + logger,
				GoVersion:    "1.23",
				Framework:    "stdlib",
				Logger:       logger,
				ConfigFormat: "env",
			}

			outputDir := t.TempDir()
			gen := New(cfg, outputDir)

			if err := gen.Generate(); err != nil {
				t.Fatalf("Generate failed for logger %s: %v", logger, err)
			}

			loggerPath := filepath.Join(outputDir, "test-"+logger, "internal", "observability", "logger.go")
			if _, err := os.Stat(loggerPath); os.IsNotExist(err) {
				t.Errorf("Expected logger.go to exist for logger %s", logger)
			}
		})
	}
}

func TestGenerator_Generate_AllDatabases(t *testing.T) {
	tests := []struct {
		name     string
		database string
		file     string
	}{
		{"postgres", "postgres", "internal/database/postgres.go"},
		{"mysql", "mysql", "internal/database/mysql.go"},
		{"mongodb", "mongodb", "internal/database/mongodb.go"},
		{"redis", "redis", "internal/cache/redis.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
cfg := &config.Config{
				ProjectName:  "test-" + tt.database,
				ModulePath:   "github.com/test/test-" + tt.database,
				GoVersion:    "1.23",
				Framework:    "stdlib",
				Logger:       "slog",
				Databases:    []string{tt.database},
				ConfigFormat: "env",
			}

			outputDir := t.TempDir()
			gen := New(cfg, outputDir)

			if err := gen.Generate(); err != nil {
				t.Fatalf("Generate failed for database %s: %v", tt.database, err)
			}

			dbPath := filepath.Join(outputDir, "test-"+tt.database, tt.file)
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				t.Errorf("Expected %s to exist for database %s", tt.file, tt.database)
			}
		})
	}
}

func TestGenerator_buildDependencies_AllOptions(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected []string
	}{
		{
			name: "gin framework",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Framework:    "gin",
				ConfigFormat: "env",
			},
			expected: []string{"github.com/gin-gonic/gin"},
		},
		{
			name: "echo framework",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Framework:    "echo",
				ConfigFormat: "env",
			},
			expected: []string{"github.com/labstack/echo"},
		},
		{
			name: "fiber framework",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Framework:    "fiber",
				ConfigFormat: "env",
			},
			expected: []string{"github.com/gofiber/fiber"},
		},
		{
			name: "zerolog logger",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Logger:       "zerolog",
				ConfigFormat: "env",
			},
			expected: []string{"github.com/rs/zerolog"},
		},
		{
			name: "mysql database",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Databases:    []string{"mysql"},
				ConfigFormat: "env",
			},
			expected: []string{"github.com/go-sql-driver/mysql"},
		},
		{
			name: "mongodb database",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Databases:    []string{"mongodb"},
				ConfigFormat: "env",
			},
			expected: []string{"go.mongodb.org/mongo-driver"},
		},
		{
			name: "redis cache",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				Databases:    []string{"redis"},
				ConfigFormat: "env",
			},
			expected: []string{"github.com/redis/go-redis"},
		},
		{
			name: "yaml config format",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				ConfigFormat: "yaml",
			},
			expected: []string{"gopkg.in/yaml.v3"},
		},
		{
			name: "toml config format",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				ConfigFormat: "toml",
			},
			expected: []string{"github.com/BurntSushi/toml"},
		},
		{
			name: "tracing enabled",
			config: &config.Config{
				ProjectName:   "test",
				ModulePath:    "github.com/test/test",
				GoVersion:     "1.23",
				EnableTracing: true,
				ConfigFormat:  "env",
			},
			expected: []string{"go.opentelemetry.io/otel"},
		},
		{
			name: "metrics enabled",
			config: &config.Config{
				ProjectName:   "test",
				ModulePath:    "github.com/test/test",
				GoVersion:     "1.23",
				EnableMetrics: true,
				ConfigFormat:  "env",
			},
			expected: []string{"github.com/prometheus/client_golang"},
		},
		{
			name: "testify always included",
			config: &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				ConfigFormat: "env",
			},
			expected: []string{"github.com/stretchr/testify", "go.uber.org/mock"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
gen := New(tt.config, "/tmp")
deps := gen.buildDependencies()
			depsStr := strings.Join(deps, "\n")

			for _, exp := range tt.expected {
				if !strings.Contains(depsStr, exp) {
					t.Errorf("Expected dependency %s not found", exp)
				}
			}
		})
	}
}

func TestGenerator_getConfigFieldReference_AllFields(t *testing.T) {
	tests := []struct {
		name         string
		configFormat string
		field        string
		expected     string
	}{
		{"env-LogLevel", "env", "LogLevel", "cfg.LogLevel"},
		{"yaml-LogLevel", "yaml", "LogLevel", "cfg.GetLogLevel()"},
		{"yaml-PostgresURL", "yaml", "PostgresURL", "cfg.GetPostgresURL()"},
		{"yaml-MySQLURL", "yaml", "MySQLURL", "cfg.GetMySQLURL()"},
		{"yaml-MongoURL", "yaml", "MongoURL", "cfg.GetMongoURL()"},
		{"yaml-RedisURL", "yaml", "RedisURL", "cfg.GetRedisURL()"},
		{"yaml-OTLPEndpoint", "yaml", "OTLPEndpoint", "cfg.GetOTLPEndpoint()"},
		{"yaml-ServiceName", "yaml", "ServiceName", "cfg.GetServiceName()"},
		{"json-Port", "json", "Port", "cfg.GetPort()"},
		{"toml-Environment", "toml", "Environment", "cfg.GetEnvironment()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
cfg := &config.Config{
				ProjectName:  "test",
				ModulePath:   "github.com/test/test",
				GoVersion:    "1.23",
				ConfigFormat: tt.configFormat,
			}
			gen := New(cfg, "/tmp")

			result := gen.getConfigFieldReference(tt.field)
			if result != tt.expected {
				t.Errorf("getConfigFieldReference(%s) = %s, want %s", tt.field, result, tt.expected)
			}
		})
	}
}

func TestGenerator_writeFile(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test-write",
		ModulePath:  "github.com/test/test-write",
		GoVersion:   "1.23",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	// Create project directory first
	projectDir := filepath.Join(outputDir, "test-write")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Test writing a file
	content := "test content"
	if err := gen.writeFile("test.txt", content); err != nil {
		t.Fatalf("writeFile failed: %v", err)
	}

	// Verify file exists and has correct content
	written, err := os.ReadFile(filepath.Join(projectDir, "test.txt"))
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(written) != content {
		t.Errorf("File content = %q, want %q", string(written), content)
	}
}

func TestGenerator_writeFile_CreatesSubdirectories(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "test-subdir",
		ModulePath:  "github.com/test/test-subdir",
		GoVersion:   "1.23",
	}

	outputDir := t.TempDir()
	gen := New(cfg, outputDir)

	// Create project directory first
	projectDir := filepath.Join(outputDir, "test-subdir")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Test writing a file in nested subdirectory
	if err := gen.writeFile("deep/nested/path/file.txt", "nested content"); err != nil {
		t.Fatalf("writeFile failed: %v", err)
	}

	// Verify file exists
	nestedPath := filepath.Join(projectDir, "deep", "nested", "path", "file.txt")
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Expected nested file to exist")
	}
}
