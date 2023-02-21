package redtiger

import (
	"context"
	"time"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func (s *WalletService) refreshSessionRequestMapper(req AuthRequest) pam.RefreshSessionRequestMapper {
	return func() (context.Context, pam.RefreshSessionRequest, error) {
		return s.ctx, pam.RefreshSessionRequest{
			Params: pam.RefreshSessionParams{
				Provider:     ProviderName,
				XPlayerToken: req.Token,
			},
		}, nil
	}
}

func (s *WalletService) getBalanceMapper(req BaseRequest) pam.GetBalanceRequestMapper {
	return func() (context.Context, pam.GetBalanceRequest, error) {
		return s.ctx, pam.GetBalanceRequest{
			Params: pam.GetBalanceParams{
				Provider:     ProviderName,
				XPlayerToken: req.Token,
			},
			PlayerID: req.UserID,
		}, nil
	}
}

func (s *WalletService) getTransactionsMapper(req BaseRequest, providerTransactionID string) pam.GetTransactionsRequestMapper {
	return func() (context.Context, pam.GetTransactionsRequest, error) {
		return s.ctx, pam.GetTransactionsRequest{
			PlayerID: req.UserID,
			Params: pam.GetTransactionsParams{
				Provider:              ProviderName,
				XPlayerToken:          req.Token,
				ProviderTransactionId: &providerTransactionID,
			},
		}, nil
	}
}

func (s *WalletService) getGameRoundMapper(req BaseRequest, gameRoundID string) pam.GetGameRoundRequestMapper {
	return func() (context.Context, pam.GetGameRoundRequest, error) {
		return s.ctx, pam.GetGameRoundRequest{
			PlayerID:        req.UserID,
			ProviderRoundID: gameRoundID,
			Params: pam.GetGameRoundParams{
				Provider:     ProviderName,
				XPlayerToken: req.Token,
			},
		}, nil
	}
}

func (s *WalletService) getPayoutTransactionMapper(req PayoutRequest, transType pam.TransactionType) pam.AddTransactionRequestMapper {
	return func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		// Promo payouts are not necessarily tied to a game round. If no game round ID is passed,
		// we set it to nil and avoids closing the round
		roundID := &req.Round.ID
		isGameOver := req.Round.Ends
		if transType == pam.PROMODEPOSIT && req.Round.ID == "" {
			roundID = nil
			isGameOver = false
		}
		return s.ctx, &pam.AddTransactionRequest{
			PlayerID: req.UserID,
			Params: pam.AddTransactionParams{
				Provider:     ProviderName,
				XPlayerToken: req.Token,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            req.Transaction.Payout.toAmount(),
				BonusAmount:           pam.ZeroAmount,
				PromoAmount:           req.Transaction.PayoutPromo.toAmount(),
				Currency:              req.Currency,
				ProviderTransactionId: req.Transaction.ID,
				TransactionType:       transType,
				TransactionDateTime:   time.Now(),
				ProviderGameId:        &req.Game.Key,
				ProviderRoundId:       roundID,
				IsGameOver:            &isGameOver,
				Provider:              ProviderName,
			},
		}, nil
	}
}
func (s *WalletService) getStakeTransactionMapper(req StakeRequest, transType pam.TransactionType) pam.AddTransactionRequestMapper {
	return func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		return s.ctx, &pam.AddTransactionRequest{
			PlayerID: req.UserID,
			Params: pam.AddTransactionParams{
				Provider:     ProviderName,
				XPlayerToken: req.Token,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            req.Transaction.Stake.toAmount(),
				BonusAmount:           pam.ZeroAmount,
				PromoAmount:           req.Transaction.StakePromo.toAmount(),
				Currency:              req.Currency,
				ProviderTransactionId: req.Transaction.ID,
				TransactionType:       transType,
				TransactionDateTime:   time.Now(),
				ProviderGameId:        &req.Game.Key,
				ProviderRoundId:       &req.Round.ID,
				IsGameOver:            &req.Round.Ends,
				Provider:              ProviderName,
			},
		}, nil
	}
}
func (s *WalletService) getRefundTransactionMapper(req RefundRequest, transType pam.TransactionType) pam.AddTransactionRequestMapper {
	return func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		return s.ctx, &pam.AddTransactionRequest{
			PlayerID: req.UserID,
			Params: pam.AddTransactionParams{
				Provider:     ProviderName,
				XPlayerToken: req.Token,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            req.Transaction.Stake.toAmount(),
				BonusAmount:           pam.ZeroAmount,
				PromoAmount:           req.Transaction.StakePromo.toAmount(),
				Currency:              req.Currency,
				ProviderTransactionId: req.Transaction.ID,
				TransactionType:       transType,
				TransactionDateTime:   time.Now(),
				ProviderGameId:        &req.Game.Key,
				ProviderRoundId:       &req.Round.ID,
				IsGameOver:            &req.Round.Ends,
				Provider:              ProviderName,
			},
		}, nil
	}
}
