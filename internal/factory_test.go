package internal

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type args interface{}

func TestAbstractFactory_Build(t *testing.T) {
	factory := NewAbstractFactory[args, string]()
	key := "test"
	factory.Register(key, func(_ args) (string, error) {
		return "test", nil
	})

	client, err := factory.Build(key, nil)

	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestAbstractFactory_BuildMissing(t *testing.T) {
	factory := NewAbstractFactory[args, string]()
	key := "missing"

	_, err := factory.Build(key, nil)

	assert.Error(t, err)
}

func TestAbstractFactory_BuildError(t *testing.T) {
	factory := NewAbstractFactory[args, string]()
	key := "error"
	expectedError := errors.New("error")
	factory.Register(key, func(_ args) (string, error) {
		return "", expectedError
	})

	_, err := factory.Build(key, key)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}
