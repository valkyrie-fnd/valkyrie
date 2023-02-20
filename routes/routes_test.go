package routes

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

// Verify that all providers are registered, which is easily forgotten
// since it's done implicitly in the init() function of the provider package.
func Test_AllProvidersGetsRegistered(t *testing.T) {

	fs, err := os.ReadDir("../provider")
	require.NoError(t, err)

	for _, f := range fs {
		if f.IsDir() && f.Name() != "internal" && f.Name() != "docs" {

			r, err := provider.ProviderFactory().Build(f.Name(), provider.ProviderArgs{
				PamClient: &dummyPamClient{},
			})
			require.NoError(t, err)
			assert.NotNil(t, r)
		}
	}
}

type dummyPamClient struct {
	pam.PamClient
}

func (d dummyPamClient) GetTransactionSupplier() pam.TransactionSupplier {
	return pam.OPERATOR
}
