package routes

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/rest"

	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/provider"

	// init providers
	_ "github.com/valkyrie-fnd/valkyrie/provider/caleta"
	_ "github.com/valkyrie-fnd/valkyrie/provider/evolution"
	_ "github.com/valkyrie-fnd/valkyrie/provider/redtiger"
)

// ProviderRoutes Init the provider routes
func ProviderRoutes(a *fiber.App, config *configs.ValkyrieConfig, pam pam.PamClient) error {
	// ping endpoint is public and used by load balancers for health checking
	a.Get("/ping", func(_ *fiber.Ctx) error { return nil })

	// Create providers subgroup and registry
	registry := provider.NewRegistry(a, config.ProviderBasePath)

	// Register all configured providers
	for _, c := range config.Providers {
		providerRouter, err := provider.ProviderFactory().
			Build(c.Name, provider.ProviderArgs{
				Config: c,
				Client: pam,
			})
		if err != nil {
			return fmt.Errorf("implementation of provider '%s' does not exist (%w)", c.Name, err)
		}
		log.Info().Msgf("Registering %s provider routes", providerRouter.Name)
		if err := registry.Register(providerRouter); err != nil {
			return err
		}
	}

	return nil
}

// OperatorRoutes Init the operator side routes
func OperatorRoutes(a *fiber.App, config *configs.ValkyrieConfig, client rest.HTTPClientJSONInterface) error {
	// ping endpoint is public and used by load balancers for health checking
	a.Get("/ping", func(_ *fiber.Ctx) error { return nil })

	// Create subgroup and registry
	registry := provider.NewRegistry(a, config.OperatorBasePath)

	// Register all configured providers
	for _, c := range config.Providers {
		operatorRouter, err := provider.OperatorFactory().
			Build(c.Name, provider.OperatorArgs{
				Config: c,
				Client: client,
			})
		if err != nil {
			return fmt.Errorf("implementation of operator routes for provider '%s' does not exist (%w)", c.Name, err)
		}
		log.Info().Msgf("Registering %s operator routes", operatorRouter.Name)
		if err := registry.Register(operatorRouter); err != nil {
			return err
		}
	}

	return nil
}
