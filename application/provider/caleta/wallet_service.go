package caleta

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type WalletService struct {
	PamClient pam.PamClient
	APIClient API
}

// NewService Create new caleta wallet service
func NewWalletService(pamClient pam.PamClient, apiClient API) *WalletService {
	return &WalletService{PamClient: pamClient, APIClient: apiClient}
}

// @Id           CaletaBalance
// @Summary      Balance
// @Description  Should return wallet balance for current player.
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletbalanceJSONRequestBody   true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  Walletbalance200JSONResponse
// @Failure      400     {object}  Walletbalance200JSONResponse
// @Router       /providers/caleta/wallet/balance [post]
func (s *WalletService) Walletbalance(ctx context.Context, request WalletbalanceRequestObject) (WalletbalanceResponseObject, error) {
	session, err := s.PamClient.GetSession(getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletbalance200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}
	balance, err := s.PamClient.GetBalance(balanceRequestMapper(ctx, request.Body))
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
// @Description  OPTIONAL - Change the initial token received for a new one that will be used on wallet transactions.
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletcheckJSONRequestBody     true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  Walletcheck200JSONResponse
// @Failure      400     {object}  Walletcheck200JSONResponse
// @Router       /providers/caleta/wallet/check [post]
func (s *WalletService) Walletcheck(ctx context.Context, request WalletcheckRequestObject) (WalletcheckResponseObject, error) {
	session, err := s.PamClient.RefreshSession(refreshSessionMapper(ctx, request.Body.Token, ""))
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
// @Description  Called when the user places a bet (debit).
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletbetJSONRequestBody       true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  Walletbet200JSONResponse
// @Failure      400     {object}  Walletbet200JSONResponse
// @Router       /providers/caleta/wallet/bet [post]
func (s *WalletService) Walletbet(ctx context.Context, request WalletbetRequestObject) (WalletbetResponseObject, error) {
	session, err := s.PamClient.GetSession(getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletbet200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}

	var tranRes *pam.TransactionResult
	if request.Body.IsFree {
		tranRes, err = s.PamClient.AddTransaction(promoBetTransactionMapper(ctx, &request))
	} else {
		tranRes, err = s.PamClient.AddTransaction(betTransactionMapper(ctx, &request))
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
// @Description  Called when the user wins (credit).
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      TransactionwinJSONRequestBody  true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  Transactionwin200JSONResponse
// @Failure      400     {object}  Transactionwin200JSONResponse
// @Router       /providers/caleta/wallet/win [post]
func (s *WalletService) Transactionwin(ctx context.Context, request TransactionwinRequestObject) (TransactionwinResponseObject, error) {
	session, err := s.PamClient.GetSession(getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Transactionwin200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}

	// If gamewise settlement, get the round transactions
	roundTransactions := s.getRoundTransactions(ctx, s.PamClient.GetSettlementType(), request.Body.Round)

	var tranRes *pam.TransactionResult
	var requestMapper pam.AddTransactionRequestMapper
	if request.Body.IsFree {
		// Check that Bet-transaction Exist. For Promo-transactions this check needs to be done here
		transactions, tErr := s.PamClient.GetTransactions(getTransactionsMapper(ctx, request.Body))
		if tErr != nil {
			return Transactionwin200JSONResponse{Status: RSERRORTRANSACTIONDOESNOTEXIST, RequestUuid: request.Body.RequestUuid}, nil
		}

		for _, t := range transactions {
			if t.TransactionType == pam.PROMOCANCEL {
				// PromoBet transaction has been cancelled, reject win with expected error
				return Transactionwin200JSONResponse{Status: RSERRORTRANSACTIONROLLEDBACK, RequestUuid: request.Body.RequestUuid}, nil
			}
		}

		requestMapper = promoWinTransactionMapper(ctx, &request, roundTransactions)
	} else {
		requestMapper = winTransactionMapper(ctx, &request, roundTransactions)
	}
	tranRes, err = s.PamClient.AddTransaction(requestMapper)
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
// @Description  Called when there is need to roll back the effect of the referenced transaction.
// @Tags         Caleta
// @Accept       json
// @Produce      json
// @Param        req               body      WalletrollbackJSONRequestBody  true  "Request body"
// @Param        X-Auth-Signature  header    string                         true  "Signature for request"
// @Success      200     {object}  Walletrollback200JSONResponse
// @Failure      400     {object}  Walletrollback200JSONResponse
// @Router       /providers/caleta/wallet/rollback [post]
func (s *WalletService) Walletrollback(ctx context.Context, request WalletrollbackRequestObject) (WalletrollbackResponseObject, error) {
	session, err := s.PamClient.GetSession(getSessionMapper(ctx, request.Body.Token, request.Body.RequestUuid))
	if err != nil {
		errStatus := getCErrorStatus(err)
		return Walletrollback200JSONResponse{Status: errStatus, RequestUuid: request.Body.RequestUuid}, nil
	}

	// If gamewise settlement, get the round transactions
	roundTransactions := s.getRoundTransactions(ctx, s.PamClient.GetSettlementType(), request.Body.Round)

	var tranRes *pam.TransactionResult
	if request.Body.IsFree != nil && *request.Body.IsFree {
		tranRes, err = s.PamClient.AddTransaction(cancelTransactionMapper(ctx, &request, session, pam.PROMOCANCEL, roundTransactions))
	} else {
		tranRes, err = s.PamClient.AddTransaction(cancelTransactionMapper(ctx, &request, session, pam.CANCEL, roundTransactions))
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

// getRoundTransactions fetches all transactions linked to a given round (only if PAM settlement type is "GAMEWISE")
func (s *WalletService) getRoundTransactions(ctx context.Context, settlementType pam.SettlementType, round string) *[]roundTransaction {
	var roundTransactions *[]roundTransaction
	if settlementType == pam.GAMEWISE {
		if rounds, err := s.APIClient.getRoundTransactions(ctx, round); err != nil {
			log.Warn().Msg(fmt.Sprintf("Failed to get round transactions for ID %s, reason: %s", round, err.Error()))
		} else {
			roundTransactions = rounds.RoundTransactions
		}
	}

	return roundTransactions
}
