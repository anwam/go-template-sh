package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateMiddleware() error {
	content := g.getMiddlewareContent()
	return g.writeFile("internal/middleware/middleware.go", content)
}

func (g *Generator) getMiddlewareContent() string {
	imports := []string{
		`"context"`,
		`"net/http"`,
		`"time"`,
	}

	loggerType := "interface{}"
	switch g.config.Logger {
	case "slog":
		imports = append(imports, `"log/slog"`)
		loggerType = "*slog.Logger"
	case "zap":
		imports = append(imports, `"go.uber.org/zap"`)
		loggerType = "*zap.Logger"
	case "zerolog":
		imports = append(imports, `"github.com/rs/zerolog"`)
		loggerType = "*zerolog.Logger"
	}

	if g.config.Framework == "gin" {
		imports = append(imports, `"github.com/gin-gonic/gin"`)
	} else if g.config.Framework == "echo" {
		imports = append(imports, `"github.com/labstack/echo/v4"`)
	} else if g.config.Framework == "fiber" {
		imports = append(imports, `"github.com/gofiber/fiber/v2"`)
	}

	if g.config.EnableTracing {
		imports = append(imports,
			`"go.opentelemetry.io/otel"`,
			`"go.opentelemetry.io/otel/trace"`,
		)
	}

	standardMiddleware := g.getStandardMiddleware(loggerType)
	frameworkMiddleware := g.getFrameworkMiddleware()
	tracingMiddleware := g.getTracingMiddlewareCode()

	return fmt.Sprintf(`package middleware

import (
	%s
	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

%s
%s
%s
`, strings.Join(imports, "\n\t"), standardMiddleware, frameworkMiddleware, tracingMiddleware)
}

func (g *Generator) getStandardMiddleware(loggerType string) string {
	loggerImpl := ""
	
	switch g.config.Logger {
	case "slog":
		loggerImpl = `	logger.Info("HTTP request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.Duration("duration", duration),
		slog.Int("status", rr.status),
	)`
	case "zap":
		loggerImpl = `	logger.Info("HTTP request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.Duration("duration", duration),
		zap.Int("status", rr.status),
	)`
	case "zerolog":
		loggerImpl = `	logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Dur("duration", duration).
		Int("status", rr.status).
		Msg("HTTP request")`
	default:
		loggerImpl = `	logger.Info("HTTP request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"duration", duration,
		"status", rr.status,
	)`
	}

	return fmt.Sprintf(`func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logger(next http.Handler, logger %s) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)
		duration := time.Since(start)
		
%s
	})
}

func Recoverer(next http.Handler, logger %s) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
`, loggerType, loggerImpl, loggerType)
}

func (g *Generator) getFrameworkMiddleware() string {
	switch g.config.Framework {
	case "chi":
		return g.getChiMiddleware()
	case "gin":
		return g.getGinMiddleware()
	case "echo":
		return g.getEchoMiddleware()
	case "fiber":
		return g.getFiberMiddleware()
	default:
		return ""
	}
}

func (g *Generator) getChiMiddleware() string {
	loggerImpl := ""
	switch g.config.Logger {
	case "slog":
		loggerImpl = `		logger.Info("HTTP request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Duration("duration", duration),
			slog.Int("status", ww.Status()),
		)`
	case "zap":
		loggerImpl = `		logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
			zap.Int("status", ww.Status()),
		)`
	case "zerolog":
		loggerImpl = `		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", duration).
			Int("status", ww.Status()).
			Msg("HTTP request")`
	default:
		loggerImpl = `		logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", duration,
			"status", ww.Status(),
		)`
	}

	loggerType := "*slog.Logger"
	if g.config.Logger == "zap" {
		loggerType = "*zap.Logger"
	} else if g.config.Logger == "zerolog" {
		loggerType = "*zerolog.Logger"
	}

	return fmt.Sprintf(`
func Logger(logger %s) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			duration := time.Since(start)
			
%s
		})
	}
}

import "github.com/go-chi/chi/v5/middleware"
`, loggerType, loggerImpl)
}

func (g *Generator) getGinMiddleware() string {
	loggerImpl := ""
	switch g.config.Logger {
	case "slog":
		loggerImpl = `		logger.Info("HTTP request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Duration("duration", duration),
			slog.Int("status", c.Writer.Status()),
		)`
	case "zap":
		loggerImpl = `		logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("duration", duration),
			zap.Int("status", c.Writer.Status()),
		)`
	case "zerolog":
		loggerImpl = `		logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Dur("duration", duration).
			Int("status", c.Writer.Status()).
			Msg("HTTP request")`
	default:
		loggerImpl = `		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"duration", duration,
			"status", c.Writer.Status(),
		)`
	}

	loggerType := "*slog.Logger"
	if g.config.Logger == "zap" {
		loggerType = "*zap.Logger"
	} else if g.config.Logger == "zerolog" {
		loggerType = "*zerolog.Logger"
	}

	return fmt.Sprintf(`
func GinLogger(logger %s) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		
%s
	}
}
`, loggerType, loggerImpl)
}

