package example

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

// exampleProviderService implements provider.ProviderService
// Connect to the provider specifics for how to launch a game or communicate with other provider apis
type exampleProviderService struct {
	conf       *configs.ProviderConf
	httpClient rest.HTTPClient
}

func NewExampleProviderService(c configs.ProviderConf, httpClient rest.HTTPClient) *exampleProviderService {
	return &exampleProviderService{conf: &c, httpClient: httpClient}
}

// GameLaunch implements provider.ProviderService
// Some provider game launch requests are simply to build a url, while others require some communication with provider backend.
func (s *exampleProviderService) GameLaunch(c *fiber.Ctx, r *provider.GameLaunchRequest, h *provider.GameLaunchHeaders) (string, error) {
	// Could return a "static" url based on config and the request.
	// Or it could be an endpoint where the game provider returns a url for the operator
	return fmt.Sprintf("%s/gamelaunch?gameId=%s&playerId=%s", s.conf.URL, r.ProviderGameID, r.PlayerID), nil
}

// GetGameRoundRender implements provider.ProviderService
// It should return a status and update fiber.Ctx appropriately.
// It can redirect to a separate url or return the rendered html by itself
func (s *exampleProviderService) GetGameRoundRender(c *fiber.Ctx, renderReq provider.GameRoundRenderRequest) (int, error) {
	url := fmt.Sprintf("%s/gameround/render?roundId=%s", s.conf.URL, renderReq.GameRoundID)
	c.Set("Location", url)
	return fiber.StatusFound, nil
}
