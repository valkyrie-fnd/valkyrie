package redtiger

import (
	"errors"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type RTErrorCode int

const (
	APIAuthError         RTErrorCode = 100
	InvalidInput         RTErrorCode = 200
	GenericError         RTErrorCode = 201
	NotAuthorized        RTErrorCode = 301
	UserNotFound         RTErrorCode = 302
	BannedUser           RTErrorCode = 303
	InsufficientFunds    RTErrorCode = 304
	InvalidUserCurrency  RTErrorCode = 305
	UserLimitedPlaying   RTErrorCode = 306
	TransactionNotFound  RTErrorCode = 400
	DuplicateTransaction RTErrorCode = 401
	InternalServerError  RTErrorCode = 500
	UnderMaintenanceMode RTErrorCode = 501
)

var errCodes = map[pam.ValkErrorCode]RTErrorCode{
	pam.ValkErrAPISession:          APIAuthError,
	pam.ValkErrReqInput:            InvalidInput,
	pam.ValkErrAuth:                NotAuthorized,
	pam.ValkErrOpSessionNotFound:   NotAuthorized,
	pam.ValkErrGetBalance:          InternalServerError,
	pam.ValkErrStakeValue:          InternalServerError,
	pam.ValkErrPromoValue:          InternalServerError,
	pam.ValkErrWithdraw:            InternalServerError,
	pam.ValkErrWithdrawCurrency:    InvalidUserCurrency,
	pam.ValkErrInterpretBalance:    InternalServerError,
	pam.ValkErrPayoutValue:         InternalServerError,
	pam.ValkErrPayoutPromoValue:    InternalServerError,
	pam.ValkErrPayoutNegativeStake: InvalidInput,
	pam.ValkErrPayoutZero:          InvalidInput,
	pam.ValkErrDeposit:             InternalServerError,
	pam.ValkErrRefundValue:         InternalServerError,
	pam.ValkErrRefundPromoValue:    InternalServerError,
	pam.ValkErrCancel:              InternalServerError,
	pam.ValkErrOpTransCurrency:     InvalidUserCurrency,
	pam.ValkErrOpUserNotFound:      NotAuthorized,
	pam.ValkErrOpAccountNotFound:   UserNotFound,
	pam.ValkErrOpBonusOverdraft:    InsufficientFunds,
	pam.ValkErrOpCashOverdraft:     InsufficientFunds,
	pam.ValkErrOpPromoOverdraft:    InsufficientFunds,
	pam.ValkErrOpCancelNotFound:    TransactionNotFound,
	pam.ValkErrOpTransNotFound:     TransactionNotFound,
	pam.ValkErrOpNegativeStake:     InvalidInput,
	pam.ValkErrOpRoundExists:       InvalidInput,
	pam.ValkErrOpCancelExists:      DuplicateTransaction,
	pam.ValkErrAlreadySettled:      DuplicateTransaction,
	pam.ValkErrOpCancelNonWithdraw: InvalidInput,
	pam.ValkErrOpBetNotAllowed:     BannedUser,
	pam.ValkErrUndefined:           GenericError,
}

func getError(valkError pam.ValkErrorCode) RTErrorCode {
	return errCodes[valkError]
}

func newRTErrorResponse(msg string, code RTErrorCode) ErrorResponse {
	// In case of auth errors limit details
	if code == NotAuthorized {
		msg = "Not authorized"
	}

	return ErrorResponse{
		Success: false,
		Error: Error{
			Message: msg,
			Code:    code,
		},
	}
}

func createRtErrorResponse(err error) ErrorResponse {
	code := extractValkErrorCode(err)
	return newRTErrorResponse(err.Error(), getError(code))
}

func extractValkErrorCode(err error) pam.ValkErrorCode {
	var valkErr pam.ValkyrieError
	if errors.As(err, &valkErr) {
		return valkErr.ValkErrorCode
	} else {
		return pam.GetErrUndefined()
	}
}
