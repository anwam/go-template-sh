package generator

// generateServerPackage generates the server package using embedded templates.
func (g *Generator) generateServerPackage() error {
	data := ServerTemplateData{
		ModulePath:    g.config.ModulePath,
		PortRef:       g.getConfigFieldReference("Port"),
		EnvRef:        g.getConfigFieldReference("Environment"),
		EnableTracing: g.config.EnableTracing,
		EnableMetrics: g.config.EnableMetrics,
	}

	templateName := g.getServerTemplateName()
	return g.writeEmbeddedTemplate("internal/server/server.go", templateName, data)
}

func (g *Generator) getServerTemplateName() string {
	switch g.config.Framework {
	case "chi":
		return "server_chi.go.tmpl"
	case "gin":
		return "server_gin.go.tmpl"
	case "echo":
		return "server_echo.go.tmpl"
	case "fiber":
		return "server_fiber.go.tmpl"
	default:
		return "server_stdlib.go.tmpl"
	}
}
