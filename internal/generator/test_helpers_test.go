package generator

import (
"github.com/anwam/go-template-sh/internal/config"
"github.com/anwam/go-template-sh/internal/fsys"
)

// createTestConfig returns a minimal valid config for testing.
func createTestConfig() *config.Config {
	return &config.Config{
		ProjectName:  "test-project",
		ModulePath:   "github.com/test/test-project",
		GoVersion:    "1.23",
		Framework:    "stdlib",
		Logger:       "slog",
		ConfigFormat: "env",
	}
}

// createMemoryFS returns a new in-memory file system for testing.
func createMemoryFS() *fsys.MemoryFileSystem {
	return fsys.NewMemory()
}

// createTestGenerator returns a Generator with an in-memory file system.
func createTestGenerator(cfg *config.Config) (*Generator, *fsys.MemoryFileSystem) {
	mfs := createMemoryFS()
	gen := New(cfg, "/output", WithFileSystem(mfs))
	return gen, mfs
}
