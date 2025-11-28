package cmd

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/anwam/go-template-sh/internal/config"
	"github.com/anwam/go-template-sh/internal/generator"
	"github.com/anwam/go-template-sh/internal/prompt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-template-sh",
	Short: "Generate Go HTTP server project templates",
	Long: `A CLI tool to scaffold production-ready Go HTTP server projects.
Follows twelve-factor app methodology and Go community best practices.
Includes complete observability (logging, tracing, metrics).

Examples:
  # Interactive mode (default)
  go-template-sh

  # Non-interactive mode with all options
  go-template-sh --name my-api --module github.com/user/my-api --framework chi

  # Dry-run to see what would be generated
  go-template-sh --name my-api --dry-run

  # Show version
  go-template-sh version`,
	RunE: runGenerate,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-template-sh %s\n", Version)
		fmt.Printf("  Git commit: %s\n", GitCommit)
		fmt.Printf("  Built:      %s\n", BuildDate)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add version subcommand
	rootCmd.AddCommand(versionCmd)

	// Output settings
	rootCmd.Flags().StringP("output", "o", ".", "Output directory for the generated project")

	// Project identification
	rootCmd.Flags().StringP("name", "n", "", "Project name")
	rootCmd.Flags().StringP("module", "m", "", "Go module path (e.g., github.com/user/project)")

	// Project configuration
	rootCmd.Flags().String("go-version", "1.23", "Go version (1.21, 1.22, 1.23, 1.24)")
	rootCmd.Flags().StringP("framework", "f", "stdlib", "HTTP framework (stdlib, chi, gin, echo, fiber)")
	rootCmd.Flags().StringSlice("database", nil, "Database(s) to include (postgres, mysql, mongodb, redis)")
	rootCmd.Flags().StringP("logger", "l", "slog", "Logger (slog, zap, zerolog)")
	rootCmd.Flags().String("config-format", "env", "Config format (env, yaml, json, toml)")
	rootCmd.Flags().String("ci", "", "CI/CD configuration (github, gitlab, or empty for none)")

	// Feature flags
	rootCmd.Flags().Bool("tracing", true, "Enable OpenTelemetry tracing")
	rootCmd.Flags().Bool("metrics", true, "Enable Prometheus metrics")
	rootCmd.Flags().Bool("docker", true, "Generate Dockerfile and docker-compose.yml")
	rootCmd.Flags().Bool("env-sample", true, "Generate documented .env.example file")

	// Mode flags
	rootCmd.Flags().Bool("dry-run", false, "Show what would be generated without writing files")
	rootCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt (non-interactive)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Check if we're in non-interactive mode
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	outputDir, _ := cmd.Flags().GetString("output")

	// Try to build config from flags first
	cfg, isNonInteractive, err := buildConfigFromFlags(cmd)
	if err != nil {
		return err
	}

	if isNonInteractive {
		// Non-interactive mode - use flags only
		fmt.Println("🚀 go-template-sh - Go HTTP Server Template Generator")
		fmt.Println()
	} else {
		// Interactive mode
		fmt.Println("🚀 Welcome to go-template-sh - Go HTTP Server Template Generator")
		fmt.Println()

		projectName, _ := cmd.Flags().GetString("name")
		cfg, err = prompt.CollectConfiguration(projectName)
		if err != nil {
			return fmt.Errorf("failed to collect configuration: %w", err)
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Show configuration summary
	printConfigSummary(cfg, outputDir, dryRun)

	if dryRun {
		fmt.Println()
		fmt.Println("🔍 Dry-run mode - no files will be written")
		printDryRunSummary(cfg)
		return nil
	}

	// Confirm generation (unless --yes flag is set or in dry-run)
	if !skipConfirm && !isNonInteractive {
		confirmed := false
		confirmPrompt := &survey.Confirm{
			Message: "Ready to generate your project. Continue?",
			Default: true,
		}
		if err := survey.AskOne(confirmPrompt, &confirmed); err != nil {
			return err
		}

		if !confirmed {
			fmt.Println("❌ Project generation cancelled")
			return nil
		}
	}

	gen := generator.New(cfg, outputDir)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Project generated successfully!")
	fmt.Printf("📁 Location: %s/%s\n", outputDir, cfg.ProjectName)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", cfg.ProjectName)
	fmt.Println("  go mod download")
	fmt.Println("  make run")
	fmt.Println()

	return nil
}

// buildConfigFromFlags attempts to build a config from CLI flags.
// Returns (config, isNonInteractive, error)
func buildConfigFromFlags(cmd *cobra.Command) (*config.Config, bool, error) {
	name, _ := cmd.Flags().GetString("name")
	module, _ := cmd.Flags().GetString("module")

	// If name is provided, we're in non-interactive mode
	if name == "" {
		return nil, false, nil
	}

	// Build config from flags
	cfg := &config.Config{
		ProjectName: name,
	}

	// Module path - default to github.com/user/name if not provided
	if module != "" {
		cfg.ModulePath = module
	} else {
		cfg.ModulePath = fmt.Sprintf("github.com/user/%s", name)
	}

	goVersion, _ := cmd.Flags().GetString("go-version")
	cfg.GoVersion = goVersion

	framework, _ := cmd.Flags().GetString("framework")
	cfg.Framework = framework

	databases, _ := cmd.Flags().GetStringSlice("database")
	cfg.Databases = databases

	logger, _ := cmd.Flags().GetString("logger")
	cfg.Logger = logger

	configFormat, _ := cmd.Flags().GetString("config-format")
	cfg.ConfigFormat = configFormat

	ci, _ := cmd.Flags().GetString("ci")
	cfg.CI = ci

	tracing, _ := cmd.Flags().GetBool("tracing")
	cfg.EnableTracing = tracing

	metrics, _ := cmd.Flags().GetBool("metrics")
	cfg.EnableMetrics = metrics

	docker, _ := cmd.Flags().GetBool("docker")
	cfg.IncludeDocker = docker

	envSample, _ := cmd.Flags().GetBool("env-sample")
	cfg.EnvSample = envSample

	return cfg, true, nil
}

func printConfigSummary(cfg *config.Config, outputDir string, dryRun bool) {
	fmt.Println("📋 Configuration Summary:")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("  Project:     %s\n", cfg.ProjectName)
	fmt.Printf("  Module:      %s\n", cfg.ModulePath)
	fmt.Printf("  Go Version:  %s\n", cfg.GoVersion)
	fmt.Printf("  Framework:   %s\n", cfg.Framework)
	fmt.Printf("  Logger:      %s\n", cfg.Logger)

	if len(cfg.Databases) > 0 {
		fmt.Printf("  Databases:   %s\n", strings.Join(cfg.Databases, ", "))
	} else {
		fmt.Printf("  Databases:   none\n")
	}

	fmt.Printf("  Tracing:     %v\n", cfg.EnableTracing)
	fmt.Printf("  Metrics:     %v\n", cfg.EnableMetrics)
	fmt.Printf("  Docker:      %v\n", cfg.IncludeDocker)

	if cfg.CI != "" {
		fmt.Printf("  CI/CD:       %s\n", cfg.CI)
	} else {
		fmt.Printf("  CI/CD:       none\n")
	}

	fmt.Printf("  Config:      %s\n", cfg.ConfigFormat)
	fmt.Printf("  Output:      %s/%s\n", outputDir, cfg.ProjectName)
	fmt.Println(strings.Repeat("-", 40))
}

func printDryRunSummary(cfg *config.Config) {
	fmt.Println()
	fmt.Println("📁 Files that would be generated:")
	fmt.Println()

	// Core files
	files := []string{
		"go.mod",
		fmt.Sprintf("cmd/%s/main.go", cfg.ProjectName),
		"internal/config/config.go",
		"internal/server/server.go",
		"internal/handlers/handlers.go",
		"internal/middleware/middleware.go",
		"internal/observability/observability.go",
		"internal/observability/logger.go",
		"Makefile",
		"README.md",
		".gitignore",
		".env.example",
	}

	// Database files
	if cfg.HasDatabase("postgres") {
		files = append(files, "internal/database/postgres.go")
	}
	if cfg.HasDatabase("mysql") {
		files = append(files, "internal/database/mysql.go")
	}
	if cfg.HasDatabase("mongodb") {
		files = append(files, "internal/database/mongodb.go")
	}
	if cfg.HasDatabase("redis") {
		files = append(files, "internal/cache/redis.go")
	}

	// Docker files
	if cfg.IncludeDocker {
		files = append(files, "Dockerfile", "docker-compose.yml", ".dockerignore")
	}

	// CI files
	if cfg.CI == "github" {
		files = append(files, ".github/workflows/ci.yml")
	} else if cfg.CI == "gitlab" {
		files = append(files, ".gitlab-ci.yml")
	}

	// Config files
	if cfg.ConfigFormat == "yaml" {
		files = append(files, "config.yaml.example")
	} else if cfg.ConfigFormat == "json" {
		files = append(files, "config.json.example")
	} else if cfg.ConfigFormat == "toml" {
		files = append(files, "config.toml.example")
	}

	for _, f := range files {
		fmt.Printf("  📄 %s/%s\n", cfg.ProjectName, f)
	}

	// Directories
	fmt.Println()
	fmt.Println("📂 Directories that would be created:")
	dirs := []string{
		fmt.Sprintf("cmd/%s", cfg.ProjectName),
		"internal/config",
		"internal/server",
		"internal/handlers",
		"internal/middleware",
		"internal/observability",
		"internal/mocks",
		"pkg",
		"docs",
	}

	if cfg.NeedsSQL() || cfg.NeedsNoSQL() {
		dirs = append(dirs, "internal/database")
	}
	if cfg.NeedsCache() {
		dirs = append(dirs, "internal/cache")
	}
	if cfg.CI == "github" {
		dirs = append(dirs, ".github/workflows")
	}

	for _, d := range dirs {
		fmt.Printf("  📁 %s/%s/\n", cfg.ProjectName, d)
	}
}
