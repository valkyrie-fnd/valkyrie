package provider

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/valkyrie-fnd/valkyrie/valkhttp"
)

type GameLaunchController struct {
	ps ProviderService
}

func NewGameLaunchController(s ProviderService) *GameLaunchController {
	return &GameLaunchController{s}
}

var validate = validator.New()

// GameLaunchEndpoint Execute provider gamelaunch request
func (ctrl GameLaunchController) GameLaunchEndpoint(ctx *fiber.Ctx) error {
	// Get locals from middleware
	g := &GameLaunchRequest{}
	if err := ctx.BodyParser(g); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	err := validate.Struct(g)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(validationErrorsMap(err))
	}

	h := &GameLaunchHeaders{}
	err = ctx.ReqHeaderParser(h)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	err = validate.Struct(h)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(validationErrorsMap(err))
	}

	url, err := ctrl.ps.GameLaunch(ctx, g, h)
	if err != nil {
		hErr := &valkhttp.HTTPError{}
		if errors.As(err, hErr) {
			return ctx.Status(hErr.Code).SendString(hErr.Error())
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	return ctx.JSON(GameLaunchResponse{GameURL: url})
}

// ValidatorErrors func for show validation errors for each invalid fields.
func validationErrorsMap(err error) map[string]string {
	// Define fields map.
	fields := map[string]string{}

	// Make error message for each invalid field.
	validationErrors := validator.ValidationErrors{}
	if errors.As(err, &validationErrors) {
		for _, err := range validationErrors {
			fields[err.Field()] = err.Error()
		}
	}

	return fields
}
