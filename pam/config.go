package pam

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/valkyrie-fnd/valkyrie/configs"
)

const defaultSettlementType = "mixed"

func GetConfig[T any](c configs.PamConf) (*T, error) {
	var config T
	err := mapstructure.Decode(c, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func GetName(c configs.PamConf) (string, error) {
	val, found := c["name"]
	if !found {
		return "", fmt.Errorf("required pam field \"name\" not found")
	}
	name, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("pam field \"name\" has unknown type %v", val)
	}
	return name, nil
}

func GetSettlementType(c configs.PamConf) (string, error) {
	if val, found := c["settlement_type"]; found {
		settlementType, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("pam field \"settlement_type\" has unknown type %v", val)
		}

		switch settlementType {
		case "gamewise", "mixed":
			return settlementType, nil
		default:
			return "", fmt.Errorf("pam field \"settlement_type\" has invalid value %s", settlementType)
		}
	} else {
		return defaultSettlementType, nil
	}
}
