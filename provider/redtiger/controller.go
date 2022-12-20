package redtiger

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ProviderController struct {
	service Service
}

type Service interface {
	Auth(req AuthRequest) (*AuthResponseWrapper, *ErrorResponse)
	Stake(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse)
	Payout(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse)
	Refund(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse)
	PromoBuyin(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse)
	PromoSettle(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse)
	PromoRefund(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse)
	WithContext(ctx context.Context) Service
}

func NewProviderController(service Service) *ProviderController {
	return &ProviderController{service: service}
}

type requestType interface {
	AuthRequest | StakeRequest | PayoutRequest | RefundRequest | BaseRequest
}
type responseType interface {
	AuthResponseWrapper | StakeResponseWrapper | PayoutResponseWrapper | RefundResponseWrapper
}

func execController[T requestType, R responseType](c *fiber.Ctx, svcFunc func(req T) (*R, *ErrorResponse)) error {
	var req T
	// Parse request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newRTErrorResponse(fmt.Sprintf("Invalid input. err: %s", err.Error()), InvalidInput))
	}

	// Validate request
	validationErrors := validate.Struct(req)
	if validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newRTErrorResponse(fmt.Sprintf("Invalid input. err: %s", validationErrors.Error()), InvalidInput))
	}

	// Call service
	resp, err := svcFunc(req)

	// If error, return it
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// Auth handler function
func (ctrl *ProviderController) Auth(c *fiber.Ctx) error {
	return execController(c, func(req AuthRequest) (*AuthResponseWrapper, *ErrorResponse) {
		return ctrl.service.WithContext(c.UserContext()).Auth(req)
	})
}

// Payout handler function
func (ctrl *ProviderController) Payout(c *fiber.Ctx) error {
	return execController(c, func(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
		return ctrl.service.WithContext(c.UserContext()).Payout(req)
	})
}

// Refund handler function
func (ctrl *ProviderController) Refund(c *fiber.Ctx) error {
	return execController(c, func(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
		return ctrl.service.WithContext(c.UserContext()).Refund(req)
	})
}

// Stake handler function
func (ctrl *ProviderController) Stake(c *fiber.Ctx) error {
	return execController(c, func(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
		return ctrl.service.WithContext(c.UserContext()).Stake(req)
	})
}

// PromoBuyin handler function
func (ctrl *ProviderController) PromoBuyin(c *fiber.Ctx) error {
	return execController(c, func(req StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
		return ctrl.service.WithContext(c.UserContext()).PromoBuyin(req)
	})
}

// PromoRefund handler function
func (ctrl *ProviderController) PromoRefund(c *fiber.Ctx) error {
	return execController(c, func(req RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
		return ctrl.service.WithContext(c.UserContext()).PromoRefund(req)
	})
}

// PromoSettle handler function
func (ctrl *ProviderController) PromoSettle(c *fiber.Ctx) error {
	return execController(c, func(req PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
		return ctrl.service.WithContext(c.UserContext()).PromoSettle(req)
	})
}
