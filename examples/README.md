# Example Configurations

This directory contains example configurations for common use cases.

## Basic REST API

**Use Case**: Simple REST API with PostgreSQL

Configuration:
- Framework: Chi
- Database: PostgreSQL
- Logger: slog
- Tracing: Yes
- Metrics: Yes
- Docker: Yes

**Best for**: Standard CRUD APIs, microservices

## High-Performance API

**Use Case**: High-throughput API with Redis caching

Configuration:
- Framework: Fiber
- Databases: PostgreSQL, Redis
- Logger: Zerolog
- Tracing: Yes
- Metrics: Yes
- Docker: Yes

**Best for**: High-traffic applications, real-time systems

## Minimal Service

**Use Case**: Lightweight service with no external dependencies

Configuration:
- Framework: net/http (stdlib)
- Database: None
- Logger: slog
- Tracing: No
- Metrics: No
- Docker: No

**Best for**: Proxy services, simple webhooks, utility services
