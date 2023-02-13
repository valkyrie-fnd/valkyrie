package evolution

import (
	"errors"
	"net/http"

	httpclient "github.com/valkyrie-fnd/valkyrie/httpclient"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type ProviderError struct {
	response   *StandardResponse
	message    string
	httpStatus int
}

func (e ProviderError) Error() string {
	return e.message
}

type statusCode struct {
	code     string
	httpCode int
}

var (
	StatusOK                 = statusCode{"OK", http.StatusOK}                         // Success
	StatusTemporaryError     = statusCode{"TEMPORARY_ERROR", http.StatusOK}            // There is a temporary problem with the game server.
	StatusUnknownError       = statusCode{"UNKNOWN_ERROR", http.StatusOK}              // Please contact Customer Support for assistance.
	StatusInvalidSID         = statusCode{"INVALID_SID", http.StatusOK}                // There has been a problem with the Live Casino. User authentication failed or your session may be expired, please close the browser and try again. Error Code: EV01
	StatusInsufficientFunds  = statusCode{"INSUFFICIENT_FUNDS", http.StatusOK}         // You do not have sufficient funds to place this bet.
	StatusInvalidTokenID     = statusCode{"INVALID_TOKEN_ID", http.StatusUnauthorized} // There has been a problem with the Live Casino. User authentication failed or your session may be expired, please close the browser and try again. Error Code: EV01
	StatusInvalidParameter   = statusCode{"INVALID_PARAMETER", http.StatusOK}          // Please contact Customer Support for assistance.
	StatusBetAlreadySettled  = statusCode{"BET_ALREADY_SETTLED", http.StatusOK}        // Success 	Bet has been already settled in third party system.
	StatusBetDoesNotExist    = statusCode{"BET_DOES_NOT_EXIST", http.StatusOK}         // Please contact Customer Support for assistance.
	StatusOperationInProcess = statusCode{"OPERATION_IN_PROCESS", 200}                 // Retryable Error 	Transaction is being process and no result is currently available.
)

var retryableErrors = map[statusCode]bool{
	StatusTemporaryError:     true,
	StatusOperationInProcess: true,
}

var errCodes = map[pam.ValkErrorCode]statusCode{
	pam.ValkErrUndefined:         StatusUnknownError,
	pam.ValkErrOpSessionExpired:  StatusInvalidSID,
	pam.ValkErrOpSessionNotFound: StatusInvalidSID,
	pam.ValkErrOpCashOverdraft:   StatusInsufficientFunds,
	pam.ValkErrAuth:              StatusInvalidTokenID,
	pam.ValkErrOpUserNotFound:    StatusInvalidSID,
	pam.ValkErrAlreadySettled:    StatusBetAlreadySettled,
	pam.ValkErrBetNotFound:       StatusBetDoesNotExist,
	pam.ValkErrOpCancelNotFound:  StatusBetDoesNotExist,
	pam.ValkErrOpTransNotFound:   StatusBetDoesNotExist,
}

var httpErrCodes = map[int]statusCode{
	http.StatusUnauthorized: StatusInvalidSID,
	http.StatusBadRequest:   StatusInvalidParameter,
}

func toProviderError(err error, uuid string, balance, bonus Amount) ProviderError {
	status := StatusUnknownError
	valkErr := &pam.ValkyrieError{}
	if errors.As(err, valkErr) {
		if st, found := errCodes[valkErr.ValkErrorCode]; found {
			status = st
		}
	}

	httpErr := &httpclient.HTTPError{}
	if errors.As(err, httpErr) {
		if st, found := httpErrCodes[httpErr.Code]; found {
			status = st
		}
	}

	return createError(err.Error(), status, uuid, balance, bonus)
}

func createError(msg string, status statusCode, uuid string, balance, bonus Amount) ProviderError {
	return ProviderError{
		httpStatus: status.httpCode,
		message:    msg,
		response: &StandardResponse{
			Status:         status.code,
			UUID:           uuid,
			Balance:        balance,
			Bonus:          bonus,
			Retransmission: retryableErrors[status],
		},
	}
}

func defaultErrorResponse(status, uuID string) *StandardResponse {
	return &StandardResponse{
		Status:  status,
		Balance: ZeroAmount,
		UUID:    uuID,
	}
}

func unwrap(err error) *ProviderError {
	var e ProviderError

	if errors.As(err, &e) {
		return &e
	} else {
		return &ProviderError{httpStatus: 500, message: "UNKNOWN_ERROR"}
	}
}
