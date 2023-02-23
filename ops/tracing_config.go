package ops

import (
	"context"
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

type ExporterType string

const (
	StdOut        ExporterType = "stdout"
	OTLPTraceHTTP ExporterType = "otlptracehttp"
	None          ExporterType = ""
)

// noTracingConfig default empty TracingConfig
var noTracingConfig = TracingConfig{}

type TracingConfig struct {
	Exporter ExporterType
	Version  string
	configs.TelemetryConfig
	configs.TraceConfig
}

// Tracing returns a TracingConfig based on the provided Valkyrie config
func Tracing(vConf *configs.ValkyrieConfig) *TracingConfig {
	cfg := TracingConfig{}
	cfg.TraceConfig = vConf.Tracing
	cfg.Version = vConf.Version
	cfg.TelemetryConfig = vConf.Telemetry

	switch ExporterType(vConf.Tracing.TraceType) {
	case StdOut:
		cfg.Exporter = StdOut
	case OTLPTraceHTTP:
		cfg.Exporter = OTLPTraceHTTP
	case None:
		cfg.Exporter = None
		return &noTracingConfig
	default:
		log.Warn().Msgf("unsupported tracing type [%s]", vConf.Tracing.TraceType)
		return &noTracingConfig
	}

	log.Info().Str("traceType", cfg.TraceType).Msg("Configured tracing")

	return &cfg
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
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceNamespace(cfg.Namespace),
			semconv.ServiceVersion(cfg.Version),
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

func createProviderExporter(cfg *TracingConfig) (tracesdk.SpanExporter, error) {
	var (
		exp tracesdk.SpanExporter
		err error
	)
	switch cfg.Exporter {
	case OTLPTraceHTTP:
		exp, err = otlptracehttp.New(context.Background(), getOTLPTraceOptions(cfg)...)
	case StdOut:
		exp, err = stdouttrace.New()
	}

	return exp, err
}

// getOTLPTraceOptions returns options given a tracing config
func getOTLPTraceOptions(cfg *TracingConfig) []otlptracehttp.Option {
	options := []otlptracehttp.Option{
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression), // enable compression by default
	}
	url := cfg.URL

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
