package redtiger

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/go-querystring/query"
	"github.com/mitchellh/mapstructure"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

type RedTigerService struct {
	Conf *configs.ProviderConf
}

var validate = validator.New()

func (service RedTigerService) GameLaunch(_ *fiber.Ctx, g *provider.GameLaunchRequest,
	h *provider.GameLaunchHeaders) (string, error) {
	if h.SessionKey == "" {
		return "", fmt.Errorf("Missing SessionKey")
	}
	glr := &GameLaunchRequest{
		Token:    h.SessionKey,
		Currency: g.Currency,
		UserID:   g.PlayerID,
		Casino:   g.Casino,
	}
	launchConfig, err := getLaunchConfig(g.LaunchConfig)
	if err != nil {
		return "", err
	}
	launchConfQuery, err := query.Values(launchConfig)
	if err != nil {
		return "", err
	}
	gameLaunchReqQuery, err := query.Values(glr)
	if err != nil {
		return "", err
	}
	// Generate Gamelaunch url
	url := fmt.Sprintf(
		"%s/%s?%s&%s",
		service.Conf.URL,
		g.ProviderGameID,
		gameLaunchReqQuery.Encode(),
		launchConfQuery.Encode())
	return url, nil
}
func (service RedTigerService) GetGameRoundRender(*fiber.Ctx, provider.GameRoundRenderRequest) (int, error) {
	return 404, fmt.Errorf("Not available")
}
func getLaunchConfig(conf map[string]interface{}) (*rtGameLaunchConfig, error) {
	var launchConfig rtGameLaunchConfig
	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &launchConfig,
		TagName:  "url",
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	_ = decoder.Decode(conf)
	err := validate.Struct(launchConfig)
	if err != nil {
		return nil, err
	}
	return &launchConfig, nil
}
