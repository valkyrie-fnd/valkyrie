package vplugin

import "github.com/valkyrie-fnd/valkyrie/pam"

func handleErrors[T any](pamError *pam.PamError, httpErr error, entity *T) error {
	if pamError != nil {
		// PamError has precedence since it contains more detailed error info from remote pam.
		return pam.ToValkyrieError(pamError)
	}
	if httpErr != nil {
		return pam.ErrorWrapper("http client error", pam.ValkErrUndefined, httpErr)
	}
	if entity == nil {
		return pam.ValkyrieError{ValkErrorCode: pam.ValkErrUndefined, ErrMsg: "nil entity"}
	}
	return nil
}
