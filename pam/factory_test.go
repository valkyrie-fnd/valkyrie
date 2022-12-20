package pam

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

type mockPamClient struct {
	PamClient
}

func TestGetPamClient(t *testing.T) {
	ClientFactory().Register("foo", func(_ ClientArgs) (PamClient, error) {
		return &mockPamClient{}, nil
	})

	pamClient, err := GetPamClient(ClientArgs{
		Context: nil,
		Client:  nil,
		Config: configs.PamConf{
			"name": "foo",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, &mockPamClient{}, pamClient)
}

func TestGetPamClientMissing(t *testing.T) {
	_, err := GetPamClient(ClientArgs{
		Context: nil,
		Client:  nil,
		Config: configs.PamConf{
			"name": "bar",
		},
	})
	assert.Error(t, err)
}

func TestGetPamClientInvalidArgs(t *testing.T) {
	_, err := GetPamClient(ClientArgs{})
	assert.Error(t, err)
}
