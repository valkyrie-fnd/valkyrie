# OpenTelemetry Collector

An OpenTelemetry Collector is used to aggregate tracing/metrics data and export it to a wide variety of configurable
observability systems (jaeger, grafana, google cloud etc.).

Valkyrie sends its tracing/metric data in OpenTelemetry Protocol (OTLP) format over HTTP to the Collector.

You can test running a local setup of otel-collector using:

```bash
docker-compose -f docker-compose-local.yaml up
```

The local also includes:

* Jaeger for visualising traces exported by otel-collector
* Prometheus for collecting metrics from otel-collector
