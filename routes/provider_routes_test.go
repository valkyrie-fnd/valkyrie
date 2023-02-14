package routes

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/pam"
)

func Test_ProviderRoutes(t *testing.T) {
	tests := []struct {
		name         string
		conf         configs.ProviderConf
		wantHandlers int
		pamClient    pam.PamClient
	}{
		{
			name: "Evolution",
			conf: configs.ProviderConf{
				Name: "Evolution",
				Auth: map[string]any{
					"api_key":      "",
					"casino_token": "",
					"casino_key":   "",
				},
				URL: "url",
			},
			wantHandlers: 9,
			pamClient:    &mockPamClient{},
		},
		{
			name: "Red Tiger",
			conf: configs.ProviderConf{
				Name: "Red Tiger",
				Auth: map[string]any{
					"api_key":     "",
					"recon_token": "",
				},
				URL: "url",
			},
			wantHandlers: 12,
			pamClient:    &mockPamClient{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			app := fiber.New()
			err := ProviderRoutes(app, &configs.ValkyrieConfig{Providers: []configs.ProviderConf{test.conf}}, test.pamClient, nil)
			assert.NoError(tt, err)
			assert.Equal(tt, test.wantHandlers, int(app.HandlersCount()))
		})
	}
}

func Test_OperatorRoutes(t *testing.T) {
	tests := []struct {
		name         string
		conf         configs.ProviderConf
		wantHandlers int
	}{
		{
			name: "Evolution",
			conf: configs.ProviderConf{
				Name: "Evolution",
				Auth: map[string]any{
					"api_key":      "",
					"casino_token": "",
					"casino_key":   "",
				},
				URL: "url",
			},
			wantHandlers: 4,
		},
		{
			name: "Red Tiger",
			conf: configs.ProviderConf{
				Name: "Red Tiger",
				Auth: map[string]any{
					"api_key":     "",
					"recon_token": "",
				},
				URL: "url",
			},
			wantHandlers: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			app := fiber.New()
			err := OperatorRoutes(app, &configs.ValkyrieConfig{Providers: []configs.ProviderConf{test.conf}}, nil)
			assert.NoError(tt, err)
			assert.Equal(tt, test.wantHandlers, int(app.HandlersCount()))
		})
	}
}

type mockPamClient struct{}

// GetSession Return session
func (p *mockPamClient) GetSession(_ pam.GetSessionRequestMapper) (*pam.Session, error) {
	panic("not implemented") // TODO: Implement
}

// RefreshSession returns a new session token
func (p *mockPamClient) RefreshSession(_ pam.RefreshSessionRequestMapper) (*pam.Session, error) {
	panic("not implemented") // TODO: Implement
}

// GetBalance get balance from PAM
func (p *mockPamClient) GetBalance(_ pam.GetBalanceRequestMapper) (*pam.Balance, error) {
	panic("not implemented") // TODO: Implement
}

// GetTransactions get transactions from pam
func (p *mockPamClient) GetTransactions(_ pam.GetTransactionsRequestMapper) ([]pam.Transaction, error) {
	panic("not implemented") // TODO: Implement
}

// AddTransaction returns transactionId and balance. When transaction fails balance can still be returned. On failure error will be returned
func (p *mockPamClient) AddTransaction(_ pam.AddTransactionRequestMapper) (*pam.TransactionResult, error) {
	panic("not implemented") // TODO: Implement
}

// GetGameRound gets gameRound from PAM
func (p *mockPamClient) GetGameRound(_ pam.GetGameRoundRequestMapper) (*pam.GameRound, error) {
	panic("not implemented") // TODO: Implement
}

// GetSettlementType returns the type of settlement the PAM supports
func (p *mockPamClient) GetSettlementType() pam.SettlementType {
	panic("not implemented") // TODO: Implement
}

// GetTransactionHandling return the type of transaction handling the PAM supports
func (p *mockPamClient) GetTransactionHandling() pam.TransactionHandling {
	return pam.OPERATOR
}
