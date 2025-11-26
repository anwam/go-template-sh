# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-11-26

### Added
- Initial release of go-template-sh
- Interactive CLI tool for generating Go HTTP server templates
- Support for 5 HTTP frameworks:
  - net/http (standard library)
  - Chi
  - Gin
  - Echo
  - Fiber
- Support for 4 database systems:
  - PostgreSQL (pgx/v5)
  - MySQL
  - MongoDB
  - Redis (caching)
- Support for 3 logging libraries:
  - slog (standard library)
  - Zap (uber-go/zap)
  - Zerolog
- Complete observability features:
  - Structured logging
  - Distributed tracing (OpenTelemetry)
  - Prometheus metrics
- Twelve-factor app compliance:
  - Environment-based configuration
  - Graceful shutdown
  - Health and readiness checks
  - Stateless processes
  - Log streaming
- Production-ready features:
  - Request ID generation
  - Panic recovery middleware
  - Request timeout handling
  - Context propagation
- Docker support:
  - Multi-stage Dockerfile
  - docker-compose.yml with all services
  - Health checks for containers
- CI/CD configurations:
  - GitHub Actions workflow
  - GitLab CI pipeline
- Comprehensive documentation:
  - README.md with full feature list
  - QUICKSTART.md with step-by-step guide
  - CONTRIBUTING.md for contributors
  - PROJECT_SUMMARY.md with technical details
  - Examples directory with common configurations
- Clean project structure:
  - Organized internal packages
  - Separation of concerns
  - Idiomatic Go code
- Build tools:
  - Makefile with common commands
  - .gitignore for Go projects
  - .dockerignore for Docker builds

### Technical Details
- Built with Cobra CLI framework
- Interactive prompts using AlecAivazis/survey
- Template-based code generation
- Modular generator architecture
- Input validation and error handling
- ~3,174 lines of Go code
- 14 Go source files
- MIT License

### Supported Go Versions
- Go 1.21+
- Go 1.22
- Go 1.23

## [Unreleased]

### Planned Features
- GraphQL support with gqlgen
- gRPC service templates
- Message queue integrations (RabbitMQ, Kafka)
- Authentication middleware templates (JWT, OAuth2)
- API documentation generation (Swagger/OpenAPI)
- Database migration tools integration
- WebSocket support
- Enhanced test templates
- Non-interactive mode with config files
- Project update functionality
- Plugin system for custom templates

### Ideas
- Multiple language support (beyond English)
- Template customization options
- Interactive project modification
- Deployment configuration (Kubernetes, Terraform)
- Service mesh integration (Istio, Linkerd)
- Performance profiling setup
- Security scanning integration

---

## Version History

- **1.0.0** (2024-11-26): Initial release with core features

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.

## Support

- Issues: https://github.com/anwam/go-template-sh/issues
- Discussions: https://github.com/anwam/go-template-sh/discussions
- Documentation: See README.md and QUICKSTART.md
