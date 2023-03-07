package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Pipeline_Execute(t *testing.T) {
	pipeline := NewPipeline[int]()

	err := pipeline.Execute(context.TODO(), 123, func(pc PipelineContext[int]) error {
		assert.Equal(t, 123, pc.Payload())
		return nil
	})

	assert.NoError(t, err)
}

func Test_Pipeline_Error(t *testing.T) {
	pipeline := NewPipeline[int]()

	err := pipeline.Execute(context.TODO(), 123, func(pc PipelineContext[int]) error {
		return assert.AnError
	})

	assert.Error(t, err)
}

func Test_Pipeline_Registered_Handlers(t *testing.T) {
	pipeline := NewPipeline[int]()
	before := 0
	after := 0

	handler := func(pc PipelineContext[int]) error {

		before++
		err := pc.Next()
		after++

		return err
	}

	pipeline.Register(handler, handler)

	err := pipeline.Execute(context.TODO(), 1, func(pc PipelineContext[int]) error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, before)
	assert.Equal(t, 2, after)
}

func Test_Pipeline_Handler_Updating_Context(t *testing.T) {
	pipeline := NewPipeline[int]()

	handler := func(pc PipelineContext[int]) error {

		ctx := pc.Context()
		pc.SetContext(context.WithValue(ctx, "foo", "bar")) //nolint:staticcheck

		err := pc.Next()

		return err
	}

	pipeline.Register(handler)

	err := pipeline.Execute(context.TODO(), 1, func(pc PipelineContext[int]) error {
		assert.Equal(t, "bar", pc.Context().Value("foo"))
		return nil
	})

	assert.NoError(t, err)
}
