package caleta

import (
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"

	"github.com/valkyrie-fnd/valkyrie/rest"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"

	"github.com/valkyrie-fnd/valkyrie/provider"
)

type gameURLQuery struct {
	Country      Country      `url:"country"`
	Currency     Currency     `url:"currency"`
	DepositURL   string       `url:"deposit_url,omitempty"`
	GameCode     GameCode     `url:"game_code"`
	Lang         Language     `url:"lang"`
	LobbyURL     string       `url:"lobby_url,omitempty"`
	OperatorID   OperatorId   `url:"operator_id"`
	SubPartnerID SubPartnerId `url:"sub_partner_id,omitempty"`
	Token        Token        `url:"token,omitempty"`
	User         User         `url:"user,omitempty"`
}

func (service *caletaService) GameLaunch(ctx *fiber.Ctx, g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (string, error) {
	switch service.caletaConfig.GameLaunchType {
	case Static:
		return service.staticGameLaunch(ctx, g, h)
	case Request:
		return service.requestingGameLaunch(ctx, g, h)
	default:
		return "", fmt.Errorf("Invalid Gamelaunch type: %s", service.caletaConfig.GameLaunchType)
	}
}

func (service *caletaService) staticGameLaunch(_ *fiber.Ctx, g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (string, error) {
	q, err := service.getGameURLQuery(g, h)
	if err != nil {
		return "", err
	}

	values, err := query.Values(q)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/open_game?%s", service.config.URL, values.Encode()), nil
}

func (service *caletaService) getGameURLQuery(g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (*gameURLQuery, error) {
	launchConfig, err := getLaunchConfig(g.LaunchConfig)
	if err != nil {
		return nil, err
	}

	return &gameURLQuery{
		Country:      Country(g.Country),
		Currency:     Currency(g.Currency),
		DepositURL:   utils.OrZeroValue(launchConfig.DepositURL),
		GameCode:     g.ProviderGameID,
		Lang:         Language(g.Language),
		LobbyURL:     launchConfig.LobbyURL,
		OperatorID:   service.authConfig.OperatorID,
		SubPartnerID: launchConfig.SubPartnerID,
		Token:        h.SessionKey,
		User:         g.PlayerID,
	}, nil
}

func (service *caletaService) requestingGameLaunch(ctx *fiber.Ctx, g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (string, error) {
	body, err := service.getGameLaunchBody(g, h)
	if err != nil {
		return "", err
	}

	req := &rest.HTTPRequest{
		URL:     fmt.Sprintf("%s%s", service.config.URL, "/api/game/url"),
		Headers: map[string]string{},
		Body:    body,
	}
	resp := InlineResponse200{}

	err = service.headerSigner.sign(body, req.Headers)
	if err != nil {
		return "", err
	}

	err = service.client.PostJSON(ctx.UserContext(), req, &resp)
	if err != nil {
		return "", err
	}
	if resp.Url == nil {
		return "", fmt.Errorf("url missing from response")
	}

	return *resp.Url, nil
}

func (service *caletaService) getGameLaunchBody(g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (*GameUrlBody, error) {
	launchConfig, err := getLaunchConfig(g.LaunchConfig)
	if err != nil {
		return nil, err
	}

	return &GameUrlBody{
		Country:      Country(g.Country),
		Currency:     Currency(g.Currency),
		DepositUrl:   launchConfig.DepositURL,
		GameCode:     g.ProviderGameID,
		Lang:         Language(g.Language),
		LobbyUrl:     launchConfig.LobbyURL,
		OperatorId:   service.authConfig.OperatorID,
		SubPartnerId: launchConfig.SubPartnerID,
		Token:        &h.SessionKey,
		User:         &g.PlayerID,
	}, nil
}

type caletaLaunchConfig struct {
	DepositURL   *string `mapstructure:"deposit_url,omitempty"`
	LobbyURL     string  `mapstructure:"lobby_url,omitempty"`
	SubPartnerID string  `mapstructure:"sub_partner_id,omitempty"`
}

func getLaunchConfig(input map[string]interface{}) (*caletaLaunchConfig, error) {
	config := caletaLaunchConfig{}
	err := mapstructure.Decode(input, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
