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

// getAuthConf parse provider specific auth configuration
func getAuthConf(c configs.ProviderConf) (AuthConf, error) {
	var auth AuthConf
	err := mapstructure.Decode(c.Auth, &auth)
	if err != nil {
		return auth, err
	}
	return auth, nil
}
