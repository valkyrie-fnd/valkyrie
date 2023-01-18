package caleta

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

const (
	ProviderName = "Caleta"
)

func init() {
	provider.ProviderFactory().
		Register(ProviderName, func(args provider.ProviderArgs) (*provider.Router, error) {
			service := NewService(args.Client)
			return NewProviderRouter(args.Config, service)
		})
	provider.OperatorFactory().
		Register(ProviderName, func(args provider.OperatorArgs) (*provider.Router, error) {
			return NewOperatorRouter(args.Config, args.Client)
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

func NewOperatorRouter(config configs.ProviderConf, _ rest.HTTPClientJSONInterface) (*provider.Router, error) {
	service, err := NewStaticURLGameLaunchService(config)
	if err != nil {
		return nil, err
	}

	controller := provider.NewGameLaunchController(service)

	routes := []provider.Route{
		{
			Path:        "/gamelaunch",
			Method:      "POST",
			HandlerFunc: controller.GameLaunchEndpoint,
		},
		{
			Path:   "/api/v1/gamerounds/:gameRoundId/render",
			Method: "Get",
			HandlerFunc: func(c *fiber.Ctx) error {
				return nil
			},
		},
	}

	return &provider.Router{
		Name:        ProviderName,
		BasePath:    config.BasePath,
		Routes:      routes,
		Middlewares: []fiber.Handler{},
	}, nil
}
