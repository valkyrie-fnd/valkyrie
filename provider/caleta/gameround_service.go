package caleta

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

func (service *caletaService) GetGameRoundRender(ctx *fiber.Ctx, gameRoundID string) (string, error) {
	body := GameroundJSONRequestBody{
		Round:      &gameRoundID,
		OperatorId: service.authConfig.OperatorID,
	}
	req := &rest.HTTPRequest{
		URL:     fmt.Sprintf("%s%s", service.config.URL, "/api/game/round"),
		Headers: map[string]string{},
		Body:    body,
	}
	err := service.headerSigner.sign(body, req.Headers)
	if err != nil {
		log.Ctx(ctx.UserContext()).Error().Err(err).Msg("Failed to sign request")
		return "", rest.NewHTTPError(fiber.StatusInternalServerError, "Failed to sign request")
	}

	resp := InlineResponse200{}
	err = service.client.PostJSON(ctx.UserContext(), req, &resp)
	if err != nil {
		return "", err
	}
	if resp.Url == nil {
		return "", rest.NewHTTPError(fiber.StatusInternalServerError, "url missing from response")
	}

	return *resp.Url, nil
}
