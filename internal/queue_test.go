package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSafeCloseQ(t *testing.T) {
	q := NewSafeCloseQ(0, WithShutdownTimeout[string](1*time.Nanosecond))
	assert.Equal(t, 1*time.Nanosecond, q.shutdownTimeout)
}

func TestSafeness(t *testing.T) {

	q := NewSafeCloseQ(2, WithShutdownTimeout[string](1*time.Nanosecond), WithShutdownHandling(func(c chan string) {
		x := <-c
		assert.Equal(t, "b", x)
	}))
	assert.Nil(t, q.Enqueue("a"))
	assert.Nil(t, q.Enqueue("b"))

	x := q.Next()
	assert.Equal(t, "a", x)

	q.Close()

	assert.Error(t, q.Enqueue("xx"), "enqueue should return error after q has been closed")
}
