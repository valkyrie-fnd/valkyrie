package pam

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// UnsupportedOperation Used in implementations when it doesn't support specific PamClient method
var UnsupportedOperation = errors.New("operation not supported by PAM")

// ValkErrorCode Error codes in valkyrie
type ValkErrorCode int

const (
	ValkErrUndefined ValkErrorCode = iota
	ValkErrAPISession
	ValkErrAuth
	ValkErrGetBalance
	ValkErrStakeValue
	ValkErrPromoValue
	ValkErrWithdraw
	ValkErrWithdrawCurrency
	ValkErrInterpretBalance
	ValkErrPayoutValue
	ValkErrPayoutPromoValue
	ValkErrDeposit
	ValkErrRefundValue
	ValkErrRefundPromoValue
	ValkErrCancel
	ValkErrPayoutNegativeStake
	ValkErrPayoutZero
	ValkErrOpUserNotFound
	ValkErrOpGameNotFound
	ValkErrOpTransNotFound
	ValkErrOpCashOverdraft
	ValkErrOpBonusOverdraft
	ValkErrOpSessionNotFound
	ValkErrOpSessionExpired
	ValkErrOpMissingProvider
	ValkErrOpTransCurrency
	ValkErrOpNegativeStake
	ValkErrOpZeroStake
	ValkErrOpRoundExists
	ValkErrOpCancelNotFound
	ValkErrOpCancelExists
	ValkErrOpCancelNonWithdraw
	ValkErrOpBetNotAllowed
	ValkErrAlreadySettled
	ValkErrBetNotFound
	ValkErrOpAccountNotFound
	ValkErrOpAPIToken
	ValkErrReqInput
	ValkErrOpRoundNotFound
	ValkErrDuplicateTrans
	ValkErrOpPromoOverdraft
	ValkErrTimeout
)

type ValkyrieError struct {
	OrigError error
	ErrMsg    string
	ValkErrorCode
}

func (e ValkyrieError) Error() string {
	if e.OrigError != nil {
		return fmt.Sprintf("Code: %d, Msg: %s, Orig error: %s", e.ValkErrorCode, e.ErrMsg, e.OrigError.Error())
	} else {
		return fmt.Sprintf("Code: %d, Msg: %s", e.ValkErrorCode, e.ErrMsg)
	}
}

func (e ValkyrieError) Unwrap() error {
	return e.OrigError
}

var pamToValkError = map[ErrorCode]ValkErrorCode{
	PAMERRUNDEFINED:             ValkErrUndefined,
	PAMERRBETNOTALLOWED:         ValkErrOpBetNotAllowed,
	PAMERRCANCELNONWITHDRAW:     ValkErrOpCancelNonWithdraw,
	PAMERRCANCELNOTFOUND:        ValkErrOpCancelNotFound,
	PAMERRBONUSOVERDRAFT:        ValkErrOpBonusOverdraft,
	PAMERRCASHOVERDRAFT:         ValkErrOpCashOverdraft,
	PAMERRPROMOOVERDRAFT:        ValkErrOpPromoOverdraft,
	PAMERRGAMENOTFOUND:          ValkErrOpGameNotFound,
	PAMERRMISSINGPROVIDER:       ValkErrOpMissingProvider,
	PAMERRNEGATIVESTAKE:         ValkErrOpNegativeStake,
	PAMERRPLAYERNOTFOUND:        ValkErrOpSessionNotFound,
	PAMERRSESSIONEXPIRED:        ValkErrOpSessionExpired,
	PAMERRSESSIONNOTFOUND:       ValkErrOpSessionNotFound,
	PAMERRTRANSALREADYCANCELLED: ValkErrOpCancelExists,
	PAMERRTRANSCURRENCY:         ValkErrOpTransCurrency,
	PAMERRTRANSNOTFOUND:         ValkErrOpTransNotFound,
	PAMERRACCNOTFOUND:           ValkErrOpAccountNotFound,
	PAMERRAPITOKEN:              ValkErrOpAPIToken,
	PAMERRROUNDNOTFOUND:         ValkErrOpRoundNotFound,
	PAMERRDUPLICATETRANS:        ValkErrDuplicateTrans,
	PAMERRTIMEOUT:               ValkErrTimeout,
}

// ToValkyrieError map PamError to ValkyrieError
func ToValkyrieError(err *PamError) error {
	code, found := pamToValkError[err.Code]
	if !found {
		code = ValkErrUndefined
		log.Warn().Msgf("Unable to map PamError code '%v' (%s) to ValkErrorCode, defaulting to '%v'",
			err.Code, err.Error(), code)
	}

	return ErrorWrapper(err.Error(), code, err)
}

func (e ValkyrieError) GetErrorCode(valkError ValkyrieError) int {
	return int(valkError.ValkErrorCode)
}

// GetErrUndefined Return ValkErrUndefined error code
func GetErrUndefined() ValkErrorCode {
	return ValkErrUndefined
}

// ErrorWrapper wrap err in a ValkyrieError if it not already is one
func ErrorWrapper(msg string, valkErrCode ValkErrorCode, err error) error {
	valkErr := ValkyrieError{
		ErrMsg:        msg,
		ValkErrorCode: valkErrCode,
		OrigError:     err,
	}
	return handleError(err, valkErr)
}

// If input error is a Valkyrie-error, just return it. Otherwise,
// return the assigned Valkyrie-error, or - if this is nil  - the
// original err
func handleError(err error, valkError ValkyrieError) error {
	if errors.As(err, &ValkyrieError{}) {
		return err
	}

	// Check for empty ValkyrieError struct input
	if valkError != (ValkyrieError{}) {
		return valkError
	} else {
		return err
	}
}

func (e PamError) Error() string {
	return fmt.Sprintf("%s %s", e.Code, e.Message)
}
