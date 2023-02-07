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

func TestGetSettlementType(t *testing.T) {
	input := map[string]any{
		"settlement_type": "mixed",
	}

	settlementType, err := GetSettlementType(input)
	assert.NoError(t, err)
	assert.Equal(t, "mixed", settlementType)
}

func TestGetSettlementTypeMissingButDefaultShouldKickIn(t *testing.T) {
	input := map[string]any{}

	settlementType, err := GetSettlementType(input)
	assert.NoError(t, err)
	assert.Equal(t, "mixed", settlementType)
}

func TestGetSettlementTypeInvalidType(t *testing.T) {
	input := map[string]any{
		"settlement_type": 123,
	}

	_, err := GetSettlementType(input)
	assert.Error(t, err)
}

func TestGetSettlementTypeInvalidValue(t *testing.T) {
	input := map[string]any{
		"settlement_type": "foo",
	}

	_, err := GetSettlementType(input)
	assert.Error(t, err)
}
