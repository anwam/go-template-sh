package prompt

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/anwam/go-template-sh/internal/config"
)

func CollectConfiguration(projectName string) (*config.Config, error) {
	cfg := &config.Config{}

	if projectName == "" {
		namePrompt := &survey.Input{
			Message: "Project name:",
			Help:    "Name of your project (e.g., my-api-service)",
		}
		if err := survey.AskOne(namePrompt, &projectName, survey.WithValidator(validateProjectName)); err != nil {
			return nil, err
		}
	}
	cfg.ProjectName = projectName

	modulePrompt := &survey.Input{
		Message: "Go module path:",
		Default: fmt.Sprintf("github.com/yourusername/%s", projectName),
		Help:    "Go module path (e.g., github.com/username/project)",
	}
	if err := survey.AskOne(modulePrompt, &cfg.ModulePath); err != nil {
		return nil, err
	}

	goVersionPrompt := &survey.Select{
		Message: "Go version:",
		Options: []string{"1.23", "1.22", "1.21"},
		Default: "1.23",
	}
	if err := survey.AskOne(goVersionPrompt, &cfg.GoVersion); err != nil {
		return nil, err
	}

	frameworkPrompt := &survey.Select{
		Message: "HTTP framework:",
		Options: []string{
			"net/http (standard library)",
			"Chi",
			"Gin",
			"Echo",
			"Fiber",
		},
		Help: "Choose your preferred HTTP framework",
	}
	var frameworkChoice string
	if err := survey.AskOne(frameworkPrompt, &frameworkChoice); err != nil {
		return nil, err
	}
	cfg.Framework = parseFramework(frameworkChoice)

	dbPrompt := &survey.MultiSelect{
		Message: "Database(s):",
		Options: []string{
			"PostgreSQL",
			"MySQL",
			"MongoDB",
			"Redis (cache)",
			"None",
		},
		Help: "Select one or more databases (use space to select)",
	}
	var dbChoices []string
	if err := survey.AskOne(dbPrompt, &dbChoices); err != nil {
		return nil, err
	}
	cfg.Databases = parseDatabases(dbChoices)

	loggerPrompt := &survey.Select{
		Message: "Logger:",
		Options: []string{
			"slog (standard library)",
			"Zap",
			"Zerolog",
		},
		Default: "slog (standard library)",
	}
	var loggerChoice string
	if err := survey.AskOne(loggerPrompt, &loggerChoice); err != nil {
		return nil, err
	}
	cfg.Logger = parseLogger(loggerChoice)

	tracingPrompt := &survey.Confirm{
		Message: "Enable distributed tracing (OpenTelemetry)?",
		Default: true,
		Help:    "Add OpenTelemetry for distributed tracing",
	}
	if err := survey.AskOne(tracingPrompt, &cfg.EnableTracing); err != nil {
		return nil, err
	}

	metricsPrompt := &survey.Confirm{
		Message: "Enable Prometheus metrics?",
		Default: true,
		Help:    "Add Prometheus metrics endpoint",
	}
	if err := survey.AskOne(metricsPrompt, &cfg.EnableMetrics); err != nil {
		return nil, err
	}

	dockerPrompt := &survey.Confirm{
		Message: "Generate Dockerfile and docker-compose.yml?",
		Default: true,
	}
	if err := survey.AskOne(dockerPrompt, &cfg.IncludeDocker); err != nil {
		return nil, err
	}

	ciPrompt := &survey.Select{
		Message: "CI/CD configuration:",
		Options: []string{
			"GitHub Actions",
			"GitLab CI",
			"None",
		},
		Default: "GitHub Actions",
	}
	var ciChoice string
	if err := survey.AskOne(ciPrompt, &ciChoice); err != nil {
		return nil, err
	}
	cfg.CI = parseCI(ciChoice)

	configFormatPrompt := &survey.Select{
		Message: "Configuration file format:",
		Options: []string{
			"Environment variables (.env)",
			"YAML (config.yaml)",
			"JSON (config.json)",
			"TOML (config.toml)",
		},
		Default: "Environment variables (.env)",
		Help:    "Choose how your application will load configuration",
	}
	var configFormatChoice string
	if err := survey.AskOne(configFormatPrompt, &configFormatChoice); err != nil {
		return nil, err
	}
	cfg.ConfigFormat = parseConfigFormat(configFormatChoice)

	envSamplePrompt := &survey.Confirm{
		Message: "Generate documented .env.example file?",
		Default: true,
		Help:    "Generate a sample .env file with comments explaining each variable",
	}
	if err := survey.AskOne(envSamplePrompt, &cfg.EnvSample); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateProjectName(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid type")
	}
	if len(str) == 0 {
		return fmt.Errorf("project name cannot be empty")
	}
	matched, _ := regexp.MatchString("^[a-z0-9-_]+$", str)
	if !matched {
		return fmt.Errorf("project name must contain only lowercase letters, numbers, hyphens, and underscores")
	}
	return nil
}

func parseFramework(choice string) string {
	switch {
	case strings.Contains(choice, "net/http"):
		return "stdlib"
	case strings.Contains(choice, "Chi"):
		return "chi"
	case strings.Contains(choice, "Gin"):
		return "gin"
	case strings.Contains(choice, "Echo"):
		return "echo"
	case strings.Contains(choice, "Fiber"):
		return "fiber"
	default:
		return "stdlib"
	}
}

func parseDatabases(choices []string) []string {
	var result []string
	for _, choice := range choices {
		switch {
		case strings.Contains(choice, "PostgreSQL"):
			result = append(result, "postgres")
		case strings.Contains(choice, "MySQL"):
			result = append(result, "mysql")
		case strings.Contains(choice, "MongoDB"):
			result = append(result, "mongodb")
		case strings.Contains(choice, "Redis"):
			result = append(result, "redis")
		}
	}
	return result
}

func parseLogger(choice string) string {
	switch {
	case strings.Contains(choice, "slog"):
		return "slog"
	case strings.Contains(choice, "Zap"):
		return "zap"
	case strings.Contains(choice, "Zerolog"):
		return "zerolog"
	default:
		return "slog"
	}
}

func parseCI(choice string) string {
	switch {
	case strings.Contains(choice, "GitHub"):
		return "github"
	case strings.Contains(choice, "GitLab"):
		return "gitlab"
	default:
		return ""
	}
}

func parseConfigFormat(choice string) string {
	switch {
	case strings.Contains(choice, "YAML"):
		return "yaml"
	case strings.Contains(choice, "JSON"):
		return "json"
	case strings.Contains(choice, "TOML"):
		return "toml"
	default:
		return "env"
	}
}
