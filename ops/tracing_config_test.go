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
			name: "Jaeger config selected",
			want: &TracingConfig{
				Exporter: Jaeger,
				TraceConfig: configs.TraceConfig{
					TraceType:   "jaeger",
					URL:         "the.jaeger.host:222/path",
					ServiceName: "my-service",
				},
			},
		},
		{
			name: "Google config selected",
			want: &TracingConfig{
				Exporter: Google,
				TraceConfig: configs.TraceConfig{
					TraceType:       "googleCloudTrace",
					ServiceName:     "my-service",
					GoogleProjectID: "my-project",
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
