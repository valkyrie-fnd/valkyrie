// Package routine provides functions for working with go routines
package routine

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var routines sync.WaitGroup

// Go runs the argument function as a separate go routine similar to 'go' syntax, but also
// tracks start and stop.
//
// This allows for gracefully waiting for routines to finish using WaitForFinish or WaitForFinishWithTimeout.
func Go(f func()) {
	routines.Add(1)
	go func() {
		defer routines.Done()
		f()
	}()
}

// WaitForFinishWithTimeout waits for routines to finish, or timeouts after specified duration.
func WaitForFinishWithTimeout(d time.Duration) {
	done := make(chan struct{})

	go func() {
		routines.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return
	case <-time.After(d):
		log.Error().Msg("Timeout waiting for routines to finish")
	}
}

// WaitForFinish indefinitely waits for routines to finish.
func WaitForFinish() {
	routines.Wait()
}
