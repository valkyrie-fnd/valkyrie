package caleta

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

func (service *caletaService) GetGameRound(ctx *fiber.Ctx, gameRoundId string) (string, error) {
	body := GameroundJSONRequestBody{
		Round:      &gameRoundId,
		OperatorId: service.authConfig.OperatorID,
	}
	req := &rest.HTTPRequest{
		URL:     fmt.Sprintf("%s%s", service.config.URL, "/api/game/round"),
		Headers: map[string]string{},
		Body:    body,
	}
	err := service.headerSigner.sign(body, req.Headers)
	if err != nil {
		return "", err
	}

	resp := InlineResponse200{}
	service.client.PostJSON(ctx.UserContext(), req, &resp)
	return *resp.Url, fmt.Errorf("Not available")
}
