package evolution

import (
	"github.com/mitchellh/mapstructure"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

// AuthConf Evolution specific Auth configuration from valkyrie config file
type AuthConf struct {
	APIKey      string `mapstructure:"api_key"`
	CasinoToken string `mapstructure:"casino_token"`
	CasinoKey   string `mapstructure:"casino_key,omitempty"`
}

// GetAuthConf parse provider specific auth configuration
func GetAuthConf(c configs.ProviderConf) (AuthConf, error) {
	var auth AuthConf
	err := mapstructure.Decode(c.Auth, &auth)
	if err != nil {
		return auth, err
	}
	return auth, nil
}
