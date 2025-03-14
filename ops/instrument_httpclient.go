package ops

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/valkyrie-fnd/valkyrie/internal/pipeline"
)

type FastHTTPPayload interface {
	Request() *fasthttp.Request
	Response() *fasthttp.Response
}

// InstrumentHTTPClient will instrument a fasthttp-based pipeline with telemetry handlers
func InstrumentHTTPClient[T FastHTTPPayload](pipeline *pipeline.Pipeline[T]) {
	pipeline.Register(HTTPTracingHandler[T](), HTTPLoggingHandler[T](), HTTPMetricHandler[T]())
}

func HTTPTracingHandler[T FastHTTPPayload]() pipeline.Handler[T] {
	const tracerName = "http-client"

	return func(cc pipeline.PipelineContext[T]) error {
		ctx, span := otel.Tracer(tracerName).Start(cc.Context(), string(cc.Payload().Request().URI().Path()))
		defer span.End()

		cc.SetContext(ctx)
		addTraceHeaders(ctx, &cc.Payload().Request().Header)

		err := cc.Next()

		traceHTTPAttributes(span, cc.Payload().Request(), cc.Payload().Response(), err)

		return err
	}
}

func addTraceHeaders(ctx context.Context, headers *fasthttp.RequestHeader) {
	tracingHeaders := GetTracingHeaders(ctx)

	// only propagate traceparent and tracestate, as other headers (baggage) might leak sensitive information
	// https://www.w3.org/TR/trace-context/#privacy-considerations
	if value, found := tracingHeaders["traceparent"]; found {
		headers.Add("traceparent", value)
	}

	if value, found := tracingHeaders["tracestate"]; found {
		headers.Add("tracestate", value)
	}
}

func HTTPLoggingHandler[T FastHTTPPayload]() pipeline.Handler[T] {
	return func(pc pipeline.PipelineContext[T]) error {
		log.Ctx(pc.Context()).
			Debug().
			Func(logHTTPRequest(pc.Payload().Request())).
			Msg("http client request")

		err := pc.Next()

		var event *zerolog.Event
		if err != nil {
			event = log.Ctx(pc.Context()).Error()
		} else {
			event = log.Ctx(pc.Context()).Debug()
		}
		event.Func(logHTTPResponse(pc.Payload().Request(), pc.Payload().Response(), err)).
			Msg("http client response")

		return err
	}
}

func HTTPMetricHandler[T FastHTTPPayload]() pipeline.Handler[T] {
	const (
		instrumentationName          = "http-client"
		metricNameHTTPClientDuration = "http.client.duration"
		metricNameHTTPClientActive   = "http.client.active_requests"
		metricNameHTTPClientErrors   = "http.client.errors"
	)
	var noopHandler pipeline.Handler[T] = func(c pipeline.PipelineContext[T]) error {
		return c.Next()
	}

	httpClientDuration, err := otel.Meter(instrumentationName).Int64Histogram(metricNameHTTPClientDuration,
		metric.WithUnit(unitMilliseconds),
		metric.WithDescription("measures the duration for outbound HTTP client requests"))
	if err != nil {
		return noopHandler
	}

	httpClientActive, err := otel.Meter(instrumentationName).Int64UpDownCounter(metricNameHTTPClientActive,
		metric.WithUnit(unitDimensionless),
		metric.WithDescription("measures the number of concurrent HTTP client requests currently in-flight"))
	if err != nil {
		return noopHandler
	}

	httpClientErrors, err := otel.Meter(instrumentationName).Int64Counter(metricNameHTTPClientErrors,
		metric.WithUnit(unitDimensionless),
		metric.WithDescription("measures the number of requests with errors from the HTTP client"))
	if err != nil {
		return noopHandler
	}

	return func(pc pipeline.PipelineContext[T]) error {
		attributes := httpClientReqAttributes(pc.Payload().Request())
		start := time.Now()
		httpClientActive.Add(pc.Context(), 1, metric.WithAttributes(attributes...))

		err := pc.Next()

		attributes = append(attributes, httpClientRespAttributes(pc.Payload().Response())...)

		httpClientDuration.Record(pc.Context(), time.Since(start).Milliseconds(), metric.WithAttributes(attributes...))
		httpClientActive.Add(pc.Context(), -1, metric.WithAttributes(attributes...))
		if err != nil || pc.Payload().Response().StatusCode() >= 500 {
			httpClientErrors.Add(pc.Context(), 1, metric.WithAttributes(attributes...))
		}

		return err
	}
}

func httpClientReqAttributes(req *fasthttp.Request) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.HTTPMethod(string(utils.CopyBytes(req.Header.Method()))),
		semconv.HTTPScheme(string(utils.CopyBytes(req.Header.Protocol()))),
		semconv.NetHostName(string(utils.CopyBytes(req.Host()))),
	}
}

func httpClientRespAttributes(resp *fasthttp.Response) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.HTTPStatusCode(resp.StatusCode()),
	}
}
