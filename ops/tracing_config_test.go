package ops

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

func TestTracing(t *testing.T) {
	tests := []struct {
		name   string
		config *configs.ValkyrieConfig
		want   *TracingConfig
	}{
		{
			name: "OLTP config selected",
			config: &configs.ValkyrieConfig{
				Telemetry: configs.TelemetryConfig{
					Tracing: configs.TraceConfig{TraceType: "otlptracehttp"},
				},
			},
			want: &TracingConfig{
				Exporter: OTLPTraceHTTP,
				TraceConfig: configs.TraceConfig{
					TraceType: "otlptracehttp",
				},
			},
		},
		{
			name:   "missing config turns off tracing",
			config: &configs.ValkyrieConfig{},
			want:   &noTracingConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Tracing(tt.config))
		})
	}
}

func Test_getOTLPTraceOptions(t *testing.T) {
	tests := []struct {
		name string
		cfg  *TracingConfig
		// otlptracehttp.Option uses internal struct otlpconfig.Config, making testing hopeless
		// just assert num options for now
		want int
	}{
		{
			name: "empty url",
			cfg:  &TracingConfig{},
			want: 2,
		},
		{
			name: "https endpoint path",
			cfg: &TracingConfig{
				TraceConfig: configs.TraceConfig{
					URL: "https://test/foo",
				},
			},
			want: 3,
		},
		{
			name: "http endpoint path",
			cfg: &TracingConfig{
				TraceConfig: configs.TraceConfig{
					URL: "http://test/foo",
				},
			},
			want: 4,
		},
		{
			name: "https endpoint",
			cfg: &TracingConfig{
				TraceConfig: configs.TraceConfig{
					URL: "https://test",
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, len(getOTLPTraceOptions(tt.cfg)), "getOTLPTraceOptions(%v)", tt.cfg)
		})
	}
}
