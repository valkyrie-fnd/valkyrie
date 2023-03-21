package caleta

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/valkhttp"
)

func Test_getCErrorStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Status
	}{
		{
			"generic error",
			errors.New("boom"),
			RSERRORUNKNOWN,
		},
		{
			"valkyrie timeout error",
			pam.ValkyrieError{
				ValkErrorCode: pam.ValkErrTimeout,
			},
			RSERRORTIMEOUT,
		},
		{
			"valkyrie session expired error",
			pam.ValkyrieError{
				ValkErrorCode: pam.ValkErrOpSessionExpired,
			},
			RSERRORTOKENEXPIRED,
		},
		{
			"http timeout error",
			valkhttp.TimeoutError,
			RSERRORTIMEOUT,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getCErrorStatus(tt.err), "getCErrorStatus(%v)", tt.err)
		})
	}
}
