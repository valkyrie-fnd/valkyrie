package pam

import (
	"errors"
	"fmt"
	"testing"
)

func Test_handleError(t *testing.T) {
	type args struct {
		err       error
		valkError ValkyrieError
	}
	testError := errors.New("error")
	valkyrieError := ValkyrieError{ErrMsg: "err", ValkErrorCode: ValkErrUndefined}
	emptyValkyrieError := ValkyrieError{}
	wrappedValkyrieError := fmt.Errorf("wrapped: %w", valkyrieError)
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"error does not wrap ValkyrieError", args{testError, valkyrieError}, valkyrieError},
		{"error is used by default when ValkyrieError is empty", args{testError, emptyValkyrieError}, testError},
		{"error wraps ValkyrieError", args{wrappedValkyrieError, emptyValkyrieError}, wrappedValkyrieError},
		{"error wraps ValkyrieError is used even when ValkyrieError is provided", args{wrappedValkyrieError, valkyrieError}, wrappedValkyrieError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleError(tt.args.err, tt.args.valkError); !errors.Is(err, tt.wantErr) {
				t.Errorf("handleError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToValkyrieError(t *testing.T) {
	pamError := &PamError{PAMERRUNDEFINED, "message"}
	valkyrieError := ValkyrieError{ErrMsg: "PAM_ERR_UNDEFINED message", ValkErrorCode: GetErrUndefined(), OrigError: pamError}
	pamErrorNegativeStake := &PamError{PAMERRNEGATIVESTAKE, "neg"}
	valkyrieErrorNegativeStake := ValkyrieError{ErrMsg: "PAM_ERR_NEGATIVE_STAKE neg", ValkErrorCode: ValkErrOpNegativeStake, OrigError: pamErrorNegativeStake}
	tests := []struct {
		name     string
		pamError *PamError
		wantErr  error
	}{
		{"pam err undefined", pamError, valkyrieError},
		{"pam err negative stake", pamErrorNegativeStake, valkyrieErrorNegativeStake},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ToValkyrieError(tt.pamError); !errors.Is(err, tt.wantErr) {
				t.Errorf("ToValkyrieError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
