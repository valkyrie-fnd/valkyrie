package redtiger

import (
	"context"
	"errors"
	"fmt"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type WalletService struct {
	pamClient pam.PamClient
	ctx       context.Context
}

// NewService Create new red tiger provider service
func NewService(pamClient pam.PamClient) *WalletService {
	return &WalletService{pamClient: pamClient, ctx: context.Background()}
}

func (s *WalletService) WithContext(ctx context.Context) Service {
	return &WalletService{pamClient: s.pamClient, ctx: ctx}
}

// Auth implements Service
// @Id           RTAuth
// @Summary      Auth
// @Description  Authenticate toward Red tiger.
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body         AuthRequest  true  "Request body"
// @Param        Authorization     header       string       true  "API Key"
// @Success      200     {object}  AuthResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/auth [post]
func (s *WalletService) Auth(req AuthRequest) (*AuthResponseWrapper, *ErrorResponse) {
	session, err := s.pamClient.RefreshSession(s.refreshSessionRequestMapper(req))
	if err != nil {
		e := createRtErrorResponse(fmt.Errorf("Failed to Auth: %w", err))
		return nil, &e
	}

	// Use the new token returned by refresh session for subsequent calls
	req.Token = session.Token
	if req.UserID == "" {
		req.UserID = session.PlayerId
	}

	balance, err := s.pamClient.GetBalance(s.getBalanceMapper(req.BaseRequest))
	if err != nil {
		e := createRtErrorResponse(fmt.Errorf("Failed to Auth: %w", err))
		return nil, &e
	}

	return &AuthResponseWrapper{
		Success: true,
		Result: AuthResponse{
			BaseResponse: BaseResponse{
				Token:    session.Token,
				Currency: session.Currency,
			},
			UserID:   req.UserID,
			Country:  session.Country,
			Language: session.Language,
			Casino:   req.Casino,
			Balance: Balance{
				Cash:  Money(balance.CashAmount),
				Bonus: Money(balance.BonusAmount),
			},
		},
	}, nil
}

// Payout implements Service
// @Id           RTPayout
// @Summary      Payout
// @Description  When a bet settles with a payout (credit).
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body         PayoutRequest  true  "Request body"
// @Param        Authorization     header       string         true  "API Key"
// @Success      200     {object}  PayoutResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/payout [post]
func (s *WalletService) Payout(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
	return handlePayout(s, req, pam.DEPOSIT)

}

// Stake implements Service
// @Id           RTStake
// @Summary      Stake
// @Description  When a bet has been placed (debit).
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       StakeRequest  true  "Request body"
// @Param        Authorization     header     string        true  "API Key"
// @Success      200     {object}  StakeResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/stake [post]
func (s *WalletService) Stake(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
	return handleStake(s, req, pam.WITHDRAW)
}

// Refund implements Service
// @Id           RTRefund
// @Summary      Refund
// @Description  Used to refund a placed bet.
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       RefundRequest  true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  RefundResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/refund [post]
func (s *WalletService) Refund(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
	return handleRefund(s, req, pam.CANCEL)
}

// PromoBuyin implements Service
// @Id           RTPromoBuyin
// @Summary      PromoBuyin
// @Description  Promotion buyin, request the same as stake/bet.
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       StakeRequest   true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  StakeResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/promo/buyin [post]
func (s *WalletService) PromoBuyin(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
	return handleStake(s, req, pam.PROMOWITHDRAW)
}

// PromoRefund implements Service
// @Id           RTPromoRefund
// @Summary      PromoRefund
// @Description  Refund promotion buyin.
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       RefundRequest  true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  RefundResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/promo/refund [post]
func (s *WalletService) PromoRefund(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
	return handleRefund(s, req, pam.PROMOCANCEL)
}

// PromoSettle implements Service
// @Id           RTPromoSettle
// @Summary      PromoSettle
// @Description  Promotion settlement for a placed buyin.
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       PayoutRequest  true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  PayoutResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/promo/settle [post]
func (s *WalletService) PromoSettle(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
	return handlePayout(s, req, pam.PROMODEPOSIT)
}

func handleStake(s *WalletService, req StakeRequest, transType pam.TransactionType) (*StakeResponseWrapper, *ErrorResponse) {
	existingTransactions, err := s.pamClient.GetTransactions(s.getTransactionsMapper(req.BaseRequest, req.Transaction.ID))
	if err != nil {
		var valkErr pam.ValkyrieError
		if errors.As(err, &valkErr) {
			if valkErr.ValkErrorCode != pam.ValkErrOpTransNotFound {
				e := createRtErrorResponse(err)
				return nil, &e
			}
		}
	}
	// no existing transaction
	if existingTransactions == nil {
		gameEvent, gErr := s.pamClient.GetGameRound(s.getGameRoundMapper(req.BaseRequest, req.Round.ID))
		if gErr != nil {
			var valkErr pam.ValkyrieError
			isValkErr := errors.As(gErr, &valkErr)
			if !isValkErr || isValkErr && valkErr.ValkErrorCode != pam.ValkErrOpRoundNotFound {
				e := createRtErrorResponse(gErr)
				return nil, &e
			}
		}
		if gameEvent != nil && gameEvent.ProviderRoundId == req.Round.ID && gameEvent.EndTime != nil {
			e := newRTErrorResponse("Additional bet on game round not allowed", InvalidInput)
			return nil, &e
		}
	}
	tranRes, err := s.pamClient.AddTransaction(s.getStakeTransactionMapper(req, transType))
	if err != nil {
		e := createRtErrorResponse(err)
		return nil, &e
	}
	if tranRes.Balance == nil {
		tranRes.Balance = &pam.Balance{
			BonusAmount: pam.ZeroAmount,
			CashAmount:  pam.ZeroAmount,
		}
	}
	response := StakeResponseWrapper{
		Response: Response{
			Success: true,
		},
		Result: StakeResponse{
			BaseResponse: BaseResponse{
				Token:    req.Token,
				Currency: req.Currency,
			},
			ID: *tranRes.TransactionId,
			Stake: Balance{
				Cash:  req.Transaction.Stake,
				Bonus: zeroMoney(),
			},
			Balance: Balance{
				Cash:  Money(tranRes.Balance.CashAmount),
				Bonus: Money(tranRes.Balance.BonusAmount),
			},
		},
	}

	return &response, nil
}

func handlePayout(s *WalletService, req PayoutRequest, transType pam.TransactionType) (*PayoutResponseWrapper, *ErrorResponse) {
	tranRes, err := s.pamClient.AddTransaction(s.getPayoutTransactionMapper(req, transType))
	if err != nil {
		e := createRtErrorResponse(err)
		return nil, &e
	}

	if tranRes.Balance == nil {
		tranRes.Balance = &pam.Balance{
			BonusAmount: pam.ZeroAmount,
			CashAmount:  pam.ZeroAmount,
		}
	}

	return &PayoutResponseWrapper{
		Response: Response{
			Success: true,
		},
		Result: PayoutResponse{
			BaseResponse: BaseResponse{
				Token:    req.Token,
				Currency: req.Currency,
			},
			ID: *tranRes.TransactionId,
			Payout: Balance{
				Cash:  req.Transaction.Payout,
				Bonus: zeroMoney(),
			},
			Balance: Balance{
				Cash:  Money(tranRes.Balance.CashAmount),
				Bonus: Money(tranRes.Balance.BonusAmount),
			},
		},
	}, nil
}

func handleRefund(s *WalletService, req RefundRequest, transType pam.TransactionType) (*RefundResponseWrapper, *ErrorResponse) {
	tranRes, err := s.pamClient.AddTransaction(s.getRefundTransactionMapper(req, transType))
	if err != nil {
		e := createRtErrorResponse(err)
		return nil, &e
	}

	if tranRes.Balance == nil {
		tranRes.Balance = &pam.Balance{
			BonusAmount: pam.ZeroAmount,
			CashAmount:  pam.ZeroAmount,
		}
	}

	resp := RefundResponseWrapper{
		Response: Response{
			Success: true,
		},
		Result: RefundResult{
			Token: req.Token,
			ID:    *tranRes.TransactionId,
			Stake: Balance{
				Cash:  req.Transaction.Stake,
				Bonus: zeroMoney(),
			},
		},
		Balance: Balance{
			Cash:  Money(tranRes.Balance.CashAmount),
			Bonus: Money(tranRes.Balance.BonusAmount),
		},
	}
	return &resp, nil
}
