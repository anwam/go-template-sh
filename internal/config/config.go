package config

import (
	"fmt"
	"regexp"
	"slices"
)

type Config struct {
	ProjectName   string
	ModulePath    string
	GoVersion     string
	Framework     string
	Databases     []string
	Logger        string
	EnableTracing bool
	EnableMetrics bool
	IncludeDocker bool
	CI            string
	ConfigFormat  string // "env", "json", "yaml", or "toml"
	EnvSample     bool   // Generate sample .env file with documentation
}

// Validate checks that the configuration is valid for project generation.
func (c *Config) Validate() error {
	if c.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}

	// Project name must be lowercase alphanumeric with hyphens/underscores
	projectNameRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9-_]*$`)
	if !projectNameRegex.MatchString(c.ProjectName) {
		return fmt.Errorf("project name must start with lowercase letter or number and contain only lowercase letters, numbers, hyphens, and underscores")
	}

	if c.ModulePath == "" {
		return fmt.Errorf("module path is required")
	}

	// Module path should look like a valid Go module path
	modulePathRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)+$`)
	if !modulePathRegex.MatchString(c.ModulePath) {
		return fmt.Errorf("module path must be a valid Go module path (e.g., github.com/user/project)")
	}

	if c.GoVersion == "" {
		return fmt.Errorf("Go version is required")
	}

	validGoVersions := []string{"1.21", "1.22", "1.23", "1.24"}
	if !slices.Contains(validGoVersions, c.GoVersion) {
		return fmt.Errorf("Go version must be one of: %v", validGoVersions)
	}

	validFrameworks := []string{"stdlib", "chi", "gin", "echo", "fiber"}
	if c.Framework != "" && !slices.Contains(validFrameworks, c.Framework) {
		return fmt.Errorf("framework must be one of: %v", validFrameworks)
	}

	validLoggers := []string{"slog", "zap", "zerolog"}
	if c.Logger != "" && !slices.Contains(validLoggers, c.Logger) {
		return fmt.Errorf("logger must be one of: %v", validLoggers)
	}

	validConfigFormats := []string{"", "env", "yaml", "json", "toml"}
	if !slices.Contains(validConfigFormats, c.ConfigFormat) {
		return fmt.Errorf("config format must be one of: env, yaml, json, toml")
	}

	return nil
}

func (c *Config) HasDatabase(db string) bool {
	return slices.Contains(c.Databases, db)
}

func (c *Config) NeedsCache() bool {
	return c.HasDatabase("redis")
}

func (c *Config) NeedsSQL() bool {
	return c.HasDatabase("postgres") || c.HasDatabase("mysql")
}

func (c *Config) NeedsNoSQL() bool {
	return c.HasDatabase("mongodb")
}
