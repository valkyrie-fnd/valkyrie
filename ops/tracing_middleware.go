package ops

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	google "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

// TracingMiddleware adds tracing
func TracingMiddleware(cfg *TracingConfig, apps ...*fiber.App) {
	if err := ConfigureTracing(cfg); err != nil {
		return
	}

	for _, app := range apps {
		app.Use(filterPath("/ping", otelfiber.Middleware(otelfiber.WithServerName(cfg.ServiceName))))
	}

	if cfg.Exporter == Google {
		for _, app := range apps {
			app.Use(googleTraceLogging(cfg.GoogleProjectID))
		}
	}
}

func ConfigureTracing(cfg *TracingConfig) error {
	// No config - no setup
	if *cfg == noTracingConfig {
		return errors.New("no tracing config")
	}

	exp, err := createProviderExporter(cfg)
	if err != nil {
		log.Warn().Err(err).Msg("failed to setup tracing provider")
		return err
	}

	// Always be sure to batch in production.
	bsp := tracesdk.NewBatchSpanProcessor(exp)

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSpanProcessor(bsp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
		// Set sampling based on upstream
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(cfg.SampleRatio))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

func googleTraceLogging(projectID string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Logger with googleLoggingHook registered
		logger := log.Ctx(ctx.UserContext()).Hook(googleLoggingHook{
			spanContext: trace.SpanFromContext(ctx.UserContext()).SpanContext(),
			projectID:   projectID,
		})
		// Associate our hooked logger with the fiber.Ctx
		ctx.SetUserContext(logger.WithContext(ctx.UserContext()))

		return ctx.Next()
	}
}

type googleLoggingHook struct {
	projectID   string
	spanContext trace.SpanContext
}

// Run adds standard tracing fields supported by Google Cloud Logging:
// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
func (h googleLoggingHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	if h.spanContext.IsValid() {
		e.Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", h.projectID, h.spanContext.TraceID()))
		e.Str("logging.googleapis.com/spanId", h.spanContext.SpanID().String())
		e.Bool("logging.googleapis.com/trace_sampled", h.spanContext.IsSampled())
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

func createProviderExporter(cfg *TracingConfig) (tracesdk.SpanExporter, error) {
	var (
		exp tracesdk.SpanExporter
		err error
	)
	switch cfg.Exporter {
	case Jaeger:
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.URL)))
	case Google:
		exp, err = google.New(google.WithProjectID(cfg.GoogleProjectID))
	case StdOut:
		exp, err = stdouttrace.New()
	}

	return exp, err
}
