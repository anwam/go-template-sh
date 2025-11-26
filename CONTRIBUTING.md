# Contributing to go-template-sh

Thank you for considering contributing to go-template-sh! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-template-sh.git
   cd go-template-sh
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the tool:
   ```bash
   make build
   ```

## Development Workflow

1. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Test your changes:
   ```bash
   make test
   ./go-template-sh --name test-project --output /tmp
   ```

4. Commit your changes with a descriptive message:
   ```bash
   git commit -m "Add feature: description of your feature"
   ```

5. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

6. Create a Pull Request

## Code Structure

```
go-template-sh/
├── cmd/                    # CLI commands
│   └── root.go            # Root command and CLI setup
├── internal/
│   ├── config/            # Configuration structure
│   ├── prompt/            # Interactive prompts
│   └── generator/         # Template generation logic
│       ├── generator.go   # Main generator
│       ├── templates.go   # Go module and config templates
│       ├── server.go      # Server templates
│       ├── handlers.go    # Handler templates
│       ├── middleware.go  # Middleware templates
│       ├── observability.go # Observability templates
│       ├── database.go    # Database templates
│       ├── files.go       # Makefile, README, .env templates
│       ├── docker.go      # Docker templates
│       └── ci.go          # CI/CD templates
└── main.go                # Entry point

```

## Adding New Features

### Adding a New HTTP Framework

1. Update `internal/prompt/prompt.go`:
   - Add framework to options list
   - Add parsing in `parseFramework()`

2. Update `internal/generator/server.go`:
   - Add new `generate[Framework]Server()` function
   - Add case in `getServerContent()`

3. Update `internal/generator/handlers.go`:
   - Add framework-specific handlers
   - Add case in `getFrameworkSpecificHandlers()`

4. Update `internal/generator/middleware.go`:
   - Add framework-specific middleware
   - Add case in `getFrameworkMiddleware()`

5. Update dependencies in `internal/generator/templates.go`:
   - Add framework dependency to `buildDependencies()`

### Adding a New Database

1. Update `internal/prompt/prompt.go`:
   - Add database to options
   - Add parsing in `parseDatabases()`

2. Update `internal/config/config.go`:
   - Add helper method if needed

3. Update `internal/generator/database.go`:
   - Add `generate[Database]DB()` function

4. Update `internal/generator/templates.go`:
   - Add config fields in `getDatabaseConfigFields()`
   - Add load statements in `getConfigLoadStatements()`
   - Add dependency in `buildDependencies()`

5. Update `internal/generator/docker.go`:
   - Add service to `generateDockerCompose()`

### Adding a New Logger

1. Update `internal/prompt/prompt.go`:
   - Add logger to options
   - Add parsing in `parseLogger()`

2. Update `internal/generator/observability.go`:
   - Add logger implementation in `getLoggerFileContent()`
   - Add logger initialization in `getLoggerInitCode()`

3. Update `internal/generator/middleware.go`:
   - Add logger-specific middleware implementations

4. Update `internal/generator/templates.go`:
   - Add dependency in `buildDependencies()`

## Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Run `golangci-lint run` before submitting
- Add comments for exported functions
- Keep functions focused and small

## Testing

Before submitting a PR:

1. Build the tool: `make build`
2. Test generation with various configurations
3. Verify generated projects build correctly
4. Check that all twelve-factor principles are maintained

### Manual Testing Checklist

Generate projects with:
- [ ] Each framework (stdlib, Chi, Gin, Echo, Fiber)
- [ ] Each logger (slog, Zap, Zerolog)
- [ ] Each database (PostgreSQL, MySQL, MongoDB, Redis)
- [ ] Tracing enabled/disabled
- [ ] Metrics enabled/disabled
- [ ] Docker enabled/disabled
- [ ] Each CI option (GitHub, GitLab, None)

Verify generated projects:
- [ ] `go mod download` succeeds
- [ ] `go build` succeeds
- [ ] All endpoints respond correctly
- [ ] Docker compose starts successfully (if enabled)
- [ ] CI config is valid (if enabled)

## Pull Request Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Include screenshots for UI changes (if applicable)
- Ensure all tests pass
- Update documentation if needed
- Keep PRs focused on a single feature or fix

## Reporting Issues

When reporting issues, please include:

- go-template-sh version
- Go version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Any error messages or logs

## Feature Requests

We welcome feature requests! Please:

- Check if the feature has already been requested
- Provide a clear use case
- Explain how it aligns with twelve-factor app principles
- Consider contributing the implementation

## Questions?

Feel free to open an issue for questions or join discussions.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
