package ops

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func Test_pathSkippingMiddleware(t *testing.T) {

	failing := func(c *fiber.Ctx) error {
		return assert.AnError
	}

	app := fiber.New()
	app.Use(filterPath("x", failing))
	app.Get("/*", func(c *fiber.Ctx) error { return c.SendString("Hello, World 👋!") })

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/x", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	_, err = app.Test(httptest.NewRequest(http.MethodGet, "/y", nil))
	assert.NoError(t, err)
}

func Test_GoogleErrorHook_Global(t *testing.T) {
	globalLogger := log.Logger
	defer func() {
		log.Logger = globalLogger
	}()

	// Add googleErrorHook to global logger
	captor := strings.Builder{}
	log.Logger = log.Output(&captor).Hook(googleErrorHook{})

	log.Error().Msg("oops")

	logMessage := captor.String()
	assert.Contains(t, logMessage, "oops")
	assert.Contains(t, logMessage, "@type")
	assert.Contains(t, logMessage, "ReportedErrorEvent")
}

func Test_GoogleErrorHook_Inherited(t *testing.T) {
	globalLogger := log.Logger
	defer func() {
		log.Logger = globalLogger
	}()

	// Add googleErrorHook to global logger
	captor := strings.Builder{}
	log.Logger = log.Output(&captor).Hook(googleErrorHook{})

	childLogger := log.With().Logger()
	childLogger.Error().Msg("oops")

	logMessage := captor.String()
	assert.Contains(t, logMessage, "oops")
	assert.Contains(t, logMessage, "@type")
	assert.Contains(t, logMessage, "ReportedErrorEvent")
}
