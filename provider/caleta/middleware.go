package caleta

import (
	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/provider/caleta/auth"
)

// VerifySignature middleware for verifying auth signature header
func VerifySignature(v auth.Verifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		signature := c.GetReqHeaders()["X-Auth-Signature"]
		body := c.Request().Body()
		err := v.Verify(signature, body)
		if err != nil {
			// any body works, we just want the RequestUuid
			var req WalletBalanceBody
			_ = c.BodyParser(&req)
			return c.Status(fiber.StatusOK).JSON(BalanceResponse{Status: RSERRORINVALIDSIGNATURE, RequestUuid: req.RequestUuid})
		}
		return c.Next()
	}
}
