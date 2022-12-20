package evolution

import (
	"github.com/gofiber/fiber/v2"
)

func NewAPITokenValidator(apiTokenParamName, apiToken string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Query(apiTokenParamName)

		if token != apiToken {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		return c.Next()
	}
}
