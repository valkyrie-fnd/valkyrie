package routes

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

var pingHandler = func(_ *fiber.Ctx) error { return nil }

// MonitoringRoutes func for mounting monitoring routes.
func MonitoringRoutes(a *fiber.App) {

	// Create routes group.
	route := a.Group("/monitoring")

	// Monitoring
	route.Get("/metrics", monitor.New(monitor.Config{Title: "Valkyrie Metrics"}))

	// Pprof if the environment variables is present
	if _, userPprof := os.LookupEnv("PPROF"); userPprof {
		a.Use(pprof.New())
	}
}
