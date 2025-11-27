package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/anwam/go-template-sh/internal/generator/templates"
)

// TemplateData holds all data passed to templates.
type TemplateData struct {
	// Project configuration
	ProjectName  string
	ModulePath   string
	GoVersion    string
	Framework    string
	Logger       string
	ConfigFormat string

	// Feature flags
	HasPostgres   bool
	HasMySQL      bool
	HasMongoDB    bool
	HasRedis      bool
	EnableTracing bool
	EnableMetrics bool
	IncludeDocker bool
	NeedsSQL      bool
	NeedsNoSQL    bool
	NeedsCache    bool

	// Computed values
	ConfigFieldRef func(string) string
	LoggerType     string
	LoggerImport   string
}

// ServerTemplateData holds data for server templates.
type ServerTemplateData struct {
	ModulePath    string
	PortRef       string
	EnvRef        string
	EnableTracing bool
	EnableMetrics bool
}

// DockerTemplateData holds data for Docker templates.
type DockerTemplateData struct {
	ProjectName string
	GoVersion   string
}

// MakefileTemplateData holds data for Makefile templates.
type MakefileTemplateData struct {
	ProjectName   string
	GoVersion     string
	IncludeDocker bool
}

// NewTemplateData creates TemplateData from a config.
func (g *Generator) NewTemplateData() *TemplateData {
	return &TemplateData{
		ProjectName:    g.config.ProjectName,
		ModulePath:     g.config.ModulePath,
		GoVersion:      g.config.GoVersion,
		Framework:      g.config.Framework,
		Logger:         g.config.Logger,
		ConfigFormat:   g.config.ConfigFormat,
		HasPostgres:    g.config.HasDatabase("postgres"),
		HasMySQL:       g.config.HasDatabase("mysql"),
		HasMongoDB:     g.config.HasDatabase("mongodb"),
		HasRedis:       g.config.HasDatabase("redis"),
		EnableTracing:  g.config.EnableTracing,
		EnableMetrics:  g.config.EnableMetrics,
		IncludeDocker:  g.config.IncludeDocker,
		NeedsSQL:       g.config.NeedsSQL(),
		NeedsNoSQL:     g.config.NeedsNoSQL(),
		NeedsCache:     g.config.NeedsCache(),
		ConfigFieldRef: g.getConfigFieldReference,
		LoggerType:     g.getLoggerType(),
		LoggerImport:   g.getLoggerImport(),
	}
}

func (g *Generator) getLoggerType() string {
	switch g.config.Logger {
	case "slog":
		return "*slog.Logger"
	case "zap":
		return "*zap.Logger"
	case "zerolog":
		return "*zerolog.Logger"
	default:
		return "*slog.Logger"
	}
}

func (g *Generator) getLoggerImport() string {
	switch g.config.Logger {
	case "slog":
		return `"log/slog"`
	case "zap":
		return `"go.uber.org/zap"`
	case "zerolog":
		return `"github.com/rs/zerolog"`
	default:
		return `"log/slog"`
	}
}

// executeTemplate parses and executes a template with the given data.
func executeTemplate(name, tmplText string, data any) (string, error) {
	funcMap := template.FuncMap{
		"join":     strings.Join,
		"contains": strings.Contains,
		"lower":    strings.ToLower,
		"upper":    strings.ToUpper,
		"title":    strings.Title,
		"indent": func(spaces int, s string) string {
			pad := strings.Repeat(" ", spaces)
			lines := strings.Split(s, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = pad + line
				}
			}
			return strings.Join(lines, "\n")
		},
		"trimSuffix": strings.TrimSuffix,
		"trimPrefix": strings.TrimPrefix,
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(tmplText)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// loadEmbeddedTemplate loads a template from the embedded filesystem.
func loadEmbeddedTemplate(name string) (string, error) {
	data, err := templates.FS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("failed to load embedded template %s: %w", name, err)
	}
	return string(data), nil
}

// executeEmbeddedTemplate loads and executes an embedded template.
func executeEmbeddedTemplate(name string, data any) (string, error) {
	tmplText, err := loadEmbeddedTemplate(name)
	if err != nil {
		return "", err
	}
	return executeTemplate(name, tmplText, data)
}

// writeEmbeddedTemplate loads an embedded template, executes it, and writes to a file.
func (g *Generator) writeEmbeddedTemplate(relativePath, templateName string, data any) error {
	content, err := executeEmbeddedTemplate(templateName, data)
	if err != nil {
		return err
	}
	return g.writeFile(relativePath, content)
}

// writeTemplate parses, executes and writes a template to a file.
func (g *Generator) writeTemplate(relativePath, name, tmplText string, data any) error {
	content, err := executeTemplate(name, tmplText, data)
	if err != nil {
		return err
	}
	return g.writeFile(relativePath, content)
}
