package evolution

import (
	"context"
	"time"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

// Functions encapsulating the work of mapping provider specific requests (plus additionally
// fetched data, if required) to pam equivalents

func (service *ProviderService) balanceRequestMapper(r RequestBase) pam.GetBalanceRequestMapper {
	return func() (context.Context, pam.GetBalanceRequest, error) {
		return service.ctx, pam.GetBalanceRequest{
			Params: pam.GetBalanceParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				XCorrelationID: r.UUID,
			},
			PlayerID: r.UserID,
		}, nil
	}
}

func (service *ProviderService) refreshSessionRequestMapper(r CheckRequest) pam.RefreshSessionRequestMapper {
	return func() (context.Context, pam.RefreshSessionRequest, error) {
		return service.ctx, pam.RefreshSessionRequest{
			Params: pam.RefreshSessionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				XCorrelationID: r.UUID,
			},
		}, nil
	}
}

func (service *ProviderService) debitRequestMapper(r DebitRequest, transTime time.Time) pam.AddTransactionRequestMapper {
	return func(rounder pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		cashAmt, roundingErr := rounder(r.Transaction.Amount.toAmt())
		if roundingErr != nil {
			return service.ctx, nil, roundingErr
		}

		return service.ctx, &pam.AddTransactionRequest{
			PlayerID: r.UserID,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				XCorrelationID: r.UUID,
			},
			Body: pam.AddTransactionJSONRequestBody{
				Currency:              r.Currency,
				CashAmount:            *cashAmt,
				BonusAmount:           pam.ZeroAmount,
				PromoAmount:           pam.ZeroAmount,
				TransactionType:       pam.WITHDRAW,
				ProviderTransactionId: r.Transaction.ID,
				ProviderBetRef:        &r.Transaction.RefID,
				ProviderGameId:        &r.Game.Details.Table.ID,
				ProviderRoundId:       &r.Game.ID,
				TransactionDateTime:   transTime,
				Provider:              ProviderName,
			},
		}, nil
	}
}

func (service *ProviderService) creditTransRequestMapper(r CreditRequest, transTime time.Time) pam.AddTransactionRequestMapper {
	return func(rounder pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		cashAmt, roundingErr := rounder(r.Transaction.Amount.toAmt())
		if roundingErr != nil {
			return service.ctx, nil, roundingErr
		}

		return service.ctx, &pam.AddTransactionRequest{
			PlayerID: r.UserID,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				XCorrelationID: r.UUID,
			},
			Body: pam.AddTransactionJSONRequestBody{
				Currency:              r.Currency,
				CashAmount:            *cashAmt,
				BonusAmount:           pam.ZeroAmount,
				PromoAmount:           pam.ZeroAmount,
				TransactionType:       pam.DEPOSIT,
				ProviderTransactionId: r.Transaction.ID,
				ProviderBetRef:        &r.Transaction.RefID,
				ProviderGameId:        &r.Game.Details.Table.ID,
				ProviderRoundId:       &r.Game.ID,
				TransactionDateTime:   transTime,
				Provider:              ProviderName,
			},
		}, nil
	}
}

func (service *ProviderService) cancelTransRequestMapper(r CancelRequest, transTime time.Time) pam.AddTransactionRequestMapper {
	return func(rounder pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		cashAmt, roundingErr := rounder(r.Transaction.Amount.toAmt())
		if roundingErr != nil {
			return service.ctx, nil, roundingErr
		}

		return service.ctx, &pam.AddTransactionRequest{
			PlayerID: r.UserID,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				XCorrelationID: r.UUID,
			},
			Body: pam.AddTransactionJSONRequestBody{
				Currency:              r.Currency,
				CashAmount:            *cashAmt,
				BonusAmount:           pam.ZeroAmount,
				PromoAmount:           pam.ZeroAmount,
				TransactionType:       pam.CANCEL,
				ProviderTransactionId: r.Transaction.ID,
				ProviderBetRef:        &r.Transaction.RefID,
				ProviderGameId:        &r.Game.Details.Table.ID,
				ProviderRoundId:       &r.Game.ID,
				TransactionDateTime:   transTime,
				Provider:              ProviderName,
			},
		}, nil
	}
}

func (service *ProviderService) promoPayoutTransRequestMapper(r PromoPayoutRequest, transTime time.Time) pam.AddTransactionRequestMapper {
	return func(round pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		roundedPromoAmount, roundingErr := round(r.PromoTransaction.Amount.toAmt())

		if roundingErr != nil {
			return service.ctx, nil, roundingErr
		}

		return service.ctx, &pam.AddTransactionRequest{
			PlayerID: r.UserID,
			Params: pam.AddTransactionParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				XCorrelationID: r.UUID,
			},
			Body: pam.AddTransactionJSONRequestBody{
				Currency:              r.Currency,
				CashAmount:            *roundedPromoAmount,
				BonusAmount:           pam.ZeroAmount,
				PromoAmount:           pam.ZeroAmount,
				TransactionType:       pam.PROMODEPOSIT,
				ProviderTransactionId: r.PromoTransaction.ID,
				ProviderGameId:        &r.Game.Details.Table.ID,
				ProviderRoundId:       &r.Game.ID,
				TransactionDateTime:   transTime,
				Provider:              ProviderName,
			},
		}, nil
	}
}

func (service *ProviderService) findTransForCreditRequestMapper(r CreditRequest) pam.GetTransactionsRequestMapper {
	return func() (context.Context, pam.GetTransactionsRequest, error) {
		return service.ctx, pam.GetTransactionsRequest{
			PlayerID: r.UserID,
			Params: pam.GetTransactionsParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				ProviderBetRef: &r.Transaction.RefID,
				XCorrelationID: r.UUID,
			},
		}, nil
	}
}

func (service *ProviderService) findTransForCancelRequestMapper(r CancelRequest) pam.GetTransactionsRequestMapper {
	return func() (context.Context, pam.GetTransactionsRequest, error) {
		return service.ctx, pam.GetTransactionsRequest{
			PlayerID: r.UserID,
			Params: pam.GetTransactionsParams{
				Provider:       ProviderName,
				XPlayerToken:   r.SID,
				ProviderBetRef: &r.Transaction.RefID,
				XCorrelationID: r.UUID,
			},
		}, nil
	}
}
