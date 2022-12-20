package caleta

import (
	"errors"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func getCErrorStatus(err error) Status {
	var valkErr pam.ValkyrieError
	if errors.As(err, &valkErr) {
		if status, ok := errCodes[valkErr.ValkErrorCode]; ok {
			return status
		}
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
// RSERRORTIMEOUT
// RSERRORWRONGSYNTAX
// RSERRORWRONGTYPES
