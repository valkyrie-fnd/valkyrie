package example

import (
	"context"
	"time"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

// WalletService is used to map the providers wallet calls to Valkyrie PamClient.
type WalletService struct {
	pamClient pam.PamClient
	ctx       context.Context
}

func NewWalletService(pamClient pam.PamClient) *WalletService {
	return &WalletService{pamClient: pamClient, ctx: context.Background()}
}

func (ws *WalletService) GetBalance(ctx context.Context, r balanceRequest) (*pam.Balance, error) {
	session, err := ws.pamClient.GetSession(getSessionMapper(ctx, r.baseRequest))
	if err != nil {
		return nil, err
	}

	// Additional requests or checks can be made here. Any custom quirks needed to execute eg a balance call to a PAM.
	balance, err := ws.pamClient.GetBalance(getBalanceMapper(ctx, session))
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (ws *WalletService) Auth(ctx context.Context, r authRequest) (*pam.Session, error) {
	session, err := ws.pamClient.RefreshSession(refreshSessionMapper(ctx, r.baseRequest))
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (ws *WalletService) Bet(ctx context.Context, r betRequest) (*pam.Balance, error) {
	session, err := ws.pamClient.GetSession(getSessionMapper(ctx, r.baseRequest))
	if err != nil {
		return nil, err
	}

	res, err := ws.pamClient.AddTransaction(betMapper(ctx, session, r))
	if err != nil {
		return nil, err
	}

	return res.Balance, nil
}

func (ws *WalletService) Win(ctx context.Context, r winRequest) (*pam.Balance, error) {
	session, err := ws.pamClient.GetSession(getSessionMapper(ctx, r.baseRequest))
	if err != nil {
		return nil, err
	}

	res, err := ws.pamClient.AddTransaction(winMapper(ctx, session, r))
	if err != nil {
		return nil, err
	}

	return res.Balance, nil
}

func getSessionMapper(ctx context.Context, r baseRequest) pam.GetSessionRequestMapper {
	return func() (context.Context, pam.GetSessionRequest, error) {
		return ctx, pam.GetSessionRequest{
			Params: pam.GetSessionParams{
				Provider:     ProviderName,
				XPlayerToken: r.Token,
			},
		}, nil
	}
}

func refreshSessionMapper(ctx context.Context, r baseRequest) pam.RefreshSessionRequestMapper {
	return func() (context.Context, pam.RefreshSessionRequest, error) {
		return ctx, pam.RefreshSessionRequest{
			Params: pam.RefreshSessionParams{
				Provider:     ProviderName,
				XPlayerToken: r.Token,
			},
		}, nil
	}
}

// map the session to a "pam.GetBalanceRequest" used by valkyrie PAM implementation.
func getBalanceMapper(ctx context.Context, session *pam.Session) pam.GetBalanceRequestMapper {
	return func() (context.Context, pam.GetBalanceRequest, error) {
		return ctx, pam.GetBalanceRequest{
			Params: pam.GetBalanceParams{
				Provider:     ProviderName,
				XPlayerToken: session.Token,
			},
			PlayerID: session.PlayerId,
		}, nil
	}
}

func betMapper(ctx context.Context, session *pam.Session, r betRequest) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		return ctx, &pam.AddTransactionRequest{
			Params: pam.AddTransactionParams{
				Provider:     ProviderName,
				XPlayerToken: r.Token,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            r.Amount,
				Currency:              r.Currency,
				Provider:              ProviderName,
				ProviderGameId:        &r.GameID,
				ProviderRoundId:       &r.RoundID,
				ProviderTransactionId: r.TransactionID,
				TransactionDateTime:   time.Now(),
				TransactionType:       pam.WITHDRAW,
			},
			PlayerID: session.PlayerId,
		}, nil
	}
}

func winMapper(ctx context.Context, session *pam.Session, r winRequest) pam.AddTransactionRequestMapper {
	return func(_ pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
		return ctx, &pam.AddTransactionRequest{
			Params: pam.AddTransactionParams{
				Provider:     ProviderName,
				XPlayerToken: r.Token,
			},
			Body: pam.AddTransactionJSONRequestBody{
				CashAmount:            r.Amount,
				Currency:              r.Currency,
				Provider:              ProviderName,
				ProviderGameId:        &r.GameID,
				ProviderRoundId:       &r.RoundID,
				ProviderTransactionId: r.TransactionID,
				TransactionDateTime:   time.Now(),
				TransactionType:       pam.DEPOSIT,
			},
			PlayerID: session.PlayerId,
		}, nil
	}
}
