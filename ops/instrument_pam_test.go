package ops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func Test_pamTracingHandler(t *testing.T) {
	handler := PAMTracingHandler("name")

	pc := &mockPipelineContext[any]{
		ctx:     context.TODO(),
		payload: pam.GetBalanceRequest{},
	}

	err := handler(pc)

	assert.NoError(t, err)
}

func Test_applyTracingFromContextHandler(t *testing.T) {
	handler := ApplyTracingFromContextHandler()
	tests := []struct {
		name    string
		payload any
		wantErr error
	}{
		{
			name:    "GetBalanceRequest",
			payload: &pam.GetBalanceRequest{},
		},
		{
			name:    "GetSessionRequest",
			payload: &pam.GetSessionRequest{},
		},
		{
			name:    "RefreshSessionRequest",
			payload: &pam.RefreshSessionRequest{},
		},
		{
			name:    "GetTransactionsRequest",
			payload: &pam.GetTransactionsRequest{},
		}, {
			name:    "AddTransactionRequest",
			payload: &pam.AddTransactionRequest{},
		},
		{
			name:    "GetGameRoundRequest",
			payload: &pam.GetGameRoundRequest{},
		},
		{
			name:    "error",
			payload: 1,
			wantErr: ErrorUnknownRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pc := &mockPipelineContext[any]{
				ctx:     context.TODO(),
				payload: test.payload,
			}

			err := handler(pc)

			if test.wantErr != nil {
				assert.Equal(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_pamMetricHandler(t *testing.T) {
	handler := PAMMetricHandler("name")

	pc := &mockPipelineContext[any]{
		ctx:     context.TODO(),
		payload: pam.GetBalanceRequest{},
	}

	err := handler(pc)

	assert.NoError(t, err)
}

func Test_getRequestName(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		want    string
	}{
		{
			name:    "GetBalanceRequest",
			payload: pam.GetBalanceRequest{},
			want:    "GetBalance",
		},
		{
			name:    "GetBalanceRequest pointer",
			payload: &pam.GetBalanceRequest{},
			want:    "GetBalance",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, getRequestName(test.payload))
		})
	}
}
