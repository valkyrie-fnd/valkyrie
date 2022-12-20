package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// fiberMiddleware provide Fiber's built-in middlewares.
//
// # Sets up Cors
//
// See: https://docs.gofiber.io/api/middleware
func fiberMiddleware(apps ...*fiber.App) {
	for _, app := range apps {
		app.Use(
			// Add CORS to each route.
			cors.New(),
			recover.New(recover.Config{
				EnableStackTrace:  true,
				StackTraceHandler: recoveryHandler,
			}),
		)

	}
}

func recoveryHandler(c *fiber.Ctx, e interface{}) {
	var err error
	switch v := e.(type) {
	case string:
		err = errors.New(v)
	case error:
		err = errors.WithStack(v)
	default:
		err = errors.Errorf("unknown panic: %v\n", v)
	}

	zerolog.Ctx(c.UserContext()).Error().Stack().
		Err(err).Msg("panic recovered")
}
