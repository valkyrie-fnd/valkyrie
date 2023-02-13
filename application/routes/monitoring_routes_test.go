package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringRoutes(t *testing.T) {
	app := fiber.New()
	MonitoringRoutes(app)

	assert.Equal(t, 4, int(app.HandlersCount()))

	req := httptest.NewRequest(fiber.MethodGet, "/monitoring/ping", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err, "ping route should work")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
