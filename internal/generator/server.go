package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateServerPackage() error {
	content := g.getServerContent()
	return g.writeFile("internal/server/server.go", content)
}

func (g *Generator) getServerContent() string {
	switch g.config.Framework {
	case "chi":
		return g.generateChiServer()
	case "gin":
		return g.generateGinServer()
	case "echo":
		return g.generateEchoServer()
	case "fiber":
		return g.generateFiberServer()
	default:
		return g.generateStdlibServer()
	}
}

func (g *Generator) generateStdlibServer() string {
	imports := []string{
		`"context"`,
		`"fmt"`,
		`"net/http"`,
		`"time"`,
		fmt.Sprintf(`"%s/internal/config"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/handlers"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/middleware"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/observability"`, g.config.ModulePath),
	}

	return fmt.Sprintf(`package server

import (
	%s
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	obs        *observability.Observability
}

func New(cfg *config.Config, obs *observability.Observability) (*Server, error) {
	s := &Server{
		config: cfg,
		obs:    obs,
	}

	mux := http.NewServeMux()
	
	handler := handlers.NewHandler(cfg, obs)
	
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/ready", handler.Ready)
	mux.HandleFunc("/", handler.Index)
%s

	var h http.Handler = mux
	h = middleware.RequestID(h)
	h = middleware.Logger(h, obs.Logger)
	h = middleware.Recoverer(h, obs.Logger)
%s

	s.httpServer = &http.Server{
		Addr:         ":" + %s,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
`, strings.Join(imports, "\n\t"), g.getMetricsRoute(), g.getTracingMiddleware(), g.getConfigFieldReference("Port"))
}

func (g *Generator) generateChiServer() string {
	imports := []string{
		`"context"`,
		`"fmt"`,
		`"net/http"`,
		`"time"`,
		`"github.com/go-chi/chi/v5"`,
		`"github.com/go-chi/chi/v5/middleware"`,
		fmt.Sprintf(`"%s/internal/config"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/handlers"`, g.config.ModulePath),
		fmt.Sprintf(`custommw "%s/internal/middleware"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/observability"`, g.config.ModulePath),
	}

	return fmt.Sprintf(`package server

import (
	%s
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	obs        *observability.Observability
}

func New(cfg *config.Config, obs *observability.Observability) (*Server, error) {
	s := &Server{
		config: cfg,
		obs:    obs,
	}

	r := chi.NewRouter()
	
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(custommw.Logger(obs.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
%s

	handler := handlers.NewHandler(cfg, obs)
	
	r.Get("/health", handler.Health)
	r.Get("/ready", handler.Ready)
	r.Get("/", handler.Index)
%s

	s.httpServer = &http.Server{
		Addr:         ":" + %s,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
`, strings.Join(imports, "\n\t"), g.getTracingMiddleware(), g.getMetricsRoute(), g.getConfigFieldReference("Port"))
}

func (g *Generator) generateGinServer() string {
	imports := []string{
		`"context"`,
		`"fmt"`,
		`"net/http"`,
		`"time"`,
		`"github.com/gin-gonic/gin"`,
		fmt.Sprintf(`"%s/internal/config"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/handlers"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/middleware"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/observability"`, g.config.ModulePath),
	}

	envCheck := fmt.Sprintf("if %s == \"production\" {", g.getConfigFieldReference("Environment"))

	return fmt.Sprintf(`package server

import (
	%s
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	obs        *observability.Observability
}

func New(cfg *config.Config, obs *observability.Observability) (*Server, error) {
	%s
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		config: cfg,
		obs:    obs,
	}

	r := gin.New()
	
	r.Use(gin.Recovery())
	r.Use(middleware.GinLogger(obs.Logger))
%s

	handler := handlers.NewHandler(cfg, obs)
	
	r.GET("/health", handler.HealthGin)
	r.GET("/ready", handler.ReadyGin)
	r.GET("/", handler.IndexGin)
%s

	s.httpServer = &http.Server{
		Addr:         ":" + %s,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
`, strings.Join(imports, "\n\t"), envCheck, g.getTracingMiddleware(), g.getGinMetricsRoute(), g.getConfigFieldReference("Port"))
}

