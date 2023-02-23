package ops

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// TracingMiddleware adds tracing
func TracingMiddleware(cfg *TracingConfig, apps ...*fiber.App) {
	if err := ConfigureTracing(cfg); err != nil {
		return
	}

	for _, app := range apps {
		app.Use(filterPath("/ping", otelfiber.Middleware(otelfiber.WithServerName(cfg.ServiceName))))
	}

	if cfg.GoogleProjectID != "" {
		for _, app := range apps {
			app.Use(googleTraceLogging(cfg.GoogleProjectID))
		}
	}
}

func googleTraceLogging(projectID string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Logger with googleTracingHook registered
		logger := log.Ctx(ctx.UserContext()).Hook(googleTracingHook{
			spanContext: trace.SpanFromContext(ctx.UserContext()).SpanContext(),
			projectID:   projectID,
		})
		// Associate our hooked logger with the fiber.Ctx
		ctx.SetUserContext(logger.WithContext(ctx.UserContext()))

		return ctx.Next()
	}
}

type googleTracingHook struct {
	projectID   string
	spanContext trace.SpanContext
}

// Run adds standard tracing fields supported by Google Cloud Logging:
// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
func (h googleTracingHook) Run(e *zerolog.Event, l zerolog.Level, _ string) {
	if h.spanContext.IsValid() {
		e.Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", h.projectID, h.spanContext.TraceID()))
		e.Str("logging.googleapis.com/spanId", h.spanContext.SpanID().String())
		e.Bool("logging.googleapis.com/trace_sampled", h.spanContext.IsSampled())
	}
}

type googleErrorHook struct{}

// Run adds standard error reporting field annotation
// https://cloud.google.com/error-reporting/docs/formatting-error-messages#reported-error-example
func (h googleErrorHook) Run(e *zerolog.Event, l zerolog.Level, _ string) {
	if l == zerolog.ErrorLevel || l == zerolog.FatalLevel || l == zerolog.PanicLevel {
		e.Str("@type", "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent")
	}
}

func filterPath(suffix string, f fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if strings.HasSuffix(c.Path(), suffix) {
			return c.Next()
		}

		return f(c)
	}
}
