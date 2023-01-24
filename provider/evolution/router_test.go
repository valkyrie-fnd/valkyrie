package evolution

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

func TestNewProviderRouter(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			"check accepts post",
			"/test/evolution/check?authToken=pelle",
			http.MethodPost,
			200,
		},
		{
			"check rejects get",
			"/test/evolution/check?authToken=pelle",
			http.MethodGet,
			405,
		},
		{
			"check requires correct token",
			"/test/evolution/check?authToken=olle",
			http.MethodGet,
			401,
		},
		{
			"balance accepts post",
			"/test/evolution/balance?authToken=pelle",
			http.MethodPost,
			200,
		},
		{
			"debit accepts post",
			"/test/evolution/debit?authToken=pelle",
			http.MethodPost,
			200,
		},
		{
			"credit accepts post",
			"/test/evolution/credit?authToken=pelle",
			http.MethodPost,
			200,
		},
		{
			"cancel accepts post",
			"/test/evolution/cancel?authToken=pelle",
			http.MethodPost,
			200,
		},
		{
			"promo_payout accepts post",
			"/test/evolution/promo_payout?authToken=pelle",
			http.MethodPost,
			200,
		},
	}

	router, _ := NewProviderRouter(configs.ProviderConf{
		Auth:     map[string]any{"api_key": "pelle"},
		BasePath: "/evolution",
	}, &NilController{})
	app := fiber.New()
	reg := provider.NewRegistry(app, "/test")
	_ = reg.Register(router)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

type NilController struct{}

func (nc *NilController) Check(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) Balance(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) Debit(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) Credit(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) Cancel(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) PromoPayout(_ *fiber.Ctx) error {
	return nil
}
