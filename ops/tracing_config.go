package ops

import (
	"github.com/rs/zerolog/log"

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
	configs.TraceConfig
}

// Tracing returns a TracingConfig based on the provided Valkyrie config vTracing
func Tracing(vTracing configs.TraceConfig) *TracingConfig {
	cfg := TracingConfig{}
	cfg.TraceConfig = vTracing
	switch ExporterType(vTracing.TraceType) {
	case StdOut:
		cfg.Exporter = StdOut
	case OTLPTraceHTTP:
		cfg.Exporter = OTLPTraceHTTP
	case None:
		cfg.Exporter = None
		return &noTracingConfig
	default:
		log.Warn().Msgf("unsupported tracing type [%s]", vTracing.TraceType)
		return &noTracingConfig
	}

	log.Info().Str("traceType", cfg.TraceType).Msg("Configured tracing")

	return &cfg
}
