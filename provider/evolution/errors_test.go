package evolution

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func Test_toProviderError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ProviderError
	}{
		{
			"Valk error should map to proper provider error",
			pam.ValkyrieError{
				ValkErrorCode: pam.ValkErrOpCashOverdraft,
				ErrMsg:        "ignore",
			},
			ProviderError{
				httpStatus: StatusInsufficientFunds.httpCode,
				message:    "ignore",
				response: &StandardResponse{
					Status:  StatusInsufficientFunds.code,
					Balance: amountFromFloat(1),
					Bonus:   amountFromFloat(2),
					UUID:    "any",
				},
			},
		},
		{
			"Raw error gets mapped to unknown",
			errors.New("yikes"),
			ProviderError{
				httpStatus: StatusUnknownError.httpCode,
				message:    "ignore",
				response: &StandardResponse{
					Status:  StatusUnknownError.code,
					Balance: amountFromFloat(1),
					Bonus:   amountFromFloat(2),
					UUID:    "any",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := toProviderError(tt.err, "any", amountFromFloat(1), amountFromFloat(2))
			// we don't care about the exact message here
			res.message = tt.want.message
			assert.Equal(t, tt.want, res)
		})
	}
}
