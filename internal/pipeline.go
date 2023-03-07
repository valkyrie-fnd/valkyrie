package internal

import (
	"context"
	"sync"
)

type Handler[T any] func(cc PipelineContext[T]) error

type pipelineContext[T any] struct {
	ctx     context.Context
	payload T

	handlers []Handler[T]
	idx      int
}

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

type Pipeline[T any] struct {
	handlers []Handler[T]
	lock     sync.RWMutex
}

func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{}
}

func (p *Pipeline[T]) Register(h Handler[T], handlers ...Handler[T]) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.handlers = append(p.handlers, h)
	if len(handlers) > 0 {
		p.handlers = append(p.handlers, handlers...)
	}
}

func (p *Pipeline[T]) Handlers() []Handler[T] {
	p.lock.RLock()
	defer p.lock.RUnlock()

	copiedHandlers := make([]Handler[T], len(p.handlers))
	copy(copiedHandlers, p.handlers)

	return copiedHandlers
}

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