func (g *Generator) getEchoMiddleware() string {
	loggerImpl := ""
	switch g.config.Logger {
	case "slog":
		loggerImpl = `		logger.Info("HTTP request",
			slog.String("method", c.Request().Method),
			slog.String("path", c.Request().URL.Path),
			slog.Duration("duration", duration),
			slog.Int("status", c.Response().Status),
		)`
	case "zap":
		loggerImpl = `		logger.Info("HTTP request",
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
			zap.Duration("duration", duration),
			zap.Int("status", c.Response().Status),
		)`
	case "zerolog":
		loggerImpl = `		logger.Info().
			Str("method", c.Request().Method).
			Str("path", c.Request().URL.Path).
			Dur("duration", duration).
			Int("status", c.Response().Status).
			Msg("HTTP request")`
	default:
		loggerImpl = `		logger.Info("HTTP request",
			"method", c.Request().Method,
			"path", c.Request().URL.Path,
			"duration", duration,
			"status", c.Response().Status,
		)`
	}

	loggerType := "*slog.Logger"
	if g.config.Logger == "zap" {
		loggerType = "*zap.Logger"
	} else if g.config.Logger == "zerolog" {
		loggerType = "*zerolog.Logger"
	}

	return fmt.Sprintf(`
func EchoLogger(logger %s) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)
			
%s
			return err
		}
	}
}
`, loggerType, loggerImpl)
}

func (g *Generator) getFiberMiddleware() string {
	loggerImpl := ""
	switch g.config.Logger {
	case "slog":
		loggerImpl = `		logger.Info("HTTP request",
			slog.String("method", string(c.Request().Method())),
			slog.String("path", c.Path()),
			slog.Duration("duration", duration),
			slog.Int("status", c.Response().StatusCode()),
		)`
	case "zap":
		loggerImpl = `		logger.Info("HTTP request",
			zap.String("method", string(c.Request().Method())),
			zap.String("path", c.Path()),
			zap.Duration("duration", duration),
			zap.Int("status", c.Response().StatusCode()),
		)`
	case "zerolog":
		loggerImpl = `		logger.Info().
			Str("method", string(c.Request().Method())).
			Str("path", c.Path()).
			Dur("duration", duration).
			Int("status", c.Response().StatusCode()).
			Msg("HTTP request")`
	default:
		loggerImpl = `		logger.Info("HTTP request",
			"method", string(c.Request().Method()),
			"path", c.Path(),
			"duration", duration,
			"status", c.Response().StatusCode(),
		)`
	}

	loggerType := "*slog.Logger"
	if g.config.Logger == "zap" {
		loggerType = "*zap.Logger"
	} else if g.config.Logger == "zerolog" {
		loggerType = "*zerolog.Logger"
	}

	return fmt.Sprintf(`
func FiberLogger(logger %s) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		
%s
		return err
	}
}
`, loggerType, loggerImpl)
}

func (g *Generator) getTracingMiddlewareCode() string {
	if !g.config.EnableTracing {
		return ""
	}

	switch g.config.Framework {
	case "chi":
		return `
func Tracing(tp trace.TracerProvider) func(next http.Handler) http.Handler {
	tracer := tp.Tracer("http-server")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
			defer span.End()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}`
	case "gin":
		return `
func GinTracing(tp trace.TracerProvider) gin.HandlerFunc {
	tracer := tp.Tracer("http-server")
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}`
	case "echo":
		return `
func EchoTracing(tp trace.TracerProvider) echo.MiddlewareFunc {
	tracer := tp.Tracer("http-server")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, span := tracer.Start(c.Request().Context(), c.Request().Method+" "+c.Request().URL.Path)
			defer span.End()
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}`
	case "fiber":
		return `
func FiberTracing(tp trace.TracerProvider) fiber.Handler {
	tracer := tp.Tracer("http-server")
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(context.Background(), string(c.Request().Method())+" "+c.Path())
		defer span.End()
		c.Locals("trace_ctx", ctx)
		return c.Next()
	}
}`
	default:
		return `
func Tracing(next http.Handler, tp trace.TracerProvider) http.Handler {
	tracer := tp.Tracer("http-server")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}`
	}
}
