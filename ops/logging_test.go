package ops

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

type stubUserContextProvider struct {
	userContext context.Context
}

func (s *stubUserContextProvider) UserContext() context.Context {
	return s.userContext
}

func TestAddLoggingContext(t *testing.T) {
	// given
	captor := strings.Builder{}
	testContext := context.Background()
	fiberContext := stubUserContextProvider{userContext: log.
		With().
		Logger().
		Output(&captor).
		WithContext(testContext)}

	// when
	AddLoggingContext(&fiberContext, "hip", "hop")
	AddLoggingContext(&fiberContext, "foo", "bar")

	log.Ctx(fiberContext.UserContext()).Info().Msg("test")

	// then
	assert.Contains(t, captor.String(), "foo")
	assert.Contains(t, captor.String(), "bar")
	assert.Contains(t, captor.String(), "hip")
	assert.Contains(t, captor.String(), "hop")
	assert.Contains(t, captor.String(), "test")
}

func TestLogHTTPRequest(t *testing.T) {
	captor := strings.Builder{}
	logger := zerolog.New(&captor)

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	request.Header.SetUserAgent("bar")
	request.Header.SetContentType("text/plain")
	request.Header.Set("X-Correlation-ID", "foo")

	logger.Debug().Func(logHTTPRequest(request)).Send()

	assert.Contains(t, captor.String(), "\"userAgent\":\"bar\"")
	assert.Contains(t, captor.String(), "\"X-Correlation-ID\":\"foo\"")
	assert.Contains(t, captor.String(), "\"Content-Type\":\"text/plain\"")

	assert.NotContains(t, captor.String(), "\"X-Forwarded-For\"", "X-Forwarded-For header is never set")
}

func TestHTTPRequestContentLengthNegative(t *testing.T) {
	captor := strings.Builder{}
	logger := zerolog.New(&captor)

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	response.SetBodyString("foo")
	response.Header.SetContentLength(-2)

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	request.SetBodyString("bar")
	request.Header.SetContentLength(-2)

	logger.Debug().Func(logHTTPResponse(request, response, nil)).Send()

	assert.Contains(t, captor.String(), "\"requestSize\":-2")
	assert.Contains(t, captor.String(), "\"responseSize\":-2")
}

func TestLogHTTPResponse(t *testing.T) {
	captor := strings.Builder{}
	logger := zerolog.New(&captor)

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	response.SetBodyString("foo")
	response.Header.SetContentType("text/plain")
	response.Header.Set("X-Correlation-ID", "foo")

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)

	logger.Debug().Func(logHTTPResponse(request, response, nil)).Send()

	assert.Contains(t, captor.String(), "\"response\":\"foo\"")
	assert.Contains(t, captor.String(), "\"X-Correlation-ID\":\"foo\"")
	assert.Contains(t, captor.String(), "\"Content-Type\":\"text/plain\"")

	assert.NotContains(t, captor.String(), "\"X-Forwarded-For\"", "X-Forwarded-For header is never set")
}

func TestLogHTTPResponseSkipEmptyContentType(t *testing.T) {
	captor := strings.Builder{}
	logger := zerolog.New(&captor)

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	response.Header.SetContentType("")

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	request.Header.SetContentType("")

	logger.Debug().Func(logHTTPResponse(request, response, nil)).Send()

	assert.NotContains(t, captor.String(), "\"Content-Type\":\"\"")
}

func Test_isContentTypeJson(t *testing.T) {
	assert.True(t, isContentTypeJSON([]byte(fiber.MIMEApplicationJSON)))
	assert.True(t, isContentTypeJSON([]byte("application/vnd.kafka.v2+json")))

	assert.False(t, isContentTypeJSON([]byte(fiber.MIMEApplicationXML)))
	assert.False(t, isContentTypeJSON([]byte(fiber.MIMETextPlain)))
}

func BenchmarkHandler(b *testing.B) {
	app := fiber.New()
	TracingMiddleware(&TracingConfig{Exporter: "stdout"}, app)
	LoggingMiddleware(app)
	app.Get("/test", func(c *fiber.Ctx) error {
		log.Ctx(c.UserContext()).Info().Msg("test")
		return nil
	})

	for i := 0; i < b.N; i++ {
		_, _ = app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	}
}

func Test_propagateTraceLogging(t *testing.T) {
	captor := strings.Builder{}
	zerolog.DefaultContextLogger = &log.Logger
	app := fiber.New()
	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	TracingMiddleware(&TracingConfig{Exporter: "stdout"}, app)
	LoggingMiddleware(app)
	log.Logger = log.Output(&captor)

	_, _ = app.Test(req)

	assert.Contains(t, captor.String(), "traceparent")
}
