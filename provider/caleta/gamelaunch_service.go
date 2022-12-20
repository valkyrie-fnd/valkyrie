package caleta

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/google/go-querystring/query"
	"github.com/rs/zerolog/log"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"

	"github.com/valkyrie-fnd/valkyrie/provider/caleta/auth"
	"github.com/valkyrie-fnd/valkyrie/rest"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

func NewStaticURLGameLaunchService(config configs.ProviderConf) (*staticURLGameLaunchService, error) {
	authConfig, err := getAuthConf(config)
	if err != nil {
		return nil, err
	}

	return &staticURLGameLaunchService{
		config:     config,
		authConfig: authConfig,
	}, nil
}

// staticURLGameLaunchService implements the provider.GameLaunchService using Caleta API endpoint GET "open_game".
// by simply building the expected game url.
type staticURLGameLaunchService struct {
	config     configs.ProviderConf
	authConfig AuthConf
}

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

func (service *staticURLGameLaunchService) GameLaunch(_ *fiber.Ctx, g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (string, error) {
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

func (service *staticURLGameLaunchService) getGameURLQuery(g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (*gameURLQuery, error) {
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

func NewRequestingGameLaunchService(config configs.ProviderConf, client rest.HTTPClientJSONInterface) (*requestingGameLaunchService, error) {
	authConfig, err := getAuthConf(config)
	if err != nil {
		return nil, err
	}

	hs, err := newHeaderSigner(authConfig)
	if err != nil {
		return nil, err
	}

	return &requestingGameLaunchService{
		config:       config,
		client:       client,
		authConfig:   authConfig,
		headerSigner: hs,
	}, nil
}

// requestingGameLaunchService implements the provider.GameLaunchService using Caleta Games API endpoint POST "/api/game/url"
type requestingGameLaunchService struct {
	client       rest.HTTPClientJSONInterface
	headerSigner headerSigner
	authConfig   AuthConf
	config       configs.ProviderConf
}

func (service *requestingGameLaunchService) GameLaunch(ctx *fiber.Ctx, g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (string, error) {
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

func (service *requestingGameLaunchService) getGameLaunchBody(g *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (*GameUrlBody, error) {
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

type headerSigner interface {
	sign(body *GameUrlBody, headers map[string]string) error
}

func newHeaderSigner(authConfig AuthConf) (headerSigner, error) {
	if authConfig.SigningKey != "" {
		sig, err := NewSigner([]byte(authConfig.SigningKey))
		if err != nil {
			return nil, err
		}

		return &authHeaderSigner{
			signer: sig,
		}, nil
	} else {
		log.Warn().Msg("Missing Caleta provider 'signing_key' config, skipping header signing")
		return &noopHeaderSigner{}, nil
	}
}

type authHeaderSigner struct {
	signer auth.Signer
}

func (s *authHeaderSigner) sign(body *GameUrlBody, headers map[string]string) error {
	byteBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	signature, err := s.signer.Sign(byteBody)
	if err != nil {
		return err
	}

	headers["X-Auth-Signature"] = string(signature)

	return nil
}

type noopHeaderSigner struct{}

func (_ *noopHeaderSigner) sign(_ *GameUrlBody, _ map[string]string) error {
	return nil
}
