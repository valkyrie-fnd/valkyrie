package ops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric/global"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

func Test_ConfigureMetric(t *testing.T) {
	err := ConfigureMetrics(&configs.ValkyrieConfig{
		Telemetry: configs.TelemetryConfig{
			ServiceName: "test",
		},
		Metric: configs.MetricConfig{
			ExporterType: "stdout",
		},
	})

	assert.NoError(t, err)

	testMeters := global.MeterProvider().Meter("testMeters")

	testCounter, err := testMeters.Int64Counter("testCounter")
	assert.NoError(t, err)

	testCounter.Add(context.Background(), 1)

}

func Test_metricConfig(t *testing.T) {
	vConf := &configs.ValkyrieConfig{
		Telemetry: configs.TelemetryConfig{
			ServiceName: "service",
			Namespace:   "namespace",
		},
		Metric: configs.MetricConfig{
			ExporterType: "otlpmetrichttp",
		},
		Version: "0.1.1",
	}
	expectedMetricConfig := MetricConfig{
		Exporter: MetricOTLPHTTP,
		Version:  "0.1.1",
		MetricConfig: configs.MetricConfig{
			ExporterType: "otlpmetrichttp",
		},
		TelemetryConfig: configs.TelemetryConfig{
			ServiceName: "service",
			Namespace:   "namespace",
		},
	}

	metricConfig := metricConfig(vConf)

	assert.Equal(t, expectedMetricConfig, *metricConfig)

}

func Test_getOTLPMetricOptions(t *testing.T) {
	tests := []struct {
		name string
		cfg  *MetricConfig
		// otlpmetrichttp.Option uses internal struct Config, making testing hopeless
		// just assert num options for now
		want int
	}{
		{
			name: "empty url",
			cfg:  &MetricConfig{},
			want: 2,
		},
		{
			name: "https endpoint path",
			cfg: &MetricConfig{
				MetricConfig: configs.MetricConfig{
					URL: "https://test/foo",
				},
			},
			want: 3,
		},
		{
			name: "http endpoint path",
			cfg: &MetricConfig{
				MetricConfig: configs.MetricConfig{
					URL: "http://test/foo",
				},
			},
			want: 4,
		},
		{
			name: "https endpoint",
			cfg: &MetricConfig{
				MetricConfig: configs.MetricConfig{
					URL: "https://test",
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, len(getOTLPMetricOptions(tt.cfg)), "getOTLPMetricOptions(%v)", tt.cfg)
		})
	}
}
