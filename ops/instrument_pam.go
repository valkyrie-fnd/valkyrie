package ops

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/valkyrie-fnd/valkyrie/internal"
	"github.com/valkyrie-fnd/valkyrie/pam"
)

const (
	VPluginName    = "pam-vplugin-client"
	GenericPAMName = "genericpam-client"
	RPCSystem      = "net/rpc"
	RPCService     = "vplugin.PluginPAM"
)

var rpcAttributes = []attribute.KeyValue{semconv.RPCSystemKey.String(RPCSystem), semconv.RPCService(RPCService)}

// InstrumentVPluginPAMClient will instrument a VPlugin-based pipeline with telemetry handlers
func InstrumentVPluginPAMClient(pipeline *internal.Pipeline[any]) {
	pipeline.Register(PAMTracingHandler(VPluginName, rpcAttributes...),
		ApplyTracingFromContextHandler(),
		PAMMetricHandler(VPluginName, rpcAttributes...))
}

// InstrumentGenericPAMClient will instrument a genericpam-based pipeline with telemetry handlers
func InstrumentGenericPAMClient(pipeline *internal.Pipeline[any]) {
	pipeline.Register(PAMMetricHandler(GenericPAMName))
}

func PAMMetricHandler(name string, attributes ...attribute.KeyValue) internal.Handler[any] {
	const (
		metricNamePAMClientDuration = "pam.client.duration"
		metricNamePAMClientActive   = "pam.client.active_requests"
		metricNamePAMClientErrors   = "pam.client.errors"
	)
	var noopHandler internal.Handler[any] = func(pc internal.PipelineContext[any]) error {
		return pc.Next()
	}

	pamClientDuration, err := global.Meter(name).Int64Histogram(metricNamePAMClientDuration,
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription("measures the duration for outbound PAM client requests"))
	if err != nil {
		return noopHandler
	}

	pamClientActive, err := global.Meter(name).Int64UpDownCounter(metricNamePAMClientActive,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of concurrent PAM client requests currently in-flight"))
	if err != nil {
		return noopHandler
	}

	pamClientErrors, err := global.Meter(name).Int64Counter(metricNamePAMClientErrors,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("measures the number of requests with errors from the PAM client"))
	if err != nil {
		return noopHandler
	}

	return func(pc internal.PipelineContext[any]) error {
		attrs := make([]attribute.KeyValue, len(attributes))
		copy(attrs, attributes)
		attrs = append(attrs, semconv.EventName(getRequestName(pc.Payload())))

		start := time.Now()
		pamClientActive.Add(pc.Context(), 1, attrs...)

		err := pc.Next()

		pamClientDuration.Record(pc.Context(), time.Since(start).Milliseconds(), attrs...)
		pamClientActive.Add(pc.Context(), -1, attrs...)
		if err != nil {
			pamClientErrors.Add(pc.Context(), 1, attrs...)
		}

		return err
	}
}

func PAMTracingHandler(tracerName string, attributes ...attribute.KeyValue) internal.Handler[any] {
	return func(pc internal.PipelineContext[any]) error {
		ctx, span := otel.Tracer(tracerName).Start(pc.Context(), getRequestName(pc.Payload()), trace.WithAttributes(attributes...))
		defer span.End()

		pc.SetContext(ctx)

		return pc.Next()
	}
}

// getRequestName, return "GetSession" from "GetBalanceRequest"
func getRequestName(req any) string {
	name, _ := strings.CutSuffix(reflect.TypeOf(req).Name(), "Request")
	return name
}

var ErrorUnknownRequest = errors.New("unknown request type")

// ApplyTracingFromContextHandler applies traceparent and tracestate explicitly to the request, since
// vplugin doesn't rely on http client that can propagate it via headers.
func ApplyTracingFromContextHandler() internal.Handler[any] {
	return func(pc internal.PipelineContext[any]) error {
		req := pc.Payload()
		ctx := pc.Context()

		tp, ts := getTracingFromContext(ctx)

		switch r := req.(type) {
		case *pam.GetSessionRequest:
			r.Params.Traceparent, r.Params.Tracestate = tp, ts
		case *pam.RefreshSessionRequest:
			r.Params.Traceparent, r.Params.Tracestate = tp, ts
		case *pam.GetBalanceRequest:
			r.Params.Traceparent, r.Params.Tracestate = tp, ts
		case *pam.GetTransactionsRequest:
			r.Params.Traceparent, r.Params.Tracestate = tp, ts
		case *pam.AddTransactionRequest:
			r.Params.Traceparent, r.Params.Tracestate = tp, ts
		case *pam.GetGameRoundRequest:
			r.Params.Traceparent, r.Params.Tracestate = tp, ts
		default:
			return ErrorUnknownRequest
		}

		return pc.Next()
	}
}

func getTracingFromContext(ctx context.Context) (traceparent *pam.Traceparent, tracestate *pam.Tracestate) {
	tracingHeaders := GetTracingHeaders(ctx)

	if value, found := tracingHeaders["traceparent"]; found {
		traceparent = &value
	}

	if value, found := tracingHeaders["tracestate"]; found {
		tracestate = &value
	}

	return traceparent, tracestate
}
