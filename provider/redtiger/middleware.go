package redtiger

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func validateAPIKey(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.GetReqHeaders()["Authorization"]
		token := strings.TrimPrefix(authHeader, "Basic ")
		if token != apiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(newRTErrorResponse("API authentication error", APIAuthError))
		}
		return c.Next()
	}
}

func declineReconToken(reconToken string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var baseReq BaseRequest
		if err := c.BodyParser(&baseReq); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(newRTErrorResponse(fmt.Sprintf("Invalid base request. err: %s", err.Error()), InvalidInput))
		}
		if baseReq.Token == reconToken {
			return c.Status(fiber.StatusUnauthorized).JSON(newRTErrorResponse("API authentication error", NotAuthorized))
		}
		return c.Next()
	}
}
