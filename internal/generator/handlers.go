package generator

import (
	"fmt"
	"strings"
)

func (g *Generator) generateHandlers() error {
	content := g.getHandlersContent()
	return g.writeFile("internal/handlers/handlers.go", content)
}

func (g *Generator) getHandlersContent() string {
	imports := []string{
		`"encoding/json"`,
		`"net/http"`,
		fmt.Sprintf(`"%s/internal/config"`, g.config.ModulePath),
		fmt.Sprintf(`"%s/internal/observability"`, g.config.ModulePath),
	}

	if g.config.Framework == "gin" {
		imports = append(imports, `"github.com/gin-gonic/gin"`)
	} else if g.config.Framework == "echo" {
		imports = append(imports, `"github.com/labstack/echo/v4"`)
	} else if g.config.Framework == "fiber" {
		imports = append(imports, `"github.com/gofiber/fiber/v2"`)
	}

	if g.config.EnableMetrics {
		imports = append(imports, `"github.com/prometheus/client_golang/prometheus/promhttp"`)
	}

	frameworkHandlers := g.getFrameworkSpecificHandlers()

	return fmt.Sprintf(`package handlers

import (
	%s
)

type Handler struct {
	config *config.Config
	obs    *observability.Observability
}

func NewHandler(cfg *config.Config, obs *observability.Observability) *Handler {
	return &Handler{
		config: cfg,
		obs:    obs,
	}
}

type Response struct {
	Status  string                 ` + "`json:\"status\"`" + `
	Message string                 ` + "`json:\"message,omitempty\"`" + `
	Data    map[string]interface{} ` + "`json:\"data,omitempty\"`" + `
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Status:  "ok",
		Message: "Service is healthy",
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Status:  "ready",
		Message: "Service is ready to accept traffic",
	})
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Status:  "ok",
		Message: "Welcome to %s",
		Data: map[string]interface{}{
			"version":     "1.0.0",
			"environment": h.config.Environment,
		},
	})
}

%s
`, strings.Join(imports, "\n\t"), g.config.ProjectName, frameworkHandlers)
}

func (g *Generator) getFrameworkSpecificHandlers() string {
	switch g.config.Framework {
	case "gin":
		return g.getGinHandlers()
	case "echo":
		return g.getEchoHandlers()
	case "fiber":
		return g.getFiberHandlers()
	default:
		return ""
	}
}

func (g *Generator) getGinHandlers() string {
	metricsHandler := ""
	if g.config.EnableMetrics {
		metricsHandler = `
func (h *Handler) MetricsGin(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}`
	}

	return fmt.Sprintf(`func (h *Handler) HealthGin(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "Service is healthy",
	})
}

func (h *Handler) ReadyGin(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:  "ready",
		Message: "Service is ready to accept traffic",
	})
}

func (h *Handler) IndexGin(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "Welcome to %s",
		Data: map[string]interface{}{
			"version":     "1.0.0",
			"environment": h.config.Environment,
		},
	})
}
%s`, g.config.ProjectName, metricsHandler)
}

func (g *Generator) getEchoHandlers() string {
	metricsHandler := ""
	if g.config.EnableMetrics {
		metricsHandler = `
func (h *Handler) MetricsEcho(c echo.Context) error {
	promhttp.Handler().ServeHTTP(c.Response(), c.Request())
	return nil
}`
	}

	return fmt.Sprintf(`func (h *Handler) HealthEcho(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "Service is healthy",
	})
}

func (h *Handler) ReadyEcho(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Status:  "ready",
		Message: "Service is ready to accept traffic",
	})
}

func (h *Handler) IndexEcho(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "Welcome to %s",
		Data: map[string]interface{}{
			"version":     "1.0.0",
			"environment": h.config.Environment,
		},
	})
}
%s`, g.config.ProjectName, metricsHandler)
}

func (g *Generator) getFiberHandlers() string {
	metricsHandler := ""
	if g.config.EnableMetrics {
		metricsHandler = `
func (h *Handler) MetricsFiber(c *fiber.Ctx) error {
	return c.SendString("Metrics endpoint")
}`
	}

	return fmt.Sprintf(`func (h *Handler) HealthFiber(c *fiber.Ctx) error {
	return c.JSON(Response{
		Status:  "ok",
		Message: "Service is healthy",
	})
}

func (h *Handler) ReadyFiber(c *fiber.Ctx) error {
	return c.JSON(Response{
		Status:  "ready",
		Message: "Service is ready to accept traffic",
	})
}

func (h *Handler) IndexFiber(c *fiber.Ctx) error {
	return c.JSON(Response{
		Status:  "ok",
		Message: "Welcome to %s",
		Data: map[string]interface{}{
			"version":     "1.0.0",
			"environment": h.config.Environment,
		},
	})
}
%s`, g.config.ProjectName, metricsHandler)
}
