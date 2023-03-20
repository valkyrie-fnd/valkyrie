package evolution

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/valkyrie-fnd/valkyrie/ops"

	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

type EvoService struct {
	Client rest.HTTPClient
	Conf   *configs.ProviderConf
	Auth   AuthConf
}

func (service EvoService) GameLaunch(ctx *fiber.Ctx, g *provider.GameLaunchRequest,
	h *provider.GameLaunchHeaders) (string, error) {
	configJSON, err := json.Marshal(g.LaunchConfig)
	if err != nil {
		return "", err
	}
	var config Config
	if err = json.Unmarshal(configJSON, &config); err != nil {
		return "", err
	}

	var propagatedUUID = uuid.NewString()
	if traceparent, ok := ops.GetTracingHeaders(ctx.UserContext())["traceparent"]; ok {
		// just propagate traceparent if available, not all available tracing headers which may contain sensitive info
		propagatedUUID = traceparent
	}

	req := UserAuthenticationRequest{
		UUID: propagatedUUID,
		Player: Player{
			ID:       g.PlayerID,
			Update:   true,
			Country:  g.Country,
			Language: g.Language,
			Currency: g.Currency,
			Session: Session{
				ID: h.SessionKey,
				IP: g.SessionIP,
			},
		},
		Config: config,
	}
	// Make Auth call
	authResp, err := service.makeAuthCall(ctx.UserContext(), req)
	if err != nil {
		return "", err
	}
	gameURL := fmt.Sprintf("%s%s", service.Conf.URL, authResp.Entry)
	return gameURL, nil
}

func (service EvoService) GetGameRoundRender(ctx *fiber.Ctx, req provider.GameRoundRenderRequest) (int, error) {
	renderURL := fmt.Sprintf("%s/api/render/v1/details", service.Conf.URL)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", service.Auth.CasinoKey, service.Auth.CasinoToken)))
	r := &rest.HTTPRequest{
		URL:   renderURL,
		Query: map[string]string{"gameId": req.GameRoundID},
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Basic %s", encodedAuth),
		},
	}
	var resp []byte
	err := service.Client.Get(ctx.UserContext(), &rest.PlainParser, r, &resp)
	ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	if err != nil {
		return fiber.StatusBadRequest, err
	}
	_, err = ctx.Response().BodyWriter().Write(resp)
	if err != nil {
		return fiber.StatusInternalServerError, err
	}
	return fiber.StatusOK, nil
}

func (service EvoService) makeAuthCall(ctx context.Context, request UserAuthenticationRequest) (*UserAuthenticationResponse, error) {
	authURL := fmt.Sprintf("%s/ua/v1/%s/%s", service.Conf.URL, service.Auth.CasinoKey, service.Auth.CasinoToken)
	resp := &UserAuthenticationResponse{}
	req := &rest.HTTPRequest{
		URL:  authURL,
		Body: &request,
	}

	err := service.Client.Post(ctx, &rest.JSONParser, req, resp)
	if nil != err {
		return nil, fmt.Errorf("failed calling evo auth: %w", err)
	}

	return resp, nil
}
