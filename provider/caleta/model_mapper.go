package caleta

import (
	"context"
	"time"

	"github.com/valkyrie-fnd/valkyrie-stubs/utils"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func balanceRequestMapper(ctx context.Context, r WalletbalanceJSONRequestBody) pam.GetBalanceRequestMapper {
	return func() (context.Context, pam.GetBalanceRequest, error) {
		return ctx, pam.GetBalanceRequest{
			PlayerID: r.SupplierUser,
			Params: pam.GetBalanceParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Token,
				XCorrelationID: r.RequestUuid,
			},
		}, nil
	}
}

func getSessionMapper(ctx context.Context, token, requestID string) pam.GetSessionRequestMapper {
	return func() (context.Context, pam.GetSessionRequest, error) {
		return ctx, pam.GetSessionRequest{
			Params: pam.GetSessionParams{
				Provider:       ProviderName,
				XPlayerToken:   token,
				XCorrelationID: requestID,
			},
		}, nil
	}
}

func refreshSessionMapper(ctx context.Context, token, requestID string) pam.RefreshSessionRequestMapper {
	return func() (context.Context, pam.RefreshSessionRequest, error) {
		return ctx, pam.RefreshSessionRequest{
			Params: pam.RefreshSessionParams{
				Provider:       ProviderName,
				XPlayerToken:   token,
				XCorrelationID: requestID,
			},
		}, nil
	}
}

func betTransactionMapper(ctx context.Context, r WalletbetJSONRequestBody) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(&r.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Token,
				XCorrelationID: r.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            amt,
				Currency:              string(r.Currency),
				IsGameOver:            &r.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.GameCode,
				ProviderRoundId:       &r.Round,
				ProviderTransactionId: r.TransactionUuid,
				ProviderBetRef:        &r.TransactionUuid,
				TransactionDateTime:   time.Now(),
				TransactionType:       pam.WITHDRAW,
			},
		}, nil
	}
}
func promoBetTransactionMapper(ctx context.Context, r WalletbetJSONRequestBody) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(&r.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Token,
				XCorrelationID: r.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				PromoAmount:           amt,
				Currency:              string(r.Currency),
				IsGameOver:            &r.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.GameCode,
				ProviderRoundId:       &r.Round,
				ProviderTransactionId: r.TransactionUuid,
				ProviderBetRef:        &r.TransactionUuid,
				TransactionDateTime:   time.Now(),
				TransactionType:       pam.PROMOWITHDRAW,
			},
		}, nil
	}
}

func winTransactionMapper(ctx context.Context, r WalletWinBody) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(&r.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Token,
				XCorrelationID: r.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            amt,
				Currency:              string(r.Currency),
				IsGameOver:            &r.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.GameCode,
				ProviderRoundId:       &r.Round,
				ProviderTransactionId: r.TransactionUuid,
				ProviderBetRef:        &r.ReferenceTransactionUuid,
				TransactionDateTime:   time.Now(),
				TransactionType:       pam.DEPOSIT,
			},
		}, nil
	}
}
func promoWinTransactionMapper(ctx context.Context, r WalletWinBody) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(&r.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Token,
				XCorrelationID: r.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				PromoAmount:           amt,
				Currency:              string(r.Currency),
				IsGameOver:            &r.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.GameCode,
				ProviderRoundId:       &r.Round,
				ProviderTransactionId: r.TransactionUuid,
				ProviderBetRef:        &r.ReferenceTransactionUuid,
				TransactionDateTime:   time.Now(),
				TransactionType:       pam.PROMODEPOSIT,
			},
		}, nil
	}
}

func cancelTransactionMapper(ctx context.Context, r WalletrollbackJSONRequestBody, session *pam.Session, tt pam.TransactionType) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		return ctx, &pam.AddTransactionRequest{
			PlayerID: utils.OrZeroValue(r.User),
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Token,
				XCorrelationID: r.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				Currency:              session.Currency,
				IsGameOver:            &r.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.GameCode,
				ProviderRoundId:       &r.Round,
				ProviderTransactionId: r.TransactionUuid,
				ProviderBetRef:        &r.ReferenceTransactionUuid,
				TransactionDateTime:   time.Now(),
				TransactionType:       tt,
			},
		}, nil
	}
}

func getTransactionsMapper(ctx context.Context, r WalletWinBody) pam.GetTransactionsRequestMapper {
	return func() (context.Context, pam.GetTransactionsRequest, error) {
		return ctx, pam.GetTransactionsRequest{
			PlayerID: r.SupplierUser,
			Params: pam.GetTransactionsParams{
				Provider:              ProviderName,
				XPlayerToken:          r.Token,
				ProviderTransactionId: &r.ReferenceTransactionUuid,
				ProviderBetRef:        &r.ReferenceTransactionUuid,
				XCorrelationID:        r.RequestUuid,
			},
		}, nil
	}
}
