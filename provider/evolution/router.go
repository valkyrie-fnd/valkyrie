package evolution

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

const (
	ProviderName      = "evolution"
	apiTokenParamName = "authToken"
)

func init() {
	provider.ProviderFactory().
		Register(ProviderName, func(args provider.ProviderArgs) (*provider.Router, error) {
			if args.PamClient.GetTransactionSupplier() == pam.PROVIDER {
				return nil, fmt.Errorf("Unsupported transaction supplier")
			}
			service := NewService(args.PamClient)
			controller := NewProviderController(service)
			return NewProviderRouter(args.Config, controller)
		})
	provider.OperatorFactory().
		Register(ProviderName, func(args provider.OperatorArgs) (*provider.Router, error) {
			return NewOperatorRouter(args.Config, args.HTTPClient)
		})
}

type Controller interface {
	Check(c *fiber.Ctx) error
	Balance(c *fiber.Ctx) error
	Debit(c *fiber.Ctx) error
	Credit(c *fiber.Ctx) error
	Cancel(c *fiber.Ctx) error
	PromoPayout(c *fiber.Ctx) error
}

func NewProviderRouter(config configs.ProviderConf, controller Controller) (*provider.Router, error) {
	auth, err := GetAuthConf(config)
	if err != nil {
		return nil, err
	}
	// Define the routes
	routes := []provider.Route{
		{
			Path:        "/check",
			Method:      "POST",
			HandlerFunc: controller.Check,
		},
		{
			Path:        "/balance",
			Method:      "POST",
			HandlerFunc: controller.Balance,
		},
		{
			Path:        "/debit",
			Method:      "POST",
			HandlerFunc: controller.Debit,
		},
		{
			Path:        "/credit",
			Method:      "POST",
			HandlerFunc: controller.Credit,
		},
		{
			Path:        "/cancel",
			Method:      "POST",
			HandlerFunc: controller.Cancel,
		},
		{
			Path:        "/promo_payout",
			Method:      "POST",
			HandlerFunc: controller.PromoPayout,
		},
	}

	return &provider.Router{
		Name:     ProviderName,
		BasePath: config.BasePath,
		Routes:   routes,
		Middlewares: []fiber.Handler{
			NewAPITokenValidator(apiTokenParamName, auth.APIKey),
		},
	}, nil
}

func NewOperatorRouter(config configs.ProviderConf, httpClient rest.HTTPClientJSONInterface) (*provider.Router, error) {
	auth, err := GetAuthConf(config)
	if err != nil {
		return nil, err
	}
	evoService := EvoService{
		Auth:   auth,
		Conf:   &config,
		Client: httpClient,
	}
	glController := provider.NewGameLaunchController(&evoService)
	grCtrl := provider.NewGameRoundController(&evoService)
	routes := []provider.Route{
		{
			Path:        "/gamelaunch",
			Method:      "POST",
			HandlerFunc: glController.GameLaunchEndpoint,
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
