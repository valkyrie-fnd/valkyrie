package evolution

import (
	"context"
	"time"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type WalletService struct {
	pamClient pam.PamClient
	ctx       context.Context
}

func NewService(pamClient pam.PamClient) *WalletService {
	return &WalletService{pamClient: pamClient, ctx: context.Background()}
}

func (service *WalletService) WithContext(ctx context.Context) Service {
	return &WalletService{pamClient: service.pamClient, ctx: ctx}
}

// @Id           EvoCheck
// @Summary      Check
// @Description  Should be used for additional validation of redirected user and sid.
// @Tags         Evolution
// @Accept       json
// @Produce      json
// @Param        req               body      CheckRequest  true  "Request body"
// @Param        authToken         query     string        true  "Api token"
// @Success      200     {object}  CheckResponse
// @Failure      400     {object}  StandardResponse
// @Failure      500     {object}  StandardResponse
// @Router       /providers/evolution/check [post]
func (service *WalletService) Check(req CheckRequest) (*CheckResponse, error) {
	// Sanity check that req.UserID is valid and has an account balance.
	_, err := service.pamClient.GetBalance(service.balanceRequestMapper(req.RequestBase))
	if err != nil {
		return nil, toProviderError(err, req.UUID, ZeroAmount, ZeroAmount)
	}

	session, err := service.pamClient.RefreshSession(service.refreshSessionRequestMapper(req))
	if err != nil {
		return nil, toProviderError(err, req.UUID, ZeroAmount, ZeroAmount)
	}

	resp := CheckResponse{
		Status: "OK",
		SID:    session.Token,
		UUID:   req.UUID,
	}

	return &resp, nil
}

// @Id           EvoBalance
// @Summary      Balance
// @Description  Used to get user’s balance.
// @Tags         Evolution
// @Accept       json
// @Produce      json
// @Param        req               body      BalanceRequest  true  "Request body"
// @Param        authToken         query     string          true  "Api token"
// @Success      200     {object}  StandardResponse
// @Failure      400     {object}  StandardResponse
// @Failure      500     {object}  StandardResponse
// @Router       /providers/evolution/balance [post]
func (service *WalletService) Balance(req BalanceRequest) (*StandardResponse, error) {
	balance, err := service.pamClient.GetBalance(service.balanceRequestMapper(req.RequestBase))

	if err != nil {
		return nil, toProviderError(err, req.UUID, ZeroAmount, ZeroAmount)
	}

	return &StandardResponse{
		Status:  "OK",
		Balance: fromPamAmount(&balance.CashAmount),
		Bonus:   fromPamAmount(&balance.BonusAmount),
		UUID:    req.UUID,
	}, nil
}

// @Id           EvoDebit
// @Summary      Debit
// @Description  Used to debit from account (place bets).
// @Tags         Evolution
// @Accept       json
// @Produce      json
// @Param        req               body      DebitRequest    true  "Request body"
// @Param        authToken         query     string          true  "Api token"
// @Success      200     {object}  StandardResponse
// @Failure      400     {object}  StandardResponse
// @Failure      500     {object}  StandardResponse
// @Router       /providers/evolution/debit [post]
func (service *WalletService) Debit(req DebitRequest) (*StandardResponse, error) {
	// Send the debit transaction and ignore the success response
	transactionResp, err := service.pamClient.AddTransaction(service.debitRequestMapper(req, time.Now()))

	if err != nil {
		if transactionResp != nil && transactionResp.Balance != nil {
			return nil, toProviderError(err, req.UUID, fromPamAmount(&transactionResp.Balance.CashAmount), fromPamAmount(&transactionResp.Balance.BonusAmount))
		} else {
			return nil, toProviderError(err, req.UUID, ZeroAmount, ZeroAmount)
		}
	}

	var cashAmount, bonusAmount Amount

	if transactionResp != nil && transactionResp.Balance != nil {
		cashAmount = fromPamAmount(&transactionResp.Balance.CashAmount)
		bonusAmount = fromPamAmount(&transactionResp.Balance.BonusAmount)
	}

	return &StandardResponse{
		Status:  "OK",
		Balance: cashAmount,
		Bonus:   bonusAmount,
		UUID:    req.UUID,
	}, nil
}

// @Id           EvoCredit
// @Summary      Credit
// @Description  Used to credit user’s account (settle bets).
// @Tags         Evolution
// @Accept       json
// @Produce      json
// @Param        req               body      CreditRequest   true  "Request body"
// @Param        authToken         query     string          true  "Api token"
// @Success      200     {object}  StandardResponse
// @Failure      400     {object}  StandardResponse
// @Failure      500     {object}  StandardResponse
// @Router       /providers/evolution/credit [post]
func (service *WalletService) Credit(req CreditRequest) (*StandardResponse, error) {
	// Preflight check that the credit transaction is reasonable
	transactions, err := service.pamClient.GetTransactions(service.findTransForCreditRequestMapper(req))

	if err != nil {
		balance, balanceErr := service.pamClient.GetBalance(service.balanceRequestMapper(req.RequestBase))
		if balanceErr != nil {
			return nil, toProviderError(balanceErr, req.UUID, ZeroAmount, ZeroAmount)
		} else {
			return nil, toProviderError(err, req.UUID, fromPamAmount(&balance.CashAmount), fromPamAmount(&balance.BonusAmount))
		}
	}

	// Check that credit has a corresponding bet and that is not settled yet
	var validationError *statusCode
	if containsType(&transactions, pam.DEPOSIT, pam.CANCEL) {
		validationError = &StatusBetAlreadySettled
	} else if !containsType(&transactions, pam.WITHDRAW) {
		validationError = &StatusBetDoesNotExist
	}

	// fetch balance and bail if validation failed
	if validationError != nil {
		balance, balanceErr := service.pamClient.GetBalance(service.balanceRequestMapper(req.RequestBase))
		if balanceErr != nil {
			return nil, toProviderError(balanceErr, req.UUID, ZeroAmount, ZeroAmount)
		} else {
			return nil, createError("", *validationError, req.UUID, fromPamAmount(&balance.CashAmount), fromPamAmount(&balance.BonusAmount))
		}
	}

	// Send the credit transaction and ignore the success response
	transactionResp, err := service.pamClient.AddTransaction(service.creditTransRequestMapper(req, time.Now()))
	if err != nil {
		if transactionResp != nil && transactionResp.Balance != nil {
			return nil, toProviderError(err, req.UUID, fromPamAmount(&transactionResp.Balance.CashAmount), fromPamAmount(&transactionResp.Balance.BonusAmount))
		} else {
			return nil, toProviderError(err, req.UUID, ZeroAmount, ZeroAmount)
		}
	}

	var cashAmount, bonusAmount Amount

	if transactionResp.Balance != nil {
		cashAmount = fromPamAmount(&transactionResp.Balance.CashAmount)
		bonusAmount = fromPamAmount(&transactionResp.Balance.BonusAmount)
	}

	return &StandardResponse{
		Status:  "OK",
		Balance: cashAmount,
		Bonus:   bonusAmount,
		UUID:    req.UUID,
	}, err
}

// @Id           EvoCancel
// @Summary      Cancel
// @Description  Used to cancel user’s bet.
// @Tags         Evolution
// @Accept       json
// @Produce      json
// @Param        req               body      CancelRequest   true  "Request body"
// @Param        authToken         query     string          true  "Api token"
// @Success      200     {object}  StandardResponse
// @Failure      400     {object}  StandardResponse
// @Failure      500     {object}  StandardResponse
// @Router       /providers/evolution/cancel [post]
func (service *WalletService) Cancel(req CancelRequest) (*StandardResponse, error) {
	// Preflight check that the credit transaction is reasonable
	transactions, err := service.pamClient.GetTransactions(service.findTransForCancelRequestMapper(req))

	if err != nil {
		balance, balanceErr := service.pamClient.GetBalance(service.balanceRequestMapper(req.RequestBase))
		if balanceErr != nil {
			return nil, toProviderError(balanceErr, req.UUID, ZeroAmount, ZeroAmount)
		} else {
			return nil, toProviderError(err, req.UUID, fromPamAmount(&balance.CashAmount), fromPamAmount(&balance.BonusAmount))
		}
	}

	// Check that credit has a corresponding bet and that is not settled yet
	var validationError *statusCode
	if containsType(&transactions, pam.DEPOSIT, pam.CANCEL) {
		validationError = &StatusBetAlreadySettled
	} else if !containsType(&transactions, pam.WITHDRAW) {
		validationError = &StatusBetDoesNotExist
	}

	// fetch balance and bail if validation failed
	if validationError != nil {
		balance, balanceErr := service.pamClient.GetBalance(service.balanceRequestMapper(req.RequestBase))
		if balanceErr != nil {
			return nil, toProviderError(balanceErr, req.UUID, ZeroAmount, ZeroAmount)
		} else {
			return nil, createError("", *validationError, req.UUID, fromPamAmount(&balance.CashAmount), fromPamAmount(&balance.BonusAmount))
		}
	}

	// Send the debit transaction and ignore the success response
	transactionResp, err := service.pamClient.AddTransaction(service.cancelTransRequestMapper(req, time.Now()))
	if err != nil {
		if transactionResp != nil && transactionResp.Balance != nil {
			return nil, toProviderError(err, req.UUID, fromPamAmount(&transactionResp.Balance.CashAmount), fromPamAmount(&transactionResp.Balance.BonusAmount))
		} else {
			return nil, toProviderError(err, req.UUID, ZeroAmount, ZeroAmount)
		}
	}

	var cashAmount, bonusAmount Amount

	if transactionResp.Balance != nil {
		cashAmount = fromPamAmount(&transactionResp.Balance.CashAmount)
		bonusAmount = fromPamAmount(&transactionResp.Balance.BonusAmount)
	}
	return &StandardResponse{
		Status:  "OK",
		Balance: cashAmount,
		Bonus:   bonusAmount,
		UUID:    req.UUID,
	}, err

}

// @Id           EvoPromoPayout
// @Summary      PromoPayout
// @Description  Used to communicate promotional payout transactions.
// @Tags         Evolution
// @Accept       json
// @Produce      json
// @Param        req               body      PromoPayoutRequest  true  "Request body"
// @Param        authToken         query     string              true  "Api token"
// @Success      200     {object}  StandardResponse
// @Failure      400     {object}  StandardResponse
// @Failure      500     {object}  StandardResponse
// @Router       /providers/evolution/promo_payout [post]
func (service *WalletService) PromoPayout(req PromoPayoutRequest) (*StandardResponse, error) {
	// Send the debit transaction and ignore the success response
	transactionResp, err := service.pamClient.AddTransaction(service.promoPayoutTransRequestMapper(req, time.Now()))

	if err != nil {
		if transactionResp != nil && transactionResp.Balance != nil {
			return nil, toProviderError(err, req.UUID, fromPamAmount(&transactionResp.Balance.CashAmount), fromPamAmount(&transactionResp.Balance.BonusAmount))
		} else {
			return nil, toProviderError(err, req.UUID, ZeroAmount, ZeroAmount)
		}
	}

	var cashAmount, bonusAmount Amount

	if transactionResp.Balance != nil {
		cashAmount = fromPamAmount(&transactionResp.Balance.CashAmount)
		bonusAmount = fromPamAmount(&transactionResp.Balance.BonusAmount)
	}

	return &StandardResponse{
		Status:  "OK",
		Balance: cashAmount,
		Bonus:   bonusAmount,
		UUID:    req.UUID,
	}, err
}

func containsType(trx *[]pam.Transaction, tty ...pam.TransactionType) bool {
	for _, tr := range *trx {
		for _, ttype := range tty {
			if tr.TransactionType == ttype {
				return true
			}
		}
	}
	return false
}
