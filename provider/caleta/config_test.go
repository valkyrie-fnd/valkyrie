package caleta

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

func Test_caletaConf_static(t *testing.T) {
	c := configs.ProviderConf{
		ProviderSpecific: map[string]any{
			"game_launch_type": "static",
		},
	}
	res, _ := getCaletaConf(c)
	assert.Equal(t, res.GameLaunchType, Static)
}

func Test_caletaConf_request(t *testing.T) {
	c := configs.ProviderConf{
		ProviderSpecific: map[string]any{
			"game_launch_type": "request",
		},
	}
	res, _ := getCaletaConf(c)
	assert.Equal(t, res.GameLaunchType, Request)
}

func Test_caletaConf_empty_defaults_static(t *testing.T) {
	c := configs.ProviderConf{}
	res, _ := getCaletaConf(c)
	assert.Equal(t, res.GameLaunchType, Static)
}

func Test_caletaConf_unknown_defaults_static(t *testing.T) {
	c := configs.ProviderConf{
		ProviderSpecific: map[string]any{
			"game_launch_type": "strangeValue",
		},
	}
	res, _ := getCaletaConf(c)
	assert.Equal(t, res.GameLaunchType, Static)
}
