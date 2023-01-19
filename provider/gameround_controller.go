package provider

import "github.com/gofiber/fiber/v2"

type GameRoundController struct {
	ps ProviderService
}

func NewGameRoundController(s ProviderService) *GameRoundController {
	return &GameRoundController{s}
}
func (ctrl *GameRoundController) GetGameRoundEndpoint(c *fiber.Ctx) error {
	gameRoundID := c.Params("gameRoundId")
	// locale := c.Query("locale")
	res, err := ctrl.ps.GetGameRound(c, gameRoundID)
	if err != nil {
		// handle error
		return err
	}
	return c.SendString(res)
}
