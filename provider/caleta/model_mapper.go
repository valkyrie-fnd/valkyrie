package caleta

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func balanceRequestMapper(ctx context.Context, r *WalletbalanceJSONRequestBody) pam.GetBalanceRequestMapper {
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

func valueOrNow(t *MsgTimestamp) time.Time {
	if t == nil {
		return time.Now().UTC()
	} else {
		return t.toTime()
	}
}

func betTransactionMapper(ctx context.Context, r *WalletbetRequestObject) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(r.Body.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.Body.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Body.Token,
				XCorrelationID: r.Body.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            amt,
				Currency:              string(r.Body.Currency),
				IsGameOver:            &r.Body.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.Body.GameCode,
				ProviderRoundId:       &r.Body.Round,
				ProviderTransactionId: r.Body.TransactionUuid,
				ProviderBetRef:        &r.Body.TransactionUuid,
				TransactionDateTime:   valueOrNow(r.Params.XMsgTimestamp),
				TransactionType:       pam.WITHDRAW,
				BetCode:               r.Body.Bet,
			},
		}, nil
	}
}
func promoBetTransactionMapper(ctx context.Context, r *WalletbetRequestObject) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(r.Body.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.Body.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Body.Token,
				XCorrelationID: r.Body.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				PromoAmount:           amt,
				Currency:              string(r.Body.Currency),
				IsGameOver:            &r.Body.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.Body.GameCode,
				ProviderRoundId:       &r.Body.Round,
				ProviderTransactionId: r.Body.TransactionUuid,
				ProviderBetRef:        &r.Body.TransactionUuid,
				TransactionDateTime:   valueOrNow(r.Params.XMsgTimestamp),
				TransactionType:       pam.PROMOWITHDRAW,
				BetCode:               r.Body.Bet,
			},
		}, nil
	}
}

func winTransactionMapper(ctx context.Context, r *TransactionwinRequestObject, roundTransactions *[]roundTransaction) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(r.Body.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.Body.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Body.Token,
				XCorrelationID: r.Body.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            amt,
				Currency:              string(r.Body.Currency),
				IsGameOver:            &r.Body.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.Body.GameCode,
				ProviderRoundId:       &r.Body.Round,
				ProviderTransactionId: r.Body.TransactionUuid,
				ProviderBetRef:        &r.Body.ReferenceTransactionUuid,
				TransactionDateTime:   valueOrNow(r.Params.XMsgTimestamp),
				TransactionType:       pam.DEPOSIT,
				BetCode:               r.Body.Bet,
				RoundTransactions:     roundTransactionsMapper(roundTransactions),
			},
		}, nil
	}
}
func promoWinTransactionMapper(ctx context.Context, r *TransactionwinRequestObject, roundTransactions *[]roundTransaction) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		amt := toPamAmount(r.Body.Amount)
		return ctx, &pam.AddTransactionRequest{
			PlayerID: r.Body.SupplierUser,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Body.Token,
				XCorrelationID: r.Body.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				PromoAmount:           amt,
				Currency:              string(r.Body.Currency),
				IsGameOver:            &r.Body.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.Body.GameCode,
				ProviderRoundId:       &r.Body.Round,
				ProviderTransactionId: r.Body.TransactionUuid,
				ProviderBetRef:        &r.Body.ReferenceTransactionUuid,
				TransactionDateTime:   valueOrNow(r.Params.XMsgTimestamp),
				TransactionType:       pam.PROMODEPOSIT,
				BetCode:               r.Body.Bet,
				RoundTransactions:     roundTransactionsMapper(roundTransactions),
			},
		}, nil
	}
}

func cancelTransactionMapper(ctx context.Context, r *WalletrollbackRequestObject, session *pam.Session, tt pam.TransactionType, roundTransactions *[]roundTransaction) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		return ctx, &pam.AddTransactionRequest{
			PlayerID: utils.OrZeroValue(r.Body.User),
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.Body.Token,
				XCorrelationID: r.Body.RequestUuid,
			},
			Body: pam.AddTransactionJSONRequestBody{
				Currency:              session.Currency,
				IsGameOver:            &r.Body.RoundClosed,
				Provider:              ProviderName,
				ProviderGameId:        &r.Body.GameCode,
				ProviderRoundId:       &r.Body.Round,
				ProviderTransactionId: r.Body.TransactionUuid,
				ProviderBetRef:        &r.Body.ReferenceTransactionUuid,
				TransactionDateTime:   valueOrNow(r.Params.XMsgTimestamp),
				TransactionType:       tt,
				RoundTransactions:     roundTransactionsMapper(roundTransactions),
			},
		}, nil
	}
}

func getTransactionsMapper(ctx context.Context, r *WalletWinBody) pam.GetTransactionsRequestMapper {
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

func roundTransactionsMapper(roundTransactions *[]roundTransaction) *[]pam.RoundTransaction {
	if roundTransactions != nil {
		var roundTx []pam.RoundTransaction
		for _, t := range *roundTransactions {
			amt := toPamAmount(t.Payload.Amount)
			txID := t.Payload.TransactionUUID
			isGameOver := t.Payload.RoundClosed
			refTx := t.Payload.ReferenceTransactionUUID

			// Detect transaction type
			var transactionType pam.TransactionType
			switch {
			case refTx == nil:
				transactionType = pam.WITHDRAW
			case t.Payload.Currency != "" && t.Payload.Bet != "":
				transactionType = pam.DEPOSIT
			case t.Payload.Bet == "" && t.Payload.Currency == "":
				transactionType = pam.CANCEL
			default:
				log.Warn().Msgf("Failed to detect transactionType for txID=%s, skipping transaction", txID)
				continue
			}

			tx := pam.RoundTransaction{
				ProviderTransactionId: &txID,
				CashAmount:            &amt,
				IsGameOver:            &isGameOver,
				TransactionDateTime:   &t.CreatedTime,
				ProviderBetRef:        refTx,
				TransactionType:       transactionType,
			}

			jackpotAmt := toPamAmount(t.Payload.JackpotContribution)
			if t.Payload.JackpotContribution != 0 {
				tx.JackpotContribution = &jackpotAmt
			}

			roundTx = append(roundTx, tx)
		}
		return &roundTx
	} else {
		return nil
	}
}
