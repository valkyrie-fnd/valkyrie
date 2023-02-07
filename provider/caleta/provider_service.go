package caleta

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/go-querystring/query"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

type caletaService struct {
	apiClient      API
	providerConfig configs.ProviderConf
	caletaConfig   caletaConf
	authConfig     AuthConf
}

func NewCaletaService(apiClient API, config configs.ProviderConf) (*caletaService, error) {
	authConfig, err := getAuthConf(config)
	if err != nil {
		return nil, err
	}

	caletaConfig, err := getCaletaConf(config)
	if err != nil {
		return nil, err
	}

	return &caletaService{apiClient: apiClient, providerConfig: config, caletaConfig: caletaConfig, authConfig: authConfig}, nil
}

// GetGameRoundRender returns a game render for a given game round
func (service *caletaService) GetGameRoundRender(ctx *fiber.Ctx, gameRoundID string) (string, error) {
	resp, err := service.apiClient.getGameRoundRender(ctx.UserContext(), gameRoundID)
	if err != nil {
		return "", err
	}
	if resp.Url == nil {
		return "", rest.NewHTTPError(fiber.StatusBadRequest, fmt.Sprintf("%d: %s", resp.Code, resp.Message))
	}

	return *resp.Url, nil
}

// GameLaunch launches games
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

	return fmt.Sprintf("%s/open_game?%s", service.providerConfig.URL, values.Encode()), nil
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

	resp, err := service.apiClient.requestGameLaunch(ctx.UserContext(), *body)
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
