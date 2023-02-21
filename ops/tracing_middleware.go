package ops

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
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

	if cfg.GoogleProjectID != "" {
		for _, app := range apps {
			app.Use(googleTraceLogging(cfg.GoogleProjectID))
		}
	}
}

// ConfigureTracing configures the tracing framework based on TracingConfig.
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

	if cfg.GoogleProjectID != "" {
		// Add googleErrorHook to global logger, so that it's inherited by child loggers
		log.Logger = log.Hook(googleErrorHook{})
	}

	return nil
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

func createProviderExporter(cfg *TracingConfig) (tracesdk.SpanExporter, error) {
	var (
		exp tracesdk.SpanExporter
		err error
	)
	switch cfg.Exporter {
	case OTLPTraceHTTP:
		exp, err = otlptracehttp.New(context.Background(), getOTLPOptions(cfg.URL)...)
	case StdOut:
		exp, err = stdouttrace.New()
	}

	return exp, err
}

// getOTLPOptions returns options given a tracing URL
func getOTLPOptions(url string) []otlptracehttp.Option {
	options := []otlptracehttp.Option{
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression), // enable compression by default
	}

	if url == "" {
		options = append(options, otlptracehttp.WithInsecure()) // use HTTP by default
		return options
	}

	if scheme, remainder, found := strings.Cut(url, "://"); found {
		if scheme == "http" {
			options = append(options, otlptracehttp.WithInsecure())
		}
		url = remainder
	}

	if endpoint, path, found := strings.Cut(url, "/"); found {
		options = append(options, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithURLPath(path))
	} else {
		options = append(options, otlptracehttp.WithEndpoint(endpoint))
	}

	return options
}
