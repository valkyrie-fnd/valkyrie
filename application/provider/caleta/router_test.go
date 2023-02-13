package caleta

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

func Test_getProviderMiddlewares(t *testing.T) {
	config := configs.ProviderConf{
		Name: ProviderName,
		Auth: map[string]any{
			"verification_key": testingPublicKey,
		},
	}

	auth, _ := getAuthConf(config)
	middlewares, err := getProviderMiddlewares(auth)
	assert.NoError(t, err)
	assert.NotEmpty(t, middlewares)
}
