package provider

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func OperatorAuthorization(apiKey string) fiber.Handler {
	if apiKey == "" {
		log.Warn().Msg("No api key configured for operator, authorization check disabled")

		return func(ctx *fiber.Ctx) error {
			return ctx.Next()
		}
	}

	return func(ctx *fiber.Ctx) error {
		authorizationValue := ctx.GetReqHeaders()["Authorization"]
		headerAPIKey := strings.TrimPrefix(authorizationValue, "Bearer ")

		if apiKey != headerAPIKey {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}

		return ctx.Next()
	}
}
