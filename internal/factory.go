package internal

import (
	"fmt"
	"sync"
)

func NewAbstractFactory[Args any, Target any]() *AbstractFactory[Args, Target] {
	return &AbstractFactory[Args, Target]{
		builders: map[string]func(Args) (Target, error){},
	}
}

// AbstractFactory contains a registry of available builders
type AbstractFactory[Args any, Target any] struct {
	builders map[string]func(Args) (Target, error)
	lock     sync.RWMutex
}

// Register a build function using a key
func (factory *AbstractFactory[Args, Target]) Register(key string, buildFn func(args Args) (Target, error)) {
	factory.lock.Lock()
	defer factory.lock.Unlock()

	factory.builders[key] = buildFn
}

// Build returns a built Target, or error
func (factory *AbstractFactory[Args, Target]) Build(key string, args Args) (Target, error) {
	factory.lock.RLock()
	defer factory.lock.RUnlock()

	buildFn, found := factory.builders[key]
	var zeroVal Target
	if !found {
		return zeroVal, fmt.Errorf("'%s' not found", key)
	}
	return buildFn(args)
}
