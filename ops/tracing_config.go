package ops

import (
	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

type ExporterType string

const (
	Jaeger ExporterType = "jaeger"
	Google ExporterType = "googleCloudTrace"
	StdOut ExporterType = "stdout"
	None   ExporterType = ""
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
	case Jaeger:
		cfg.Exporter = Jaeger
	case Google:
		cfg.Exporter = Google
	case StdOut:
		cfg.Exporter = StdOut
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
