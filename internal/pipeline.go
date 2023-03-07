package internal

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
