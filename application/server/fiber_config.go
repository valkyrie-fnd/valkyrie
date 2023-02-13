package server

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

// fiberConfig func for configuration Fiber app.
// See: https://docs.gofiber.io/api/fiber#config
func fiberConfig(config configs.HTTPServerConfig) fiber.Config {
	// Return Fiber configuration.
	return fiber.Config{
		EnablePrintRoutes:     false,
		DisableStartupMessage: true,
		ReadTimeout:           config.ReadTimeout,
		WriteTimeout:          config.WriteTimeout,
		IdleTimeout:           config.IdleTimeout,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	}
}
