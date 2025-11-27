package test

import (
	"os"
	"strings"
	"testing"

	"github.com/anwam/go-template-sh/internal/config"
	"github.com/anwam/go-template-sh/internal/generator"
)

func TestYAMLConfigGeneration(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "test-yaml",
		ModulePath:    "github.com/test/test-yaml",
		GoVersion:     "1.21",
		Framework:     "chi",
		Databases:     []string{"postgres", "redis"},
		Logger:        "slog",
		EnableTracing: true,
		EnableMetrics: true,
		IncludeDocker: false,
		CI:            "",
		ConfigFormat:  "yaml",
		EnvSample:     false,
	}

	outputDir := t.TempDir()
	gen := generator.New(cfg, outputDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify config.go has accessor methods
	configPath := outputDir + "/test-yaml/internal/config/config.go"
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.go: %v", err)
	}

	contentStr := string(content)
	checks := map[string]bool{
		"GetPort":         strings.Contains(contentStr, "func (c *Config) GetPort() string"),
		"GetEnvironment":  strings.Contains(contentStr, "func (c *Config) GetEnvironment() string"),
		"GetPostgresURL":  strings.Contains(contentStr, "func (c *Config) GetPostgresURL() string"),
		"GetRedisURL":     strings.Contains(contentStr, "func (c *Config) GetRedisURL() string"),
		"GetOTLPEndpoint": strings.Contains(contentStr, "func (c *Config) GetOTLPEndpoint() string"),
		"GetServiceName":  strings.Contains(contentStr, "func (c *Config) GetServiceName() string"),
	}

	for method, found := range checks {
		if !found {
			t.Errorf("Config accessor method %s not found", method)
		}
	}

	// Verify database files use accessors
	postgresPath := outputDir + "/test-yaml/internal/database/postgres.go"
	postgresContent, err := os.ReadFile(postgresPath)
	if err != nil {
		t.Fatalf("Failed to read postgres.go: %v", err)
	}

	if !strings.Contains(string(postgresContent), "cfg.GetPostgresURL()") {
		t.Error("postgres.go should use cfg.GetPostgresURL()")
	}

	// Verify cache file uses accessors
	redisPath := outputDir + "/test-yaml/internal/cache/redis.go"
	redisContent, err := os.ReadFile(redisPath)
	if err != nil {
		t.Fatalf("Failed to read redis.go: %v", err)
	}

	if !strings.Contains(string(redisContent), "cfg.GetRedisURL()") {
		t.Error("redis.go should use cfg.GetRedisURL()")
	}

	// Verify server.go uses GetPort()
	serverPath := outputDir + "/test-yaml/internal/server/server.go"
	serverContent, err := os.ReadFile(serverPath)
	if err != nil {
		t.Fatalf("Failed to read server.go: %v", err)
	}

	if !strings.Contains(string(serverContent), "cfg.GetPort()") {
		t.Error("server.go should use cfg.GetPort()")
	}

	t.Log("✅ All checks passed!")
}

func TestEnvConfigGeneration(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "test-env",
		ModulePath:    "github.com/test/test-env",
		GoVersion:     "1.21",
		Framework:     "stdlib",
		Databases:     []string{"postgres"},
		Logger:        "slog",
		EnableTracing: false,
		EnableMetrics: false,
		IncludeDocker: false,
		CI:            "",
		ConfigFormat:  "env",
		EnvSample:     true,
	}

	outputDir := "/tmp/test-env-output"
	os.RemoveAll(outputDir)
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	gen := generator.New(cfg, outputDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify config.go has flat structure
	configPath := outputDir + "/test-env/internal/config/config.go"
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.go: %v", err)
	}

	contentStr := string(content)

	// Should have PostgresURL field
	if !strings.Contains(contentStr, "PostgresURL string") {
		t.Error("env config should have PostgresURL string field")
	}

	// Should NOT have accessor methods (flat structure)
	if strings.Contains(contentStr, "func (c *Config) GetPostgresURL()") {
		t.Error("env config should not have accessor methods")
	}

	// Verify database file uses direct field access
	postgresPath := outputDir + "/test-env/internal/database/postgres.go"
	postgresContent, err := os.ReadFile(postgresPath)
	if err != nil {
		t.Fatalf("Failed to read postgres.go: %v", err)
	}

	if !strings.Contains(string(postgresContent), "cfg.PostgresURL") {
		t.Error("postgres.go should use cfg.PostgresURL for env config")
	}

	t.Log("✅ All checks passed!")
}

