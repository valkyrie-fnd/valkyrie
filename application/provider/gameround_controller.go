package provider

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/httpclient"
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
	res, err := ctrl.ps.GetGameRoundRender(c, gameRoundID)
	if err != nil {
		herr := &httpclient.HTTPError{}
		if errors.As(err, herr) {
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
			return c.Status(herr.Code).SendString(herr.Message)
		}
		return err
	}

	c.Response().Header.Add("Location", res)
	return c.SendStatus(fiber.StatusFound)
}
