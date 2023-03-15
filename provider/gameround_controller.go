package provider

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

type GameRoundController struct {
	ps ProviderService
}

func NewGameRoundController(s ProviderService) *GameRoundController {
	return &GameRoundController{s}
}

// GetGameRoundEndpoint Returns status from provider service
func (ctrl *GameRoundController) GetGameRoundEndpoint(c *fiber.Ctx) error {
	gameRoundID := c.Params("gameRoundId")
	casinoID := c.Query("casinoId")
	res, err := ctrl.ps.GetGameRoundRender(c, GameRoundRenderRequest{gameRoundID, casinoID})
	if err != nil {
		hErr := &rest.HTTPError{}
		if errors.As(err, hErr) {
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
			return c.Status(hErr.Code).SendString(hErr.Message)
		}
		return c.Status(res).SendString(err.Error())
	}

	return c.SendStatus(res)
}
