package redtiger

import (
	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

const (
	ProviderName = "Red Tiger"
	basePath     = "/redtiger"
)

func init() {
	provider.ProviderFactory().
		Register(ProviderName, func(args provider.ProviderArgs) (*provider.Router, error) {
			service := NewService(args.Client)
			controller := NewProviderController(service)
			return NewProviderRouter(args.Config, controller)
		})
	provider.OperatorFactory().
		Register(ProviderName, func(args provider.OperatorArgs) (*provider.Router, error) {
			return NewOperatorRouter(args.Config), nil
		})
}

type Controller interface {
	Auth(c *fiber.Ctx) error
	Stake(c *fiber.Ctx) error
	Payout(c *fiber.Ctx) error
	Refund(c *fiber.Ctx) error
	PromoBuyin(c *fiber.Ctx) error
	PromoSettle(c *fiber.Ctx) error
	PromoRefund(c *fiber.Ctx) error
}

func NewProviderRouter(config configs.ProviderConf, controller Controller) (*provider.Router, error) {
	auth, err := GetAuthConf(config)
	if err != nil {
		return nil, err
	}
	routes := []provider.Route{
		{
			Path:        "/auth",
			Method:      "POST",
			HandlerFunc: controller.Auth,
		},
		{
			Path:        "/stake",
			Method:      "POST",
			HandlerFunc: controller.Stake,
			Middlewares: []fiber.Handler{declineReconToken(auth.ReconToken)},
		},
		{
			Path:        "/payout",
			Method:      "POST",
			HandlerFunc: controller.Payout,
		},
		{
			Path:        "/refund",
			Method:      "POST",
			HandlerFunc: controller.Refund,
		},
		{
			Path:        "/promo/buyin",
			Method:      "POST",
			HandlerFunc: controller.PromoBuyin,
			Middlewares: []fiber.Handler{declineReconToken(auth.ReconToken)},
		},
		{
			Path:        "/promo/settle",
			Method:      "POST",
			HandlerFunc: controller.PromoSettle,
		},
		{
			Path:        "/promo/refund",
			Method:      "POST",
			HandlerFunc: controller.PromoRefund,
		},
	}
	return &provider.Router{
		Name:     ProviderName,
		BasePath: basePath,
		Routes:   routes,
		Middlewares: []fiber.Handler{
			validateAPIKey(auth.APIKey),
		},
	}, nil
}

// NewOperatorRouter Routes operator calls to execute actions toward the provider
func NewOperatorRouter(config configs.ProviderConf) *provider.Router {
	glController := provider.NewGameLaunchController(
		GameLaunchService{
			Conf: &config,
		})
	routes := []provider.Route{
		{
			Path:        "/gamelaunch",
			Method:      "POST",
			HandlerFunc: glController.GameLaunchEndpoint,
		},
	}

	return &provider.Router{
		Name:        ProviderName,
		BasePath:    basePath,
		Routes:      routes,
		Middlewares: []fiber.Handler{},
	}
}
