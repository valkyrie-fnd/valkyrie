package routes

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

func Test_ProviderRoutes(t *testing.T) {
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
			wantHandlers: 9,
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
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			app := fiber.New()
			err := ProviderRoutes(app, &configs.ValkyrieConfig{Providers: []configs.ProviderConf{test.conf}}, nil, nil)
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