func TestTestingInfrastructureGeneration(t *testing.T) {
	cfg := &config.Config{
		ProjectName:   "test-with-testing",
		ModulePath:    "github.com/test/test-with-testing",
		GoVersion:     "1.21",
		Framework:     "chi",
		Databases:     []string{"postgres"},
		Logger:        "slog",
		EnableTracing: false,
		EnableMetrics: false,
		IncludeDocker: false,
		CI:            "",
		ConfigFormat:  "env",
		EnvSample:     false,
	}

	outputDir := "/tmp/test-testing-output"
	os.RemoveAll(outputDir)
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	gen := generator.New(cfg, outputDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	projectDir := outputDir + "/test-with-testing"

	// Verify go.mod includes testify and mock
	goModPath := projectDir + "/go.mod"
	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(goModContent), "github.com/stretchr/testify") {
		t.Error("go.mod should include testify dependency")
	}

	if !strings.Contains(string(goModContent), "go.uber.org/mock") {
		t.Error("go.mod should include uber-go/mock dependency")
	}

	// Verify handler tests exist
	handlerTestPath := projectDir + "/internal/handlers/handlers_test.go"
	if _, err := os.Stat(handlerTestPath); os.IsNotExist(err) {
		t.Error("Handler tests should be generated")
	} else {
		handlerTestContent, _ := os.ReadFile(handlerTestPath)
		if !strings.Contains(string(handlerTestContent), "suite.Suite") {
			t.Error("Handler tests should use testify suite")
		}
		if !strings.Contains(string(handlerTestContent), "TestHandlerSuite") {
			t.Error("Handler tests should have suite runner")
		}
	}

	// Verify mock interfaces exist
	mockInterfacesPath := projectDir + "/internal/mocks/interfaces.go"
	if _, err := os.Stat(mockInterfacesPath); os.IsNotExist(err) {
		t.Error("Mock interfaces should be generated for projects with databases")
	} else {
		mockContent, _ := os.ReadFile(mockInterfacesPath)
		if !strings.Contains(string(mockContent), "//go:generate mockgen") {
			t.Error("Mock interfaces should have go:generate directive")
		}
	}

	// Verify database test suite exists
	dbTestPath := projectDir + "/internal/database/database_test.go"
	if _, err := os.Stat(dbTestPath); os.IsNotExist(err) {
		t.Error("Database tests should be generated")
	} else {
		dbTestContent, _ := os.ReadFile(dbTestPath)
		if !strings.Contains(string(dbTestContent), "gomock.NewController") {
			t.Error("Database tests should use gomock")
		}
	}

	// Verify testing documentation exists
	testingDocsPath := projectDir + "/docs/TESTING.md"
	if _, err := os.Stat(testingDocsPath); os.IsNotExist(err) {
		t.Error("Testing documentation should be generated")
	} else {
		docsContent, _ := os.ReadFile(testingDocsPath)
		if !strings.Contains(string(docsContent), "testify") {
			t.Error("Testing docs should mention testify")
		}
		if !strings.Contains(string(docsContent), "uber-go/mock") {
			t.Error("Testing docs should mention uber-go/mock")
		}
	}

	// Verify Makefile has test targets
	makefilePath := projectDir + "/Makefile"
	makefileContent, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	requiredTargets := []string{
		"test:",
		"test-coverage:",
		"test-unit:",
		"test-integration:",
		"generate-mocks:",
		"install-tools:",
	}

	for _, target := range requiredTargets {
		if !strings.Contains(string(makefileContent), target) {
			t.Errorf("Makefile should contain target: %s", target)
		}
	}

	t.Log("✅ Testing infrastructure checks passed!")
}
