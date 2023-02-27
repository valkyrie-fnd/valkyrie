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

// GetGameRoundEndpoint Returns redirect status with provider url for game round rendering
func (ctrl *GameRoundController) GetGameRoundEndpoint(c *fiber.Ctx) error {
	gameRoundID := c.Params("gameRoundId")
	casinoID := c.Query("casinoId")
	res, err := ctrl.ps.GetGameRoundRender(c, GameRoundRenderRequest{gameRoundID, casinoID})
	if err != nil {
		herr := &rest.HTTPError{}
		if errors.As(err, herr) {
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
			return c.Status(herr.Code).SendString(herr.Message)
		}
		return err
	}

	c.Response().Header.Add("Location", res)
	return c.SendStatus(fiber.StatusFound)
}
