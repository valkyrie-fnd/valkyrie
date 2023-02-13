package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/application/provider/docs/generated"
)

func TestDocProviders(t *testing.T) {
	assert.NotEmpty(t, generated.SwaggerInfoprovider.ReadDoc())
}

func TestDocOperators(t *testing.T) {
	assert.NotEmpty(t, SwaggerInfoOperator.ReadDoc())
}
