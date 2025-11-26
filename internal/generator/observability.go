package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateObservability() error {
	content := g.getObservabilityContent()
	return g.writeFile("internal/observability/observability.go", content)
}

func (g *Generator) getObservabilityContent() string {
	imports := []string{
		`"context"`,
		`"net/http"`,
		fmt.Sprintf(`"%s/internal/config"`, g.config.ModulePath),
	}

	loggerField := "Logger interface{}"
	
	switch g.config.Logger {
	case "slog":
		imports = append(imports, `"log/slog"`, `"os"`)
		loggerField = "Logger *slog.Logger"
	case "zap":
		imports = append(imports, `"go.uber.org/zap"`)
		loggerField = "Logger *zap.Logger"
	case "zerolog":
		imports = append(imports, `"github.com/rs/zerolog"`, `"os"`)
		loggerField = "Logger *zerolog.Logger"
	}

	tracerField := ""
	tracerInit := ""
	tracerShutdown := ""
	if g.config.EnableTracing {
		imports = append(imports,
			`"go.opentelemetry.io/otel"`,
			`"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"`,
			`"go.opentelemetry.io/otel/sdk/resource"`,
			`"go.opentelemetry.io/otel/sdk/trace"`,
			`semconv "go.opentelemetry.io/otel/semconv/v1.21.0"`,
		)
		tracerField = `	TracerProvider trace.TracerProvider
	tracerShutdown func(context.Context) error`
		
		tracerInit = `
	tp, shutdown, err := initTracer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	obs.TracerProvider = tp
	obs.tracerShutdown = shutdown
	otel.SetTracerProvider(tp)`
	
		tracerShutdown = `
	if o.tracerShutdown != nil {
		_ = o.tracerShutdown(ctx)
	}`
	}

	metricsField := ""
	metricsInit := ""
	metricsHandler := ""
	if g.config.EnableMetrics {
		imports = append(imports,
			`"github.com/prometheus/client_golang/prometheus"`,
			`"github.com/prometheus/client_golang/prometheus/promhttp"`,
			`"github.com/prometheus/client_golang/prometheus/promauto"`,
		)
		metricsField = `	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec`
		
		metricsInit = `
	obs.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	
	obs.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)`
		
		metricsHandler = `
func (o *Observability) MetricsHandler() http.Handler {
	return promhttp.Handler()
}`
	}

	loggerInit := g.getLoggerInitialization()
	tracerImplementation := ""
	if g.config.EnableTracing {
		tracerImplementation = g.getTracerImplementation()
	}

	return fmt.Sprintf(`package observability

import (
	%s
)

type Observability struct {
	%s
%s
%s
}

func New(ctx context.Context, cfg *config.Config) (*Observability, error) {
	obs := &Observability{}
	
%s
%s
%s

	return obs, nil
}

func (o *Observability) Shutdown(ctx context.Context) error {
%s
	return nil
}
%s
%s
`, strings.Join(imports, "\n\t"), loggerField, tracerField, metricsField, loggerInit, tracerInit, metricsInit, tracerShutdown, metricsHandler, tracerImplementation)
}

func (g *Generator) getLoggerInitialization() string {
	switch g.config.Logger {
	case "slog":
		return `	obs.Logger = NewLogger(cfg)`
	case "zap":
		return `	logger, err := NewZapLogger(cfg)
	if err != nil {
		return nil, err
	}
	obs.Logger = logger`
	case "zerolog":
		return `	obs.Logger = NewZerologLogger(cfg)`
	default:
		return `	obs.Logger = NewLogger(cfg)`
	}
}

func (g *Generator) getTracerImplementation() string {
	return `
func initTracer(ctx context.Context, cfg *config.Config) (trace.TracerProvider, func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	return tp, tp.Shutdown, nil
}`
}

func (g *Generator) generateLoggerFile() error {
	content := g.getLoggerFileContent()
	return g.writeFile("internal/observability/logger.go", content)
}

func (g *Generator) getLoggerFileContent() string {
	switch g.config.Logger {
	case "slog":
		return fmt.Sprintf(`package observability

import (
	"log/slog"
	"os"

	"%s/internal/config"
)

var defaultLogger *slog.Logger

func NewLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	if cfg.Environment == "production" {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}

func SetDefaultLogger(logger *slog.Logger) {
	defaultLogger = logger
	slog.SetDefault(logger)
}
`, g.config.ModulePath)
	
	case "zap":
		return fmt.Sprintf(`package observability

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"%s/internal/config"
)

func NewZapLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config
	
	if cfg.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return zapConfig.Build()
}
`, g.config.ModulePath)
	
	case "zerolog":
		return fmt.Sprintf(`package observability

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"%s/internal/config"
)

func NewZerologLogger(cfg *config.Config) *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	var level zerolog.Level
	if cfg.Environment == "production" {
		level = zerolog.InfoLevel
	} else {
		level = zerolog.DebugLevel
	}

	logger := zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &logger
}
`, g.config.ModulePath)
	
	default:
		return ""
	}
}
