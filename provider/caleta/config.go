package caleta

import (
	"github.com/mitchellh/mapstructure"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

type AuthConf struct {
	// OperatorId configured in Caletas system
	OperatorID string `mapstructure:"operator_id"`
	// SigningKey Used to sign outgoing requests
	SigningKey string `mapstructure:"signing_key"`
	// VerificationKey Used to verify incoming requests
	VerificationKey string `mapstructure:"verification_key"`
}
type GameLaunchType string

const (
	Static  GameLaunchType = "static"
	Request GameLaunchType = "request"
)

type caletaConf struct {
	GameLaunchType GameLaunchType `mapstructure:"game_launch_type"`
}

// getAuthConf parse provider specific auth configuration
func getAuthConf(c configs.ProviderConf) (AuthConf, error) {
	var auth AuthConf
	err := mapstructure.Decode(c.Auth, &auth)
	if err != nil {
		return auth, err
	}
	return auth, nil
}

// getCaletaConf parse provide specific configuration
func getCaletaConf(c configs.ProviderConf) (caletaConf, error) {
	var cc caletaConf
	if c.ProviderSpecific != nil {
		err := mapstructure.Decode(c.ProviderSpecific, &cc)
		if err != nil {
			return cc, err
		}
		if cc.GameLaunchType != Static && cc.GameLaunchType != Request {
			// Default to Static
			cc.GameLaunchType = Static
		}
	} else {
		// Default to Static
		cc = caletaConf{
			GameLaunchType: Static,
		}
	}
	return cc, nil
}
