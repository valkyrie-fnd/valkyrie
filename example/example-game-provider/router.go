package example

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

const ProviderName = "Example Game Provider"

// init will register the provider and operator endpoints.
// If this provider is part of valkyrie config they will be exposed by Valkyrie
func init() {
	// Registering the provider endpoints.
	provider.ProviderFactory().
		Register(ProviderName, func(args provider.ProviderArgs) (*provider.Router, error) {
			walletService := NewWalletService(args.PamClient)
			return NewProviderRouter(args.Config, walletService)
		})
	// Registering the operator endpoints. Requests made from the operator toward the game provider
	provider.OperatorFactory().
		Register(ProviderName, func(args provider.OperatorArgs) (*provider.Router, error) {
			return NewOperatorRouter(args.Config, args.HTTPClient), nil
		})
}

// NewProviderRouter sets up the wallet api used by the Game provider.
func NewProviderRouter(config configs.ProviderConf, service *WalletService) (*provider.Router, error) {
	// Create a controller and setup all routes used by the providers wallet api
	// If your Api is defined in an OAPI definition, it is possible to generate the Controller code using a forked version of oapi-codegen
	// Provider Caleta is using this. Read oapi-codegen documentation as well as check out Caleta implementation.
	// Make sure to view handles.cfg.yml, models.cfg.yml, handles.gen.go and router.go. The last one where the generated code is used.
	controller := NewController(service)
	routes := []provider.Route{
		{
			Path:   "/balance",
			Method: "POST",
			// The endpoint function
			HandlerFunc: controller.WalletBalanceEndpoint,
			// any middlewares to be used only for this endpoint
			Middlewares: []fiber.Handler{},
		},
	}
	return &provider.Router{
		Name:     ProviderName,
		BasePath: config.BasePath,
		Routes:   routes,
		// middlewares to be used by all endpoints, like authentication
		Middlewares: []fiber.Handler{},
	}, nil
}

// NewOperatorRouter sets up all endpoints that can be used by the operator to make requests toward the provider.
// The router should follow the oapi definition found in /provider/docs/operator_api.yml
func NewOperatorRouter(config configs.ProviderConf, httpClient rest.HTTPClientJSONInterface) *provider.Router {
	// Provide an implementation of provider.ProviderService
	providerService := NewExampleProviderService(config, httpClient)

	gameLaunchController := provider.NewGameLaunchController(providerService)
	gameRoundRenderController := provider.NewGameRoundController(providerService)
	routes := []provider.Route{
		{
			Path:        "/gamelaunch",
			Method:      "POST",
			HandlerFunc: gameLaunchController.GameLaunchEndpoint,
			Middlewares: []fiber.Handler{},
		},
		{
			Path:        "/gamerounds/:gameRoundId/render",
			Method:      "GET",
			HandlerFunc: gameRoundRenderController.GetGameRoundEndpoint,
			Middlewares: []fiber.Handler{},
		},
	}
	return &provider.Router{
		Name:        ProviderName,
		BasePath:    config.BasePath,
		Routes:      routes,
		Middlewares: []fiber.Handler{},
	}
}
