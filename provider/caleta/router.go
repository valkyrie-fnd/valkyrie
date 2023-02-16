package caleta

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

const (
	ProviderName = "Caleta"
)

func init() {
	provider.ProviderFactory().
		Register(ProviderName, func(args provider.ProviderArgs) (*provider.Router, error) {

			var service *WalletService

			// If transaction supplier is PROVIDER, provide a transaction client
			if args.PamClient.GetTransactionSupplier() == pam.PROVIDER {
				apiClient, err := NewAPIClient(args.HTTPClient, args.Config)
				if err != nil {
					return nil, err
				}
				service = NewWalletService(args.PamClient, apiClient)
			} else {
				service = NewWalletService(args.PamClient, nil)
			}

			log.Info().Msgf("Configured for transaction supplier '%s'", args.PamClient.GetTransactionSupplier())

			return NewProviderRouter(args.Config, service)
		})
	provider.OperatorFactory().
		Register(ProviderName, func(args provider.OperatorArgs) (*provider.Router, error) {
			return NewOperatorRouter(args.Config, args.HTTPClient)
		})
}

func NewProviderRouter(config configs.ProviderConf, service StrictServerInterface) (*provider.Router, error) {
	auth, err := getAuthConf(config)
	if err != nil {
		return nil, err
	}

	middlewares, err := getProviderMiddlewares(auth)
	if err != nil {
		return nil, err
	}

	routes := Routes(ServerInterfaceWrapper{
		Handler: NewStrictHandler(service, []StrictMiddlewareFunc{}),
	})
	return &provider.Router{
		Name:        ProviderName,
		BasePath:    config.BasePath,
		Routes:      routes,
		Middlewares: middlewares,
	}, nil
}

func getProviderMiddlewares(auth AuthConf) ([]fiber.Handler, error) {
	middlewares := []fiber.Handler{}

	if auth.VerificationKey != "" {
		verifier, err := NewVerifier([]byte(auth.VerificationKey))
		if err != nil {
			return nil, err
		}
		middlewares = append(middlewares, VerifySignature(verifier))
	} else {
		log.Warn().Msg("Missing Caleta provider 'verification_key' config, skipping signature verification middleware")
	}
	return middlewares, nil
}

func NewOperatorRouter(config configs.ProviderConf, httpClient rest.HTTPClientJSONInterface) (*provider.Router, error) {
	apiClient, err := NewAPIClient(httpClient, config)
	if err != nil {
		return nil, err
	}

	caletaService, err := NewCaletaService(apiClient, config)
	if err != nil {
		return nil, err
	}

	controller := provider.NewGameLaunchController(caletaService)
	grCtrl := provider.NewGameRoundController(caletaService)

	routes := []provider.Route{
		{
			Path:        "/gamelaunch",
			Method:      "POST",
			HandlerFunc: controller.GameLaunchEndpoint,
		},
		{
			Path:        "/gamerounds/:gameRoundId/render",
			Method:      "GET",
			HandlerFunc: grCtrl.GetGameRoundEndpoint,
		},
	}

	return &provider.Router{
		Name:        ProviderName,
		BasePath:    config.BasePath,
		Routes:      routes,
		Middlewares: []fiber.Handler{},
	}, nil
}
