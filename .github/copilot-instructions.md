# Copilot Instructions for go-template-sh

This repository contains a CLI tool written in Go that generates production-ready Go HTTP server project templates.

## 🏗 Architecture & Structure

- **Entry Point**: `main.go` initializes the application and calls `cmd.Execute()`.
- **CLI Framework**: Uses `spf13/cobra` for command handling (`cmd/root.go`).
- **User Interaction**: Uses `AlecAivazis/survey` for interactive prompts (`internal/prompt`).
- **Core Logic**: The `internal/generator` package orchestrates the entire generation process.
  - `generator.go`: Main entry point for the generation logic (`Generate()` method).
  - `templates.go`: Contains logic for `go.mod` and dependency management.
  - `files.go`: Helper methods for file system operations (e.g., `writeFile`).
  - Other `*.go` files in `generator/` (e.g., `server.go`, `handlers.go`) contain specific template logic.
- **Configuration**: `internal/config/Config` struct holds all user selections (framework, database, etc.) and is passed around to control generation.

## 🧩 Code Patterns & Conventions

### Template Generation
- **Mechanism**: Templates are **hardcoded Go strings** within the `internal/generator` package, NOT external text files.
- **Variable Substitution**: Use `fmt.Sprintf` for injecting dynamic values (like module paths) into templates.
- **Example**:
  ```go
  content := fmt.Sprintf(`package main
  import "%s/internal/config"`, g.config.ModulePath)
  ```

### Adding New Features
To add a new option (e.g., a new database or framework):
1.  **Config**: Add a field to `Config` struct in `internal/config/config.go`.
2.  **Prompt**: Update `CollectConfiguration` in `internal/prompt/prompt.go` to ask the user.
3.  **Generator**:
    - Update `buildDependencies` in `internal/generator/templates.go` to add required Go modules.
    - Create or update the relevant `generate*` method in `internal/generator/`.
    - Call the new method in `Generate()` within `internal/generator/generator.go`.

### Dependency Management
- The tool **manually constructs** the `go.mod` file for the *generated* project.
- Dependencies for generated projects are defined in `buildDependencies()` in `internal/generator/templates.go`.
- **Do not** rely on `go mod tidy` running automatically during generation; the tool writes the `require` block explicitly.

## 🛠 Development Workflow

- **Build**: `go build -o go-template-sh`
- **Run**: `go run main.go` (or `./go-template-sh` after build)
- **Testing**:
  - Currently, there are **no automated tests**.
  - **Manual Verification**: Run the tool, generate a project, and verify the generated code compiles and runs.
  - *Future Goal*: Add unit tests for generator functions using in-memory file systems (e.g., `afero`).

## ⚠️ Critical Considerations

- **12-Factor App**: The *generated* code must strictly follow 12-factor principles (config via env, stateless, etc.).
- **Project Layout**: Generated projects follow the standard Go project layout (`cmd/`, `internal/`, `pkg/`).
- **Error Handling**: The generator itself should fail fast. If a file cannot be written, return the error immediately.
