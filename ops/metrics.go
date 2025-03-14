package ops

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

type MetricExporterType string

const (
	MetricStdOut   MetricExporterType = "stdout"
	MetricOTLPHTTP MetricExporterType = "otlpmetrichttp"
	MetricNone     MetricExporterType = ""

	unitDimensionless = "1"
	unitBytes         = "By"
	unitMilliseconds  = "ms"
)

// noMetricConfig default empty MetricConfig
var noMetricConfig = MetricConfig{}

type MetricConfig struct {
	Exporter    MetricExporterType
	Version     string
	ServiceName string
	Namespace   string
	configs.MetricConfig
}

// ConfigureMetrics configures metrics, including exporter and instrumentation
func ConfigureMetrics(vConf *configs.ValkyrieConfig) error {
	cfg := metricConfig(vConf)

	// No config - no setup
	if *cfg == noMetricConfig {
		return nil
	}

	exp, err := createMetricExporter(cfg)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to setup metric exporter")
		return err
	}

	// labels/tags/resources that are common to all metrics.
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceNamespace(cfg.Namespace),
		semconv.ServiceVersion(cfg.Version),
	)

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			// collects and exports metric data every 60 seconds by default.
			metric.NewPeriodicReader(exp),
		),
	)

	otel.SetMeterProvider(mp)

	log.Info().Msgf("Configured metrics using exporter=%s", cfg.Exporter)

	err = configureInstrumentation()
	if err != nil {
		return err
	}

	return nil
}

// metricConfig creates a MetricConfig from configs.ValkyrieConfig
func metricConfig(vConf *configs.ValkyrieConfig) *MetricConfig {
	cfg := MetricConfig{}

	cfg.MetricConfig = vConf.Telemetry.Metric
	cfg.Version = vConf.Version
	cfg.ServiceName = vConf.Telemetry.ServiceName
	cfg.Namespace = vConf.Telemetry.Namespace

	switch MetricExporterType(cfg.ExporterType) {
	case MetricStdOut:
		cfg.Exporter = MetricStdOut
	case MetricOTLPHTTP:
		cfg.Exporter = MetricOTLPHTTP
	case MetricNone:
		cfg.Exporter = MetricNone
		return &noMetricConfig
	default:
		log.Warn().Msgf("unsupported metric exporter type [%s]", cfg.ExporterType)
		return &noMetricConfig
	}

	return &cfg
}

// configureInstrumentation configures various instrumentation of metrics
func configureInstrumentation() error {
	log.Debug().Msg("Starting metric runtime instrumentation")
	err := runtime.Start()
	if err != nil {
		return err
	}

	return nil
}

func createMetricExporter(cfg *MetricConfig) (metric.Exporter, error) {
	var (
		exp metric.Exporter
		err error
	)
	switch cfg.Exporter {
	case MetricOTLPHTTP:
		exp, err = otlpmetrichttp.New(context.Background(), getOTLPMetricOptions(cfg)...)
	case MetricStdOut:
		exp, err = stdoutmetric.New()
	}

	return exp, err
}

// getOTLPMetricOptions returns OTLP exporter options given a metric config
func getOTLPMetricOptions(cfg *MetricConfig) []otlpmetrichttp.Option {
	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression), // enable compression by default
	}
	url := cfg.URL

	if url == "" {
		options = append(options, otlpmetrichttp.WithInsecure()) // use HTTP by default
		return options
	}

	if scheme, remainder, found := strings.Cut(url, "://"); found {
		if scheme == "http" {
			options = append(options, otlpmetrichttp.WithInsecure())
		}
		url = remainder
	}

	if endpoint, path, found := strings.Cut(url, "/"); found {
		options = append(options, otlpmetrichttp.WithEndpoint(endpoint), otlpmetrichttp.WithURLPath(path))
	} else {
		options = append(options, otlpmetrichttp.WithEndpoint(endpoint))
	}

	return options
}
