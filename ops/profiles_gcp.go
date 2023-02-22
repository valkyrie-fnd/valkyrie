package ops

import (
	"context"
	"net"
	"sync"
	"time"
)

var (
	onGCEOnce sync.Once
	onGCE     bool
)

func init() {
	OnGCE()
}

// OnGCE reports whether this process is running on Google Compute Engine.
func OnGCE() bool {
	onGCEOnce.Do(initOnGCE)
	return onGCE
}

func initOnGCE() {
	onGCE = testOnGCE(&net.Resolver{}, 3*time.Second)
}

type resolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
}

func testOnGCE(resolver resolver, timeout time.Duration) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resChan := make(chan bool, 1)

	go func() {
		addrs, err := resolver.LookupHost(ctx, "metadata.google.internal.")
		if err != nil || len(addrs) == 0 {
			resChan <- false
			return
		}
		resChan <- true
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var res bool
	select {
	case res = <-resChan:
		return res
	case <-timer.C: // timeout
		return false
	}
}
