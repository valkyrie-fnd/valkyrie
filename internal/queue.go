package internal

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// SafeCloseQ wraps a channel and provides methods that can be used to
// close the channel without loosing elements.
type SafeCloseQ[T any] struct {
	c                 chan T
	onShutdownTimeout func(chan T)
	shutdownTimeout   time.Duration
	once              sync.Once
	stopping          atomic.Bool
}

func logWarningAndContents[T any](c chan T) {
	log.Warn().Msgf("shutting down queue with %d elements", len(c))
	for m := range c {
		log.Warn().Msgf("%v", m)
	}
}

type opt[T any] func(sfq *SafeCloseQ[T])

func WithShutdownTimeout[T any](timeout time.Duration) opt[T] {
	return func(sfq *SafeCloseQ[T]) {
		sfq.shutdownTimeout = timeout
	}
}

func WithShutdownHandling[T any](fn func(chan T)) opt[T] {
	return func(sfq *SafeCloseQ[T]) {
		sfq.onShutdownTimeout = fn
	}
}

// Creates a new SafeCloseQ with specified channel size and optional config
func NewSafeCloseQ[T any](buffer int, opts ...opt[T]) *SafeCloseQ[T] {
	d := SafeCloseQ[T]{
		c:                 make(chan T, buffer),
		stopping:          atomic.Bool{},
		onShutdownTimeout: logWarningAndContents[T],
		shutdownTimeout:   5 * time.Second,
	}

	for _, opt := range opts {
		opt(&d)
	}

	return &d
}

// Close will initiate closing by blocking new elements from being accepted in
// `Enqueue()` and starting a shutdown timer to allow consumers to complete.
func (sfq *SafeCloseQ[T]) Close() {
	sfq.once.Do(func() {
		sfq.stopping.Store(true)
	})
	defer close(sfq.c)
	pause := 100 * time.Millisecond
	attempts := int(sfq.shutdownTimeout / pause)
	for i := 0; i < attempts; i++ {
		if len(sfq.c) == 0 {
			return
		}
		time.Sleep(pause)
	}
	sfq.onShutdownTimeout(sfq.c)
}

// Enqueue offers a new element to the underlying channel iff
// shutdown is not in progress.
func (sfq *SafeCloseQ[T]) Enqueue(el T) error {
	if !sfq.stopping.Load() {
		sfq.c <- el
		return nil
	} else {
		// TODO: if useful make this behavior configurable
		return errors.New("q is shutting down")
	}
}

// Next is just a wrapper for channel read and will hang until
// a new message is available.
func (sfq *SafeCloseQ[T]) Next() T {
	return <-sfq.c
}
