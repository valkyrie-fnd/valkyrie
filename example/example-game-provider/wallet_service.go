package example

import (
	"context"

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

func (ws *WalletService) GetBalance(br balanceRequest) *pam.Balance {
	// Additional requests or checks can be made here. Any custom quirks needed to execute eg a balance call to a PAM.
	balance, _ := ws.pamClient.GetBalance(ws.getBalanceMapper(br))
	return balance
}

// map the provider specific request to a "pam.GetBalanceRequest" used by valkyrie PAM implementation.
func (ws *WalletService) getBalanceMapper(br balanceRequest) pam.GetBalanceRequestMapper {
	return func() (context.Context, pam.GetBalanceRequest, error) {
		return ws.ctx, pam.GetBalanceRequest{
			PlayerID: br.PlayerID,
		}, nil
	}
}
