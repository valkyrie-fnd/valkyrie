package ops

import (
	"bytes"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggingMiddleware setup zerolog and request logging
//
// # Adds tracing information to logging
//
// # Adds debug logging for request and response
func LoggingMiddleware(apps ...*fiber.App) {
	for _, app := range apps {
		app.Use(logContextInjector)

		app.Use(propagateTraceLogging)

		log.Debug().Func(func(e *zerolog.Event) {
			e.Msg("Setting up request/response logging for debug")
			app.Use(requestResponseLogging)
		})
	}
}

// logContextInjector makes sure zerolog.Logger is injected into the request context
func logContextInjector(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// log.Ctx(ctx) will return the logger already associated with the context (if already configured),
	// and if not configured the `DefaultContextLogger`, which is configured as the global logger in ConfigureLogging.
	l := log.Ctx(ctx).With().Logger()
	c.SetUserContext(l.WithContext(ctx))
	return c.Next()
}

var pathPing = []byte("/ping")

// Adds request and response to log
func requestResponseLogging(c *fiber.Ctx) error {
	path := c.Request().URI().Path()
	if !bytes.HasSuffix(path, pathPing) {
		log.Ctx(c.UserContext()).Debug().Func(LogHTTPRequest(c.Request())).Msg("http server request")
	}

	err := c.Next()

	if !bytes.HasSuffix(path, pathPing) {
		if err != nil {
			log.Ctx(c.UserContext()).Error().Func(LogHTTPResponse(c.Request(), c.Response(), err)).Msg("http server response")
		} else {
			log.Ctx(c.UserContext()).Debug().Func(LogHTTPResponse(c.Request(), c.Response(), nil)).Msg("http server response")
		}
	}

	return err
}

func propagateTraceLogging(ctx *fiber.Ctx) (err error) {
	// extract tracing information from the context and add to the logging context
	carrier := GetTracingHeaders(ctx.UserContext())

	if len(carrier) > 0 {
		ccc := log.Ctx(ctx.UserContext()).With()
		for k, v := range carrier {
			ccc = ccc.Str(k, v)
		}

		ctx.SetUserContext(ccc.Logger().WithContext(ctx.UserContext()))
	}
	return ctx.Next()
}
