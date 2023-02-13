package provider

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type Registry struct {
	app      *fiber.App
	routes   map[string]*Router
	basePath string
}

func NewRegistry(app *fiber.App, basePath string) *Registry {
	return &Registry{
		app:      app,
		basePath: basePath,
		routes:   make(map[string]*Router),
	}
}

// Register a provider
func (pr *Registry) Register(provider *Router) error {
	basePath := pr.basePath + provider.BasePath

	// Make sure the provider (i.e. base path) is unique
	if p, found := pr.routes[basePath]; found {
		return fmt.Errorf("Base path %s is already claimed by provider %s", basePath, p.Name)
	}

	// Reserve the base path
	pr.routes[basePath] = provider

	// Create subgroup
	group := pr.app.Group(basePath)

	// Add middlewares
	for _, m := range provider.Middlewares {
		group.Use(m)
	}

	// Add routes
	for _, r := range provider.Routes {
		switch r.Method {
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
			log.Info().Msgf("Route %s %s", r.Method, basePath+r.Path)
			group.Add(r.Method, r.Path, append(r.Middlewares, r.HandlerFunc)...)
		default:
			return fmt.Errorf("unable to configure provider %s with path %s and method %s", provider.Name, r.Path, r.Method)
		}
	}

	return nil
}
