package pam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	input := map[string]any{
		"foo": "bar",
	}
	type config struct {
		Foo string `mapstruct:"foo"`
	}

	cfg, err := GetConfig[config](input)

	assert.NoError(t, err)
	assert.Equal(t, "bar", cfg.Foo)
}

func TestGetName(t *testing.T) {
	input := map[string]any{
		"name": "foo",
	}

	name, err := GetName(input)
	assert.NoError(t, err)
	assert.Equal(t, "foo", name)
}

func TestGetNameMissing(t *testing.T) {
	input := map[string]any{}

	_, err := GetName(input)
	assert.Error(t, err)
}

func TestGetNameInvalid(t *testing.T) {
	input := map[string]any{
		"name": 123,
	}

	_, err := GetName(input)
	assert.Error(t, err)
}
