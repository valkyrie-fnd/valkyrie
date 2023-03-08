// Package pipeline provides functionality for implementing a chain of responsibility by passing a PipelineContext
// along one or more Handler functions, represented by a Pipeline.
//
// This allows for multiple Handler functions to handle a PipelineContext without coupling them with
// the sender code calling Pipeline.Execute(). The chain of Handler functions can be dynamically composed at
// runtime by registering them using the Pipeline.Register() function. Registered handlers in the Pipeline are run in
// the order that they are registered.
//
// This is especially useful when there is a need to introduce cross-cutting concerns to some component, such as
// caching or tracing, without coupling the component with its respective caching or tracing libraries.
//
// Here is a small example of how the pipeline package can be used in practice:
//
//	pipeline := NewPipeline[string]()
//	pipeline.Register(func(pc PipelineContext[string]) error {
//		fmt.Println("before", pc.Payload())
//		err := pc.Next()
//		fmt.Println("after", pc.Payload())
//		return err
//	})
//	pipeline.Execute(context.Background(), "foo", func(pc PipelineContext[string]) error {
//		fmt.Println("execute", pc.Payload())
//		return nil
//	})
//
// This will print the following output:
//
//	before foo
//	execute foo
//	after foo
//
// The registered logging Handler wraps the finalizing Handler supplied to Pipeline.Execute (printing "execute") and
// logs the PipelineContext.Payload() string both before and after the final Handler logging "execute" has run.
//
// It is important that intermediary Handler functions calls PipelineContext.Next() to continue the chain of handlers.
// Any potential error returned by PipelineContext.Next() should also be returned by its respective Handler to properly
// propagate errors back to the function calling Pipeline.Execute().
package pipeline

import (
	"context"
	"sync"
)

// Handler represents a step in a chain, together forming a "pipeline".
type Handler[T any] func(cc PipelineContext[T]) error

// pipelineContext keeps track of which handler in the pipeline to
// call next, as well as any associated context and payload.
type pipelineContext[T any] struct {
	ctx     context.Context
	payload T

	handlers []Handler[T]
	idx      int
}

// Next should be called by each handler to advance to the next handler
func (c *pipelineContext[T]) Next() error {

	c.idx++

	if c.idx <= len(c.handlers) {
		return c.handlers[c.idx-1](c)
	}

	return nil
}

func (c *pipelineContext[T]) Context() context.Context {
	return c.ctx
}

func (c *pipelineContext[T]) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *pipelineContext[T]) Payload() T {
	return c.payload
}

type PipelineContext[T any] interface {
	Next() error
	Context() context.Context
	SetContext(ctx context.Context)
	Payload() T
}

// Pipeline keeps track of all registered handlers that should be executed by the pipeline.
type Pipeline[T any] struct {
	handlers []Handler[T]
	lock     sync.RWMutex
}

func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{}
}

// Register one or more handlers
func (p *Pipeline[T]) Register(h Handler[T], handlers ...Handler[T]) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.handlers = append(p.handlers, h)
	if len(handlers) > 0 {
		p.handlers = append(p.handlers, handlers...)
	}
}

// Handlers retrieves a copy of registered handlers
func (p *Pipeline[T]) Handlers() []Handler[T] {
	p.lock.RLock()
	defer p.lock.RUnlock()

	copiedHandlers := make([]Handler[T], len(p.handlers))
	copy(copiedHandlers, p.handlers)

	return copiedHandlers
}

// Execute will run all handlers in the pipeline sequentially with the associated context and payload.
// A finalizer Handler is required and will be executed last to wrap up the pipeline.
func (p *Pipeline[T]) Execute(ctx context.Context, payload T, finalizer Handler[T]) error {

	// copy handlers from the pipeline
	handlers := p.Handlers()

	// add finalizing handler
	handlers = append(handlers, finalizer)

	pc := pipelineContext[T]{
		ctx:      ctx,
		payload:  payload,
		handlers: handlers,
	}

	// start the pipeline
	return pc.Next()
}
