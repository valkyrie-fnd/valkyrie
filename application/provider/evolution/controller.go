package evolution

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/valkyrie-fnd/valkyrie/ops"
)

type ProviderController struct {
	service Service
}

type Service interface {
	Check(req CheckRequest) (*CheckResponse, error)
	Balance(req BalanceRequest) (*StandardResponse, error)
	Debit(req DebitRequest) (*StandardResponse, error)
	Credit(req CreditRequest) (*StandardResponse, error)
	Cancel(req CancelRequest) (*StandardResponse, error)
	PromoPayout(req PromoPayoutRequest) (*StandardResponse, error)
	WithContext(ctx context.Context) Service
}

var validate = validator.New()

func NewProviderController(service Service) *ProviderController {
	return &ProviderController{service: service}
}

// type constraint used to generalize the controller function
type requestType interface {
	BalanceRequest | DebitRequest | CreditRequest | CheckRequest | CancelRequest | PromoPayoutRequest
	uuid() string
}

func (r BalanceRequest) uuid() string {
	return r.UUID
}
func (r DebitRequest) uuid() string {
	return r.UUID
}
func (r CreditRequest) uuid() string {
	return r.UUID
}
func (r CheckRequest) uuid() string {
	return r.UUID
}
func (r CancelRequest) uuid() string {
	return r.UUID
}
func (r PromoPayoutRequest) uuid() string {
	return r.UUID
}

func execController[T requestType](c *fiber.Ctx, svcFunc func(req T) (any, error)) error {
	var req T

	// Parse request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(defaultErrorResponse("INVALID_PARAMETER", req.uuid()))
	}

	// Validate request
	validationErrors := validate.Struct(req)
	if validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(defaultErrorResponse("INVALID_PARAMETER", req.uuid()))
	}

	// Add request ID to logs and span
	ops.AddLoggingContext(c, "evolution.uuid", req.uuid())
	trace.SpanFromContext(c.UserContext()).SetAttributes(attribute.String("evolution.uuid", req.uuid()))

	// Call service
	resp, err := svcFunc(req)

	// If error, unwrap and respond accordingly
	if err != nil {
		log.Ctx(c.UserContext()).Error().Err(err).Send()
		e := unwrap(err)
		if e.response != nil {
			return c.Status(e.httpStatus).JSON(e.response)
		} else {
			return c.Status(e.httpStatus).JSON(defaultErrorResponse(e.message, req.uuid()))
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (controller ProviderController) Check(c *fiber.Ctx) error {
	return execController(c, func(r CheckRequest) (any, error) {
		return controller.service.WithContext(c.UserContext()).Check(r)
	})
}

func (controller ProviderController) Balance(c *fiber.Ctx) error {
	return execController(c, func(r BalanceRequest) (any, error) {
		return controller.service.WithContext(c.UserContext()).Balance(r)
	})
}

func (controller ProviderController) Debit(c *fiber.Ctx) error {
	return execController(c, func(r DebitRequest) (any, error) {
		return controller.service.WithContext(c.UserContext()).Debit(r)
	})
}

func (controller ProviderController) Credit(c *fiber.Ctx) error {
	return execController(c, func(r CreditRequest) (any, error) {
		return controller.service.WithContext(c.UserContext()).Credit(r)
	})
}

func (controller ProviderController) Cancel(c *fiber.Ctx) error {
	return execController(c, func(r CancelRequest) (any, error) {
		return controller.service.WithContext(c.UserContext()).Cancel(r)
	})
}

func (controller ProviderController) PromoPayout(c *fiber.Ctx) error {
	return execController(c, func(r PromoPayoutRequest) (any, error) {
		return controller.service.WithContext(c.UserContext()).PromoPayout(r)
	})
}
