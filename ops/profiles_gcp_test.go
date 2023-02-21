package ops

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockResolver struct {
	resolveFunc func(ctx context.Context, host string) (addrs []string, err error)
}

func (m *mockResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	return m.resolveFunc(ctx, host)
}

func Test_testOnGCE(t *testing.T) {
	tests := []struct {
		name     string
		resolver resolver
		timeout  time.Duration
		want     bool
	}{
		{
			name: "success resolving",
			resolver: &mockResolver{
				func(_ context.Context, host string) (addrs []string, err error) {
					return []string{"1.2.3.4"}, nil
				},
			},
			timeout: 100 * time.Millisecond,
			want:    true,
		},
		{
			name: "error resolving",
			resolver: &mockResolver{
				func(_ context.Context, host string) (addrs []string, err error) {
					return nil, errors.New("boom")
				},
			},
			timeout: 100 * time.Millisecond,
			want:    false,
		},
		{
			name: "timeout resolving",
			resolver: &mockResolver{
				func(_ context.Context, host string) (addrs []string, err error) {
					time.Sleep(200 * time.Millisecond)
					return []string{"1.2.3.4"}, nil
				},
			},
			timeout: 100 * time.Millisecond,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, testOnGCE(tt.resolver, tt.timeout), "testOnGCE(%v)", tt.resolver)
		})
	}
}
