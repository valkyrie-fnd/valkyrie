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

func TestHTTPEventContentLengthNegative(t *testing.T) {
	captor := strings.Builder{}
	logger := zerolog.New(&captor)
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	response.SetBodyString("foo")
	response.Header.SetContentLength(-2)
	request.SetBodyString("bar")
	request.Header.SetContentLength(-2)

	logger.Debug().Func(LogHTTPEvent(request, response, nil)).Send()

	assert.Equal(t, "{\"level\":\"debug\",\"httpRequest\":{\"requestUrl\":\"http:///\",\"requestMethod\":\"GET\",\"status\":200,\"protocol\":\"HTTP/1.1\",\"requestHeaders\":{},\"requestSize\":-2,\"responseHeaders\":{\"Content-Type\":\"text/plain; charset=utf-8\"},\"responseSize\":-2}}\n", captor.String())
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