func (g *Generator) generateEchoServer() string {
	imports := []string{
		`"context"`,
		`"net/http"`,
		`"time"`,
		`"github.com/labstack/echo/v4"`,
		`"github.com/labstack/echo/v4/middleware"`,
		fmt.Sprintf(`"%s/internal/config"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/handlers"`, g.config.ModulePath),
		fmt.Sprintf(`custommw "%s/internal/middleware"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/observability"`, g.config.ModulePath),
	}

	return fmt.Sprintf(`package server

import (
	%s
)

type Server struct {
	echo   *echo.Echo
	config *config.Config
	obs    *observability.Observability
}

func New(cfg *config.Config, obs *observability.Observability) (*Server, error) {
	s := &Server{
		config: cfg,
		obs:    obs,
		echo:   echo.New(),
	}

	s.echo.Use(middleware.RequestID())
	s.echo.Use(middleware.Recover())
	s.echo.Use(custommw.EchoLogger(obs.Logger))
	s.echo.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 60 * time.Second,
	}))
%s

	handler := handlers.NewHandler(cfg, obs)
	
	s.echo.GET("/health", handler.HealthEcho)
	s.echo.GET("/ready", handler.ReadyEcho)
	s.echo.GET("/", handler.IndexEcho)
%s

	s.echo.Server.ReadTimeout = 15 * time.Second
	s.echo.Server.WriteTimeout = 15 * time.Second
	s.echo.Server.IdleTimeout = 60 * time.Second

	return s, nil
}

func (s *Server) Start() error {
	return s.echo.Start(":" + %s)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
`, strings.Join(imports, "\n\t"), g.getTracingMiddleware(), g.getEchoMetricsRoute(), g.getConfigFieldReference("Port"))
}

func (g *Generator) generateFiberServer() string {
	imports := []string{
		`"context"`,
		`"fmt"`,
		`"time"`,
		`"github.com/gofiber/fiber/v2"`,
		`"github.com/gofiber/fiber/v2/middleware/recover"`,
		fmt.Sprintf(`"%s/internal/config"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/handlers"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/middleware"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/observability"`, g.config.ModulePath),
	}

	return fmt.Sprintf(`package server

import (
	%s
)

type Server struct {
	app    *fiber.App
	config *config.Config
	obs    *observability.Observability
}

func New(cfg *config.Config, obs *observability.Observability) (*Server, error) {
	s := &Server{
		config: cfg,
		obs:    obs,
	}

	s.app = fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	s.app.Use(recover.New())
	s.app.Use(middleware.FiberLogger(obs.Logger))
%s

	handler := handlers.NewHandler(cfg, obs)
	
	s.app.Get("/health", handler.HealthFiber)
	s.app.Get("/ready", handler.ReadyFiber)
	s.app.Get("/", handler.IndexFiber)
%s

	return s, nil
}

func (s *Server) Start() error {
	return s.app.Listen(":" + %s)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
`, strings.Join(imports, "\n\t"), g.getTracingMiddleware(), g.getFiberMetricsRoute(), g.getConfigFieldReference("Port"))
}

func (g *Generator) getMetricsRoute() string {
	if g.config.EnableMetrics {
		return `	mux.Handle("/metrics", obs.MetricsHandler())`
	}
	return ""
}

func (g *Generator) getGinMetricsRoute() string {
	if g.config.EnableMetrics {
		return `
	r.GET("/metrics", gin.WrapH(obs.MetricsHandler()))`
	}
	return ""
}

func (g *Generator) getEchoMetricsRoute() string {
	if g.config.EnableMetrics {
		return `
	s.echo.GET("/metrics", echo.WrapHandler(obs.MetricsHandler()))`
	}
	return ""
}

func (g *Generator) getFiberMetricsRoute() string {
	if g.config.EnableMetrics {
		return `
	s.app.Get("/metrics", handler.MetricsFiber)`
	}
	return ""
}

func (g *Generator) getTracingMiddleware() string {
	if g.config.EnableTracing {
		switch g.config.Framework {
		case "chi":
			return `	r.Use(custommw.Tracing(obs.TracerProvider))`
		case "gin":
			return `	r.Use(middleware.GinTracing(obs.TracerProvider))`
		case "echo":
			return `	s.echo.Use(custommw.EchoTracing(obs.TracerProvider))`
		case "fiber":
			return `	s.app.Use(middleware.FiberTracing(obs.TracerProvider))`
		default:
			return `	h = middleware.Tracing(h, obs.TracerProvider)`
		}
	}
	return ""
}
