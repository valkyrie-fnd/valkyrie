package routes

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/application/provider"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/httpclient"
	"github.com/valkyrie-fnd/valkyrie/pam"

	// init providers
	_ "github.com/valkyrie-fnd/valkyrie/application/provider/caleta"
	_ "github.com/valkyrie-fnd/valkyrie/application/provider/evolution"
	_ "github.com/valkyrie-fnd/valkyrie/application/provider/redtiger"
)

// ProviderRoutes Init the provider routes
func ProviderRoutes(a *fiber.App, config *configs.ValkyrieConfig, pam pam.PamClient, httpClient httpclient.HTTPClientJSONInterface) error {
	// ping endpoint is public and used by load balancers for health checking
	a.Get("/ping", func(_ *fiber.Ctx) error { return nil })

	// Create providers subgroup and registry
	registry := provider.NewRegistry(a, config.ProviderBasePath)

	// Register all configured providers
	for _, c := range config.Providers {
		providerRouter, err := provider.ProviderFactory().
			Build(c.Name, provider.ProviderArgs{
				Config:     c,
				PamClient:  pam,
				HTTPClient: httpClient,
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
func OperatorRoutes(a *fiber.App, config *configs.ValkyrieConfig, httpClient httpclient.HTTPClientJSONInterface) error {
	// ping endpoint is public and used by load balancers for health checking
	a.Get("/ping", func(_ *fiber.Ctx) error { return nil })

	// Add authorization for operator paths
	a.Use(config.OperatorBasePath, provider.OperatorAuthorization(config.OperatorAPIKey))

	// Create subgroup and registry
	registry := provider.NewRegistry(a, config.OperatorBasePath)

	// Register all configured providers
	for _, c := range config.Providers {
		operatorRouter, err := provider.OperatorFactory().
			Build(c.Name, provider.OperatorArgs{
				Config:     c,
				HTTPClient: httpClient,
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
