package ops

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
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
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.NoError(t, err)

	_, err = app.Test(httptest.NewRequest(http.MethodGet, "/y", nil))
	assert.NoError(t, err)
}
