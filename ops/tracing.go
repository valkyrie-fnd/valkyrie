package ops

import (
	"context"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/propagation"
)

// GetTracingHeaders returns the tracing headers
func GetTracingHeaders(ctx context.Context) map[string]string {
	// extract tracing information from the context
	carrier := propagation.MapCarrier{}

	otel.GetTextMapPropagator().Inject(ctx, &carrier)

	return carrier
}

// traceHTTPAttributes sets relevant tracing attributes based on provided fasthttp.Request and fasthttp.Response.
func traceHTTPAttributes(span trace.Span, req *fasthttp.Request, resp *fasthttp.Response, err error) {
	if err != nil {
		span.RecordError(err)
	}

	span.SetAttributes(
		// request attributes
		semconv.HTTPURLKey.String(string(utils.CopyBytes(req.URI().FullURI()))),
		semconv.HTTPMethodKey.String(string(utils.CopyBytes(req.Header.Method()))),
		semconv.HTTPRequestContentLengthKey.Int(req.Header.ContentLength()),

		// response attributes
		semconv.HTTPStatusCodeKey.Int(resp.StatusCode()),
		semconv.HTTPFlavorKey.String(string(utils.CopyBytes(resp.Header.Protocol()))),
		semconv.HTTPResponseContentLengthKey.Int(resp.Header.ContentLength()),
	)

	span.SetStatus(httpconv.ClientStatus(resp.StatusCode()))
}
