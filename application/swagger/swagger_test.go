package swagger

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/application/provider/docs/generated"
	"github.com/valkyrie-fnd/valkyrie/configs"
)

type mockRouter struct {
	fiber.Router
}

func (m *mockRouter) Get(_ string, _ ...fiber.Handler) fiber.Router {
	return m
}

func TestConfigureSwagger(t *testing.T) {
	config := &configs.ValkyrieConfig{Providers: []configs.ProviderConf{
		{Name: "Evolution"},
	}}
	err := ConfigureSwagger(config, &mockRouter{}, &mockRouter{})
	assert.NoError(t, err)

	// Processed docs contain only Evolution
	assert.Contains(t, generated.SwaggerInfoprovider.SwaggerTemplate, "Evolution")
	assert.NotContains(t, generated.SwaggerInfoprovider.SwaggerTemplate, "Caleta")
	assert.NotContains(t, generated.SwaggerInfoprovider.SwaggerTemplate, "Red Tiger")
}
