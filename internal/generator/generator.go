package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anwam/go-template-sh/internal/config"
)

type Generator struct {
	config     *config.Config
	outputDir  string
	projectDir string
}

func New(cfg *config.Config, outputDir string) *Generator {
	return &Generator{
		config:     cfg,
		outputDir:  outputDir,
		projectDir: filepath.Join(outputDir, cfg.ProjectName),
	}
}

func (g *Generator) Generate() error {
	if err := g.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	if err := g.generateGoMod(); err != nil {
		return err
	}

	if err := g.generateMainFile(); err != nil {
		return err
	}

	// Only generate the env-based config package if using env format
	// Other formats (yaml, json, toml) generate their own config.go
	if g.config.ConfigFormat == "" || g.config.ConfigFormat == "env" {
		if err := g.generateConfigPackage(); err != nil {
			return err
		}
	}

	if err := g.generateServerPackage(); err != nil {
		return err
	}

	if err := g.generateHandlers(); err != nil {
		return err
	}

	if err := g.generateMiddleware(); err != nil {
		return err
	}

	if err := g.generateObservability(); err != nil {
		return err
	}

	if err := g.generateLoggerFile(); err != nil {
		return err
	}

	if g.config.NeedsSQL() || g.config.NeedsNoSQL() {
		if err := g.generateDatabasePackages(); err != nil {
			return err
		}
	}

	if g.config.NeedsCache() {
		if err := g.generateCachePackage(); err != nil {
			return err
		}
	}

	if err := g.generateMakefile(); err != nil {
		return err
	}

	if err := g.generateEnvFile(); err != nil {
		return err
	}

	if g.config.ConfigFormat != "env" {
		if err := g.generateConfigFiles(); err != nil {
			return err
		}
	}

	if err := g.generateReadme(); err != nil {
		return err
	}

	if g.config.IncludeDocker {
		if err := g.generateDockerFiles(); err != nil {
			return err
		}
	}

	if g.config.CI != "" {
		if err := g.generateCIFiles(); err != nil {
			return err
		}
	}

	if err := g.generateGitignore(); err != nil {
		return err
	}

	if err := g.generateTestFiles(); err != nil {
		return err
	}

	return nil
}

func (g *Generator) createDirectoryStructure() error {
	dirs := []string{
		g.projectDir,
		filepath.Join(g.projectDir, "cmd", g.config.ProjectName),
		filepath.Join(g.projectDir, "internal", "config"),
		filepath.Join(g.projectDir, "internal", "server"),
		filepath.Join(g.projectDir, "internal", "handlers"),
		filepath.Join(g.projectDir, "internal", "middleware"),
		filepath.Join(g.projectDir, "internal", "observability"),
		filepath.Join(g.projectDir, "pkg"),
	}

	if g.config.NeedsSQL() || g.config.NeedsNoSQL() {
		dirs = append(dirs, filepath.Join(g.projectDir, "internal", "database"))
	}

	if g.config.NeedsCache() {
		dirs = append(dirs, filepath.Join(g.projectDir, "internal", "cache"))
	}

	// Add directories for testing
	dirs = append(dirs,
		filepath.Join(g.projectDir, "internal", "mocks"),
		filepath.Join(g.projectDir, "docs"),
	)

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (g *Generator) writeFile(relativePath, content string) error {
	fullPath := filepath.Join(g.projectDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}
