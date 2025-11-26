# go-template-sh

A powerful CLI tool to generate production-ready Go HTTP server project templates that follow the twelve-factor app methodology and Go community best practices.

## Features

- 🚀 **Interactive Prompts**: Easy-to-use CLI with intuitive prompts
- 🎯 **Multiple Frameworks**: Support for net/http, Chi, Gin, Echo, and Fiber
- 📊 **Complete Observability**: Built-in logging, tracing (OpenTelemetry), and metrics (Prometheus)
- 🗄️ **Database Support**: PostgreSQL, MySQL, MongoDB, and Redis
- 🔍 **Structured Logging**: Choose between slog, Zap, or Zerolog
- 🐳 **Docker Ready**: Dockerfile and docker-compose.yml included
- ⚙️ **CI/CD**: GitHub Actions and GitLab CI configurations
- 📐 **Clean Architecture**: Well-organized project structure
- ✅ **12-Factor App**: Follows all twelve-factor app principles
- 🛡️ **Production Ready**: Graceful shutdown, health checks, and middleware

## Installation

```bash
go install github.com/anwam/go-template-sh@latest
```

Or build from source:

```bash
git clone https://github.com/anwam/go-template-sh.git
cd go-template-sh
go build -o go-template-sh
```

## Usage

Run the CLI tool:

```bash
go-template-sh
```

Or specify options:

```bash
go-template-sh --name my-api --output ./projects
```

### Options

- `-n, --name`: Project name (if not provided, will prompt)
- `-o, --output`: Output directory (default: current directory)
- `-h, --help`: Show help message

## Interactive Configuration

The tool will guide you through the following configuration options:

1. **Project Name**: Your project's name (lowercase, alphanumeric, hyphens, underscores)
2. **Go Module Path**: Full module path (e.g., github.com/username/project)
3. **Go Version**: Choose from 1.23, 1.22, or 1.21
4. **HTTP Framework**: 
   - net/http (standard library)
   - Chi
   - Gin
   - Echo
   - Fiber
5. **Databases**: Select one or more (or none)
   - PostgreSQL
   - MySQL
   - MongoDB
   - Redis (cache)
6. **Logger**: Choose your logging library
   - slog (standard library)
   - Zap
   - Zerolog
7. **Distributed Tracing**: Enable OpenTelemetry tracing
8. **Metrics**: Enable Prometheus metrics
9. **Docker**: Generate Dockerfile and docker-compose.yml
10. **CI/CD**: Choose between GitHub Actions, GitLab CI, or none

## Generated Project Structure

```
your-project/
├── cmd/
│   └── your-project/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration (12-factor: III)
│   ├── server/
│   │   └── server.go            # HTTP server setup
│   ├── handlers/
│   │   └── handlers.go          # HTTP handlers
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware
│   ├── observability/
│   │   ├── observability.go     # Observability setup
│   │   └── logger.go            # Logger initialization
│   ├── database/                # (if databases selected)
│   │   ├── postgres.go
│   │   ├── mysql.go
│   │   └── mongodb.go
│   └── cache/                   # (if Redis selected)
│       └── redis.go
├── pkg/                         # Public packages (if needed)
├── .github/                     # (if GitHub Actions selected)
│   └── workflows/
│       └── ci.yml
├── .gitlab-ci.yml               # (if GitLab CI selected)
├── Dockerfile                   # (if Docker selected)
├── docker-compose.yml           # (if Docker selected)
├── .dockerignore
├── .gitignore
├── .env.example
├── Makefile
├── go.mod
└── README.md
```

## Twelve-Factor App Principles

The generated projects strictly follow the [twelve-factor app](https://12factor.net/) methodology:

1. **Codebase**: One codebase tracked in Git
2. **Dependencies**: Explicitly declared via go.mod
3. **Config**: Configuration via environment variables
4. **Backing Services**: Databases as attached resources
5. **Build, Release, Run**: Strict separation of stages
6. **Processes**: Execute as stateless processes
7. **Port Binding**: Self-contained HTTP server
8. **Concurrency**: Scale out via process model
9. **Disposability**: Graceful shutdown implemented
10. **Dev/Prod Parity**: Keep development and production similar
11. **Logs**: Treat logs as event streams (stdout/stderr)
12. **Admin Processes**: Run admin tasks as one-off processes

## Project Features

### Observability

Every generated project includes:

- **Structured Logging**: JSON logs with configurable levels
- **Distributed Tracing**: OpenTelemetry integration (optional)
- **Metrics**: Prometheus metrics endpoint (optional)
- **Health Checks**: `/health` and `/ready` endpoints

### Middleware

Built-in middleware for:

- Request ID generation
- Request/response logging
- Panic recovery
- Distributed tracing propagation
- Timeout handling

### Database Support

Clean database connection patterns with:

- Connection pooling
- Health checks
- Context-aware operations
- Graceful shutdown

### Docker Support

Production-ready Docker setup:

- Multi-stage builds
- Alpine-based images
- Health checks
- docker-compose with all services

### CI/CD

Ready-to-use CI/CD pipelines:

- Automated testing
- Code linting
- Coverage reports
- Build verification

## Example

```bash
$ go-template-sh

🚀 Welcome to go-template-sh - Go HTTP Server Template Generator

? Project name: my-awesome-api
? Go module path: github.com/myuser/my-awesome-api
? Go version: 1.23
? HTTP framework: Chi
? Database(s): PostgreSQL, Redis (cache)
? Logger: slog (standard library)
? Enable distributed tracing (OpenTelemetry)? Yes
? Enable Prometheus metrics? Yes
? Generate Dockerfile and docker-compose.yml? Yes
? CI/CD configuration: GitHub Actions
? Ready to generate your project. Continue? Yes

✅ Project generated successfully!
📁 Location: ./my-awesome-api

Next steps:
  cd my-awesome-api
  go mod download
  make run
```

## Generated Project Usage

After generating a project:

```bash
cd my-awesome-api

# Copy environment variables
cp .env.example .env

# Download dependencies
go mod download

# Start Docker services (if using Docker)
make docker-up

# Run the application
make run

# Run tests
make test

# Build for production
make build
```

## Requirements

- Go 1.21 or higher

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details

## Author

Created with ❤️ for the Go community

## Roadmap

- [ ] Add GraphQL support
- [ ] Add gRPC support
- [ ] Add message queue integrations (RabbitMQ, Kafka)
- [ ] Add authentication middleware templates
- [ ] Add API documentation generation (Swagger/OpenAPI)
- [ ] Add migration tool integration
- [ ] Add more test templates
- [ ] Interactive mode for adding features to existing projects

## Credits

Built using:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Survey](https://github.com/AlecAivazis/survey) - Interactive prompts
