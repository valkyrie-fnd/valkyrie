package ops

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

func TestTracing(t *testing.T) {
	tests := []struct {
		name string
		want *TracingConfig
	}{
		{
			name: "OLTP config selected",
			want: &TracingConfig{
				Exporter: OTLPTraceHTTP,
				TraceConfig: configs.TraceConfig{
					TraceType:   "otlptracehttp",
					ServiceName: "my-service",
				},
			},
		},
		{
			name: "missing config turns off tracing",
			want: &noTracingConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Tracing(tt.want.TraceConfig))
		})
	}
}
