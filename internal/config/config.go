package config

import "slices"

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
