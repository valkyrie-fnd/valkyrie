package provider

import "github.com/gofiber/fiber/v2"

type Router struct {
	Name        string
	BasePath    string
	Routes      []Route
	Middlewares []fiber.Handler
}

type Route struct {
	Path        string
	Method      string
	HandlerFunc fiber.Handler
	Middlewares []fiber.Handler
}

type GameLaunchHeaders struct {
	SessionKey string `reqHeader:"X-Player-Token" validate:"required"`
}

type GameLaunchRequest struct {
	LaunchConfig   map[string]interface{} `json:"launchConfig,omitempty"`
	Currency       string                 `json:"currency" validate:"required"`
	ProviderGameID string                 `json:"providerGameId" validate:"required"`
	PlayerID       string                 `json:"playerId" validate:"required"`
	Casino         string                 `json:"casino,omitempty"`
	Country        string                 `json:"country,omitempty"`
	Language       string                 `json:"language,omitempty"`
	SessionIP      string                 `json:"sessionIp,omitempty"`
}

type GameLaunchResponse struct {
	GameURL string `json:"gameUrl"`
}

type GameRoundRenderRequest struct {
	GameRoundID string
	CasinoID    string
}

// ProviderService Contains Provider exposed features
type ProviderService interface {
	// GameLaunch returns url to game session
	GameLaunch(*fiber.Ctx, *GameLaunchRequest, *GameLaunchHeaders) (string, error)
	// GetGameRoundRender Returns status code to return. Fiber context should be populated with the response
	GetGameRoundRender(*fiber.Ctx, GameRoundRenderRequest) (int, error)
}
