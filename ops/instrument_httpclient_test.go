package ops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/valkyrie-fnd/valkyrie/rest"
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

func Test_httpTracingHandler(t *testing.T) {
	handler := httpTracingHandler()

	pc := &mockPipelineContext[rest.PipelinePayload]{
		ctx: context.TODO(),
		payload: rest.PipelinePayload{
			Request:  fasthttp.AcquireRequest(),
			Response: fasthttp.AcquireResponse(),
		},
	}

	err := handler(pc)

	assert.NoError(t, err)
}

func Test_httpLoggingHandler(t *testing.T) {
	handler := httpLoggingHandler()

	pc := &mockPipelineContext[rest.PipelinePayload]{
		ctx: context.TODO(),
		payload: rest.PipelinePayload{
			Request:  fasthttp.AcquireRequest(),
			Response: fasthttp.AcquireResponse(),
		},
	}

	err := handler(pc)

	assert.NoError(t, err)
}

func Test_httpMetricHandler(t *testing.T) {
	handler := httpMetricHandler()

	pc := &mockPipelineContext[rest.PipelinePayload]{
		ctx: context.TODO(),
		payload: rest.PipelinePayload{
			Request:  fasthttp.AcquireRequest(),
			Response: fasthttp.AcquireResponse(),
		},
	}

	err := handler(pc)

	assert.NoError(t, err)
}
