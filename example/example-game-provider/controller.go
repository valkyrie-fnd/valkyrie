package example

import (
	"github.com/gofiber/fiber/v2"
)

// Controller use for setting up the wallet endpoint functions. Here request and header validation can be done.
// If you're using oapi-codegen, this part can be autogenerated.
// This is just an example of how the controller could be set up using generics.
type Controller struct {
	walletService *WalletService
}

func NewController(walletService *WalletService) *Controller {
	return &Controller{walletService: walletService}
}

type requestType interface {
	balanceRequest // | BetRequest etc. Add all request types the provider supports
}

// execController generic function parsing and handling all request parsing and validation that is the same for all wallet requests.
func execController[T requestType](c *fiber.Ctx, svcFunc func(req T) (any, error)) error {
	var req T

	// Parse request. In this example it is using bodyParser. If the wallet pass data in some other way that can be dealt with here.
	if err := c.BodyParser(&req); err != nil {
		// Handle errors as expected by the game provider wallet api.
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Call the walletService
	resp, err := svcFunc(req)
	// If error, return it
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (c *Controller) WalletBalanceEndpoint(ctx *fiber.Ctx) error {
	return execController(ctx, func(r balanceRequest) (any, error) {
		// walletService handles the mapping toward the Valkyrie pam api, pam.PamClient.
		balance := c.walletService.GetBalance(r)
		// return what provider wallet api needs
		return balance, nil
	})
}
