package caleta

import (
	"errors"

	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

func getCErrorStatus(err error) Status {
	var vErr pam.ValkyrieError
	if errors.As(err, &vErr) {
		if status, ok := errCodes[vErr.ValkErrorCode]; ok {
			return status
		}
	}

	if errors.Is(err, rest.TimeoutError) {
		return RSERRORTIMEOUT
	}

	return RSERRORUNKNOWN
}

var errCodes = map[pam.ValkErrorCode]Status{
	pam.ValkErrAuth:              RSERRORINVALIDTOKEN,
	pam.ValkErrOpSessionNotFound: RSERRORINVALIDTOKEN,
	pam.ValkErrOpSessionExpired:  RSERRORTOKENEXPIRED,
	pam.ValkErrOpGameNotFound:    RSERRORINVALIDGAME,
	pam.ValkErrOpRoundNotFound:   RSERRORINVALIDGAME,
	pam.ValkErrWithdrawCurrency:  RSERRORWRONGCURRENCY,
	pam.ValkErrOpTransCurrency:   RSERRORWRONGCURRENCY,
	pam.ValkErrOpCashOverdraft:   RSERRORNOTENOUGHMONEY,
	pam.ValkErrOpBonusOverdraft:  RSERRORNOTENOUGHMONEY,
	pam.ValkErrOpPromoOverdraft:  RSERRORNOTENOUGHMONEY,
	pam.ValkErrOpBetNotAllowed:   RSERRORUSERDISABLED,
	pam.ValkErrOpTransNotFound:   RSERRORTRANSACTIONDOESNOTEXIST,
	pam.ValkErrOpCancelExists:    RSERRORTRANSACTIONROLLEDBACK,
	pam.ValkErrDuplicateTrans:    RSERRORDUPLICATETRANSACTION,
	pam.ValkErrOpCancelNotFound:  RSOK, // Caleta prefers that Valkyrie just returns OK in this case
	pam.ValkErrTimeout:           RSERRORTIMEOUT,
}

// errors left:
// RSERRORBETLIMITEXCEEDED
// RSERRORWRONGSYNTAX
// RSERRORWRONGTYPES
