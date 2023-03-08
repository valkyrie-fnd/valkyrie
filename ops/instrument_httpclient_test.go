package ops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type mockPipelineContext[T any] struct {
	ctx     context.Context
	payload T
}

func (m *mockPipelineContext[T]) Next() error {
	return nil
}

func (m *mockPipelineContext[T]) Context() context.Context {
	return m.ctx
}

func (m *mockPipelineContext[T]) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *mockPipelineContext[T]) Payload() T {
	return m.payload
}

type mockPayload struct {
	request  *fasthttp.Request
	response *fasthttp.Response
}

func (m mockPayload) Request() *fasthttp.Request {
	return m.request
}

func (m mockPayload) Response() *fasthttp.Response {
	return m.response
}

func Test_httpTracingHandler(t *testing.T) {
	handler := HTTPTracingHandler[FastHTTPPayload]()

	pc := &mockPipelineContext[FastHTTPPayload]{
		ctx: context.TODO(),
		payload: mockPayload{
			request:  fasthttp.AcquireRequest(),
			response: fasthttp.AcquireResponse(),
		},
	}

	err := handler(pc)

	assert.NoError(t, err)
}

func Test_httpLoggingHandler(t *testing.T) {
	handler := HTTPLoggingHandler[FastHTTPPayload]()

	pc := &mockPipelineContext[FastHTTPPayload]{
		ctx: context.TODO(),
		payload: mockPayload{
			request:  fasthttp.AcquireRequest(),
			response: fasthttp.AcquireResponse(),
		},
	}

	err := handler(pc)

	assert.NoError(t, err)
}

func Test_httpMetricHandler(t *testing.T) {
	handler := HTTPMetricHandler[FastHTTPPayload]()

	pc := &mockPipelineContext[FastHTTPPayload]{
		ctx: context.TODO(),
		payload: mockPayload{
			request:  fasthttp.AcquireRequest(),
			response: fasthttp.AcquireResponse(),
		},
	}

	err := handler(pc)

	assert.NoError(t, err)
}
