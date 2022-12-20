package caleta

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type ProviderService struct {
	pam.PamClient
}

// NewService Create new caleta provider service
func NewService(pamClient pam.PamClient) *ProviderService {
	return &ProviderService{PamClient: pamClient}
}

// @Id           CaletaBalance
// @Summary      Balance
// @Description  Wallet balance
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletbalanceJSONRequestBody   true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  WalletbalanceResponseObject
// @Failure      400     {object}  WalletbalanceResponseObject
// @Router       /providers/caleta/wallet/balance [post]
func (s *ProviderService) Walletbalance(ctx context.Context, request WalletbalanceRequestObject) (WalletbalanceResponseObject, error) {
	session, err := s.GetSession(s.getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletbalance200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	balance, err := s.GetBalance(s.balanceRequestMapper(ctx, *request.Body))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletbalance200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	amt := fromPamAmount(balance.CashAmount)
	return Walletbalance200JSONResponse{
		Balance:     amt,
		Currency:    (*Currency)(&session.Currency),
		RequestUuid: request.Body.RequestUuid,
		Status:      RSOK,
		User:        &request.Body.SupplierUser,
	}, nil
}

// @Id           CaletaCheck
// @Summary      Check
// @Description  Wallet check
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletcheckJSONRequestBody     true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  WalletcheckResponseObject
// @Failure      400     {object}  WalletcheckResponseObject
// @Router       /providers/caleta/wallet/check [post]
func (s *ProviderService) Walletcheck(ctx context.Context, request WalletcheckRequestObject) (WalletcheckResponseObject, error) {
	session, err := s.RefreshSession(s.refreshSessionMapper(ctx, request.Body.Token, ""))
	if err != nil {
		// Walletcheck has no error status in response. Return error instead
		return nil, err
	}
	return Walletcheck200JSONResponse{
		Token: &session.Token,
	}, nil
}

// @Id           CaletaBet
// @Summary      Bet
// @Description  Wallet bet
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletbetJSONRequestBody       true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  WalletbetResponseObject
// @Failure      400     {object}  WalletbetResponseObject
// @Router       /providers/caleta/wallet/bet [post]
func (s *ProviderService) Walletbet(ctx context.Context, request WalletbetRequestObject) (WalletbetResponseObject, error) {
	session, err := s.GetSession(s.getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletbet200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	var tranRes *pam.TransactionResult
	if request.Body.IsFree {
		tranRes, err = s.AddTransaction(s.promoBetTransactionMapper(ctx, *request.Body))
	} else {
		tranRes, err = s.AddTransaction(s.betTransactionMapper(ctx, *request.Body))
	}
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletbet200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	amt := fromPamAmount(tranRes.Balance.CashAmount)
	return Walletbet200JSONResponse{
		Status:      RSOK,
		Balance:     amt,
		Currency:    (*Currency)(&session.Currency),
		RequestUuid: request.Body.RequestUuid,
		User:        &request.Body.SupplierUser,
	}, nil
}

// @Id           CaletaWin
// @Summary      Win
// @Description  Wallet win
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      TransactionwinJSONRequestBody  true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  TransactionwinResponseObject
// @Failure      400     {object}  TransactionwinResponseObject
// @Router       /providers/caleta/wallet/win [post]
func (s *ProviderService) Transactionwin(ctx context.Context, request TransactionwinRequestObject) (TransactionwinResponseObject, error) {
	session, err := s.GetSession(s.getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Transactionwin200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	var tranRes *pam.TransactionResult
	var requestMapper pam.AddTransactionRequestMapper
	if request.Body.IsFree {
		// Check that Bet-transaction Exist. For Promo-transactions this check needs to be done here
		transactions, tErr := s.GetTransactions(s.getTransactionsMapper(ctx, *request.Body))
		if tErr != nil {
			return Transactionwin200JSONResponse{Status: RSERRORTRANSACTIONDOESNOTEXIST, RequestUuid: request.Body.RequestUuid}, nil
		}

		for _, t := range transactions {
			if t.TransactionType == pam.PROMOCANCEL {
				// PromoBet transaction has been cancelled, reject win with expected error
				return Transactionwin200JSONResponse{Status: RSERRORTRANSACTIONROLLEDBACK, RequestUuid: request.Body.RequestUuid}, nil
			}
		}

		requestMapper = s.promoWinTransactionMapper(ctx, *request.Body)
	} else {
		requestMapper = s.winTransactionMapper(ctx, *request.Body)
	}
	tranRes, err = s.AddTransaction(requestMapper)
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Transactionwin200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	amt := fromPamAmount(tranRes.Balance.CashAmount)
	return Transactionwin200JSONResponse{
		Balance:     amt,
		Status:      RSOK,
		User:        &request.Body.SupplierUser,
		RequestUuid: request.Body.RequestUuid,
		Currency:    (*Currency)(&session.Currency),
	}, nil
}

// @Id           CaletaRollback
// @Summary      Rollback
// @Description  Wallet rollback
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletrollbackJSONRequestBody  true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  WalletrollbackResponseObject
// @Failure      400     {object}  WalletrollbackResponseObject
// @Router       /providers/caleta/wallet/rollback [post]
func (s *ProviderService) Walletrollback(ctx context.Context, request WalletrollbackRequestObject) (WalletrollbackResponseObject, error) {
	session, err := s.GetSession(s.getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletrollback200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	var tranRes *pam.TransactionResult
	if request.Body.IsFree != nil && *request.Body.IsFree {
		tranRes, err = s.AddTransaction(s.cancelTransactionMapper(ctx, *request.Body, session, pam.PROMOCANCEL))
	} else {
		tranRes, err = s.AddTransaction(s.cancelTransactionMapper(ctx, *request.Body, session, pam.CANCEL))
	}
	if err != nil {
		errStatus := getCErrorStatus(err)
		if errStatus != RSOK {
			// All rollback attempts should return OK and balance,
			// otherwise we risk getting stuck in infinite rollback loop from Caleta.
			log.Ctx(ctx).
				Err(err).
				Interface("transaction", *request.Body).
				Msg("Rollback failed")
		}
	}

	amt := 0
	if tranRes != nil && tranRes.Balance != nil {
		amt = *fromPamAmount(tranRes.Balance.CashAmount)
	}

	return Walletrollback200JSONResponse{
		Status:      RSOK,
		Balance:     &amt,
		Currency:    (*Currency)(&session.Currency),
		RequestUuid: request.Body.RequestUuid,
		User:        request.Body.User,
	}, nil
}
