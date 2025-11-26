# Quick Start Guide

This guide will walk you through creating your first Go HTTP server using go-template-sh.

## Prerequisites

- Go 1.21 or higher installed
- (Optional) Docker and docker-compose for running databases locally

## Step 1: Install go-template-sh

```bash
go install github.com/anwam/go-template-sh@latest
```

Or build from source:

```bash
git clone https://github.com/anwam/go-template-sh.git
cd go-template-sh
go build -o go-template-sh
```

## Step 2: Generate Your Project

Run the tool:

```bash
go-template-sh
```

You'll be prompted with a series of questions. Here's an example session:

```
🚀 Welcome to go-template-sh - Go HTTP Server Template Generator

? Project name: my-api
? Go module path: github.com/myuser/my-api
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
📁 Location: ./my-api

Next steps:
  cd my-api
  go mod download
  make run
```

## Step 3: Explore Your Project

```bash
cd my-api
```

Your project structure will look like this:

```
my-api/
├── cmd/my-api/main.go           # Entry point
├── internal/
│   ├── config/config.go         # Configuration
│   ├── server/server.go         # HTTP server
│   ├── handlers/handlers.go     # Request handlers
│   ├── middleware/middleware.go # Middleware
│   ├── observability/           # Logging, tracing, metrics
│   ├── database/postgres.go     # Database connection
│   └── cache/redis.go           # Cache connection
├── Dockerfile                   # Container image
├── docker-compose.yml           # Local development services
├── Makefile                     # Build commands
├── .env.example                 # Environment variables template
└── README.md                    # Project documentation
```

## Step 4: Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```env
ENVIRONMENT=development
PORT=8080

# Database
POSTGRES_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable

# Cache
REDIS_URL=redis://localhost:6379

# Observability
OTLP_ENDPOINT=localhost:4317
SERVICE_NAME=my-api
METRICS_ENABLED=true
```

## Step 5: Start Dependencies (Docker)

If you selected Docker, start the services:

```bash
make docker-up
```

This will start:
- PostgreSQL database
- Redis cache
- Any other selected services

## Step 6: Install Go Dependencies

```bash
go mod download
```

## Step 7: Run Your Application

```bash
make run
```

You should see output like:

```
INFO Starting server port=8080
```

## Step 8: Test Your API

Open another terminal and try these endpoints:

### Health Check
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "message": "Service is healthy"
}
```

### Readiness Check
```bash
curl http://localhost:8080/ready
```

Response:
```json
{
  "status": "ready",
  "message": "Service is ready to accept traffic"
}
```

### Index
```bash
curl http://localhost:8080/
```

Response:
```json
{
  "status": "ok",
  "message": "Welcome to my-api",
  "data": {
    "version": "1.0.0",
    "environment": "development"
  }
}
```

### Metrics (if enabled)
```bash
curl http://localhost:8080/metrics
```

Response: Prometheus metrics format

## Step 9: Run Tests

```bash
make test
```

## Step 10: Build for Production

```bash
make build
```

This creates a binary at `bin/my-api`.

Run it:
```bash
./bin/my-api
```

## Step 11: Docker Build (Optional)

Build Docker image:
```bash
docker build -t my-api:latest .
```

Run container:
```bash
docker run -p 8080:8080 --env-file .env my-api:latest
```

## Development Workflow

### Adding a New Endpoint

1. Add handler to `internal/handlers/handlers.go`:

```go
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
    // Your logic here
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(Response{
        Status: "ok",
        Data: map[string]interface{}{
            "users": []string{"user1", "user2"},
        },
    })
}
```

2. Register route in `internal/server/server.go`:

For Chi:
```go
r.Get("/users", handler.GetUsers)
```

For Gin:
```go
r.GET("/users", handler.GetUsersGin)
```

3. Restart the server and test:
```bash
curl http://localhost:8080/users
```

### Using the Database

Example PostgreSQL query:

```go
import "github.com/jackc/pgx/v5/pgxpool"

func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
    rows, err := h.db.Pool().Query(r.Context(), "SELECT * FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    
    // Process rows...
}
```

### Using the Cache

Example Redis operations:

```go
import "github.com/redis/go-redis/v9"

func (h *Handler) CacheExample(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Set value
    err := h.cache.Client().Set(ctx, "key", "value", 5*time.Minute).Err()
    
    // Get value
    val, err := h.cache.Client().Get(ctx, "key").Result()
}
```

### Adding Tracing

Tracing is automatically added to all HTTP requests. To add custom spans:

```go
import "go.opentelemetry.io/otel"

func (h *Handler) TracedOperation(w http.ResponseWriter, r *http.Request) {
    tracer := otel.Tracer("my-service")
    ctx, span := tracer.Start(r.Context(), "custom-operation")
    defer span.End()
    
    // Your operation with ctx
}
```

### Adding Metrics

Add custom metrics:

```go
import "github.com/prometheus/client_golang/prometheus/promauto"

var (
    customCounter = promauto.NewCounter(prometheus.CounterOpts{
        Name: "my_custom_counter",
        Help: "Description of counter",
    })
)

func someFunction() {
    customCounter.Inc()
}
```

## Deployment

### Building for Production

```bash
# Build binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o my-api ./cmd/my-api

# Or use Makefile
make build
```

### Environment Variables

Ensure all required environment variables are set in production:

- `ENVIRONMENT=production`
- `PORT=8080`
- Database connection strings
- OTLP endpoint for tracing
- Service name

### Docker Deployment

```bash
# Build image
docker build -t my-api:v1.0.0 .

# Push to registry
docker tag my-api:v1.0.0 registry.example.com/my-api:v1.0.0
docker push registry.example.com/my-api:v1.0.0

# Deploy
docker run -d \
  -p 8080:8080 \
  --env-file .env.production \
  --name my-api \
  registry.example.com/my-api:v1.0.0
```

### Kubernetes Deployment

Example deployment manifest:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-api
  template:
    metadata:
      labels:
        app: my-api
    spec:
      containers:
      - name: my-api
        image: registry.example.com/my-api:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: PORT
          value: "8080"
        # Add other env vars...
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

## Troubleshooting

### Application won't start

1. Check environment variables:
   ```bash
   cat .env
   ```

2. Verify database connectivity:
   ```bash
   make docker-up
   docker ps
   ```

3. Check logs for errors

### Database connection failed

1. Ensure database is running:
   ```bash
   docker ps | grep postgres
   ```

2. Test connection:
   ```bash
   psql postgres://user:password@localhost:5432/dbname
   ```

3. Check connection string in `.env`

### Port already in use

Change the port in `.env`:
```env
PORT=8081
```

## Next Steps

- Read the generated `README.md` in your project
- Explore the twelve-factor app principles
- Add your business logic
- Write tests
- Set up CI/CD
- Deploy to production

## Need Help?

- [Full Documentation](README.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Open an Issue](https://github.com/anwam/go-template-sh/issues)
