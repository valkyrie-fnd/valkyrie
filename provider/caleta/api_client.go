package caleta

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider/caleta/auth"
	"github.com/valkyrie-fnd/valkyrie/valkhttp"
)

type API interface {
	requestGameLaunch(ctx context.Context, body GameUrlBody) (*InlineResponse200, error)
	getGameRoundRender(ctx context.Context, gameRoundID, casinoID string) (*gameRoundRenderResponse, error)
	getRoundTransactions(ctx context.Context, gameRoundID string) (*transactionResponse, error)
}

type apiClient struct {
	rest         valkhttp.HTTPClient
	headerSigner headerSigner
	authConfig   AuthConf
	url          string
	operatorID   string
}

func NewAPIClient(client valkhttp.HTTPClient, config configs.ProviderConf) (*apiClient, error) {
	authConfig, err := getAuthConf(config)
	if err != nil {
		return nil, err
	}

	hs, err := newHeaderSigner(authConfig)
	if err != nil {
		return nil, err
	}

	return &apiClient{
		rest:         client,
		authConfig:   authConfig,
		headerSigner: hs,
		operatorID:   authConfig.OperatorID,
		url:          config.URL,
	}, nil
}

type headerSigner interface {
	sign(body any, headers map[string]string) error
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

func (s *authHeaderSigner) sign(body any, headers map[string]string) error {
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

func (_ *noopHeaderSigner) sign(_ any, _ map[string]string) error {
	return nil
}

func (apiClient *apiClient) requestGameLaunch(ctx context.Context, body GameUrlBody) (*InlineResponse200, error) {
	req := &valkhttp.HTTPRequest{
		URL:     fmt.Sprintf("%s%s", apiClient.url, "/api/game/url"),
		Headers: map[string]string{},
		Body:    body,
	}

	err := apiClient.headerSigner.sign(body, req.Headers)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to sign request")
		return nil, valkhttp.NewHTTPError(fiber.StatusInternalServerError, "Failed to sign request")
	}

	resp := InlineResponse200{}
	err = apiClient.rest.Post(ctx, &valkhttp.JSONParser, req, &resp)

	return &resp, err
}

func (apiClient *apiClient) getGameRoundRender(ctx context.Context, gameRoundID, casinoID string) (*gameRoundRenderResponse, error) {
	body := GameroundJSONRequestBody{
		Round:        &gameRoundID,
		OperatorId:   apiClient.operatorID,
		SubPartnerId: "default",
	}
	if casinoID != "" {
		body.SubPartnerId = casinoID
	}
	req := &valkhttp.HTTPRequest{
		URL:     fmt.Sprintf("%s%s", apiClient.url, "/api/game/round"),
		Headers: map[string]string{},
		Body:    body,
	}
	err := apiClient.headerSigner.sign(body, req.Headers)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to sign request")
		return nil, valkhttp.NewHTTPError(fiber.StatusInternalServerError, "Failed to sign request")
	}

	resp := gameRoundRenderResponse{}
	err = apiClient.rest.Post(ctx, &valkhttp.JSONParser, req, &resp)
	return &resp, err
}

func (apiClient *apiClient) getRoundTransactions(ctx context.Context, gameRoundID string) (*transactionResponse, error) {
	req := &valkhttp.HTTPRequest{
		Body: transactionRequestBody{
			RoundID:    gameRoundID,
			OperatorID: apiClient.operatorID,
		},
		URL:     fmt.Sprintf("%s%s", apiClient.url, "/api/transactions/round"),
		Headers: map[string]string{},
	}

	resp := transactionResponse{}

	err := apiClient.headerSigner.sign(req.Body, req.Headers)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to sign request")
		return nil, valkhttp.NewHTTPError(fiber.StatusInternalServerError, "Failed to sign request")
	}

	err = apiClient.rest.Post(ctx, &valkhttp.JSONParser, req, &resp)
	return &resp, err
}
