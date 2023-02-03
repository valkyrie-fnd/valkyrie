package provider

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestOperatorAuthorization(t *testing.T) {
	basePath := "/op"
	url := basePath + "/test"

	tests := []struct {
		name   string
		apiKey string
		bearer string
		status int
	}{
		{
			name:   "authorization disabled missing api key",
			status: 200,
		},
		{
			name:   "authorization enabled missing bearer",
			apiKey: "key",
			status: 401,
		},
		{
			name:   "authorization enabled mismatch bearer",
			apiKey: "key",
			bearer: "yek",
			status: 401,
		},
		{
			name:   "authorization enabled authorized request",
			apiKey: "key",
			bearer: "key",
			status: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(basePath, OperatorAuthorization(tt.apiKey))
			app.All(url, func(ctx *fiber.Ctx) error {
				return ctx.SendStatus(200)
			})

			req := httptest.NewRequest(fiber.MethodGet, url, nil)
			if tt.bearer != "" {
				req.Header.Add("Authorization", "Bearer "+tt.bearer)
			}
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}
