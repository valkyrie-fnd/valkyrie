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
				Name: "evolution",
				Auth: map[string]any{
					"api_key":      "",
					"casino_token": "",
					"casino_key":   "",
				},
				URL: "url",
			},
			wantHandlers: 5,
		},
		{
			name: "Red Tiger",
			conf: configs.ProviderConf{
				Name: "redtiger",
				Auth: map[string]any{
					"api_key":     "",
					"recon_token": "",
				},
				URL: "url",
			},
			wantHandlers: 5,
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

type mockPamClient struct {
	pam.PamClient
}

func (p *mockPamClient) GetTransactionSupplier() pam.TransactionSupplier {
	return pam.OPERATOR
}
