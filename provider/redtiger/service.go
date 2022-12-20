package redtiger

import (
	"context"
	"errors"
	"fmt"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type ProviderService struct {
	pam.PamClient
	ctx context.Context
}

// NewService Create new red tiger provider service
func NewService(pamClient pam.PamClient) *ProviderService {
	return &ProviderService{PamClient: pamClient, ctx: context.Background()}
}

func (s *ProviderService) WithContext(ctx context.Context) Service {
	return &ProviderService{PamClient: s.PamClient, ctx: ctx}
}

// Auth implements Service
// @Id           RTAuth
// @Summary      Auth
// @Description  Authenticate
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body         AuthRequest  true  "Request body"
// @Param        Authorization     header       string       true  "API Key"
// @Success      200     {object}  AuthResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/auth [post]
func (s *ProviderService) Auth(req AuthRequest) (*AuthResponseWrapper, *ErrorResponse) {
	session, err := s.RefreshSession(s.refreshSessionRequestMapper(req))
	if err != nil {
		e := createRtErrorResponse(fmt.Errorf("Failed to Auth: %w", err))
		return nil, &e
	}

	// Use the new token returned by refresh session for subsequent calls
	req.Token = session.Token
	if req.UserID == "" {
		req.UserID = session.PlayerId
	}

	balance, err := s.GetBalance(s.getBalanceMapper(req.BaseRequest))
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
// @Description  Payout
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body         PayoutRequest  true  "Request body"
// @Param        Authorization     header       string         true  "API Key"
// @Success      200     {object}  PayoutResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/payout [post]
func (s *ProviderService) Payout(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
	return handlePayout(s, req, pam.DEPOSIT)

}

// Stake implements Service
// @Id           RTStake
// @Summary      Stake
// @Description  Stake
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       StakeRequest  true  "Request body"
// @Param        Authorization     header     string        true  "API Key"
// @Success      200     {object}  StakeResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/stake [post]
func (s *ProviderService) Stake(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
	return handleStake(s, req, pam.WITHDRAW)
}

// Refund implements Service
// @Id           RTRefund
// @Summary      Refund
// @Description  Refund
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       RefundRequest  true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  RefundResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/refund [post]
func (s *ProviderService) Refund(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
	return handleRefund(s, req, pam.CANCEL)
}

// PromoBuyin implements Service
// @Id           RTPromoBuyin
// @Summary      PromoBuyin
// @Description  PromoBuyin
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       StakeRequest   true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  StakeResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/promo/buyin [post]
func (s *ProviderService) PromoBuyin(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
	return handleStake(s, req, pam.PROMOWITHDRAW)
}

// PromoRefund implements Service
// @Id           RTPromoRefund
// @Summary      PromoRefund
// @Description  PromoRefund
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       RefundRequest  true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  RefundResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/promo/refund [post]
func (s *ProviderService) PromoRefund(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
	return handleRefund(s, req, pam.PROMOCANCEL)
}

// PromoSettle implements Service
// @Id           RTPromoSettle
// @Summary      PromoSettle
// @Description  PromoSettle
// @Tags         Red Tiger
// @Accept       json
// @Produce      json
// @Param        req               body       PayoutRequest  true  "Request body"
// @Param        Authorization     header     string         true  "API Key"
// @Success      200     {object}  PayoutResponseWrapper
// @Failure      400     {object}  ErrorResponse
// @Router       /providers/redtiger/promo/settle [post]
func (s *ProviderService) PromoSettle(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
	return handlePayout(s, req, pam.PROMODEPOSIT)
}

func handleStake(s *ProviderService, req StakeRequest, transType pam.TransactionType) (*StakeResponseWrapper, *ErrorResponse) {
	existingTransactions, err := s.GetTransactions(s.getTransactionsMapper(req.BaseRequest, req.Transaction.ID))
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
		gameEvent, gErr := s.GetGameRound(s.getGameRoundMapper(req.BaseRequest, req.Round.ID))
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
	tranRes, err := s.AddTransaction(s.getStakeTransactionMapper(req, transType))
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

func handlePayout(s *ProviderService, req PayoutRequest, transType pam.TransactionType) (*PayoutResponseWrapper, *ErrorResponse) {
	tranRes, err := s.AddTransaction(s.getPayoutTransactionMapper(req, transType))
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

func handleRefund(s *ProviderService, req RefundRequest, transType pam.TransactionType) (*RefundResponseWrapper, *ErrorResponse) {
	tranRes, err := s.AddTransaction(s.getRefundTransactionMapper(req, transType))
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
