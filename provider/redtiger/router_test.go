package redtiger

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

func TestRouteMethod(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			"Auth accepts post",
			"/test/redtiger/auth",
			"POST",
			200,
		},
		{
			"Auth rejects Get",
			"/test/redtiger/auth",
			"GET",
			405,
		},
		{
			"Stake accepts post",
			"/test/redtiger/stake",
			"POST",
			200,
		},
		{
			"Stake rejects Get",
			"/test/redtiger/stake",
			"GET",
			405,
		},
		{
			"Payout accepts post",
			"/test/redtiger/payout",
			"POST",
			200,
		},
		{
			"Payout rejects Get",
			"/test/redtiger/payout",
			"GET",
			405,
		},
		{
			"Refund accepts post",
			"/test/redtiger/refund",
			"POST",
			200,
		},
		{
			"Refund rejects Get",
			"/test/redtiger/refund",
			"GET",
			405,
		},
		{
			"Promo buyin accepts post",
			"/test/redtiger/promo/buyin",
			"POST",
			200,
		},
		{
			"Promo buyin rejects Get",
			"/test/redtiger/promo/buyin",
			"GET",
			405,
		},
		{
			"Promo settle accepts post",
			"/test/redtiger/promo/settle",
			"POST",
			200,
		},
		{
			"Promo settle rejects Get",
			"/test/redtiger/promo/settle",
			"GET",
			405,
		},
		{
			"Promo refund accepts post",
			"/test/redtiger/promo/refund",
			"POST",
			200,
		},
		{
			"Promo refund rejects Get",
			"/test/redtiger/promo/refund",
			"GET",
			405,
		},
	}
	router, _ := NewProviderRouter(configs.ProviderConf{
		Auth: map[string]any{"api_key": "pelle", "recon_token": "recon"},
	}, &NilController{})
	authHeader := "Basic pelle"
	app := fiber.New()
	reg := provider.NewRegistry(app, "/test")
	_ = reg.Register(router)
	baseRequest := `{"token":"abc123", "userId":"1", "currency":"USD"}`
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(baseRequest))
			req.Header.Add("Authorization", authHeader)
			req.Header.Add("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(tt, err)
			assert.Equal(tt, test.expectedStatus, resp.StatusCode)
		})
	}
}

func TestApiKeyMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		apiKey           string
		expectedStatus   int
		expectedResponse string
	}{
		{
			"Auth should succeed when api key is correct",
			"/test/redtiger/auth",
			"pelle",
			200,
			"",
		},
		{
			"Auth should Fail when api key is incorrect",
			"/test/redtiger/auth",
			"pelle-fail",
			401,
			`{"success":false,"error":{"message":"API authentication error","code":100}}`,
		},
		{
			"Stake should succeed when api key is correct",
			"/test/redtiger/stake",
			"pelle",
			200,
			"",
		},
		{
			"Stake should Fail when api key is incorrect",
			"/test/redtiger/stake",
			"pelle-fail",
			401,
			`{"success":false,"error":{"message":"API authentication error","code":100}}`,
		},
	}
	router, _ := NewProviderRouter(configs.ProviderConf{
		Auth: map[string]any{"api_key": "pelle", "recon_token": "recon"},
	}, &NilController{})
	app := fiber.New()
	reg := provider.NewRegistry(app, "/test")
	_ = reg.Register(router)
	baseRequest := `{"token":"abc123", "userId":"1", "currency":"USD"}`
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(baseRequest))
			req.Header.Add("Authorization", fmt.Sprintf("Basic %s", test.apiKey))
			req.Header.Add("Content-Type", "application/json")
			resp, err := app.Test(req)

			b, _ := io.ReadAll(resp.Body)
			assert.NoError(tt, err)
			assert.Equal(tt, test.expectedStatus, resp.StatusCode)
			if test.expectedResponse == "" {
				assert.Empty(tt, b)
			} else {
				assert.JSONEq(tt, test.expectedResponse, string(b))
			}
		})
	}
}

func TestDeclineTokenMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		contentType      string
		userToken        string
		expectedStatus   int
		expectedResponse string
	}{
		{
			"Stake should work when token is individual",
			"/test/redtiger/stake",
			"application/json",
			"Some-secret-token",
			200,
			"",
		},
		{
			"Stake should fail if content-type is invalid",
			"/test/redtiger/stake",
			"",
			"Some-secret-token",
			401,
			`{"success":false,"error":{"message":"Invalid base request. err: Unprocessable Entity","code":200}}`,
		},
		{
			"Stake should fail if token is recon token",
			"/test/redtiger/stake",
			"application/json",
			"recon",
			401,
			`{"success":false,"error":{"message":"API authentication error","code":301}}`,
		},
	}
	router, _ := NewProviderRouter(configs.ProviderConf{
		Auth: map[string]any{"api_key": "pelle", "recon_token": "recon"},
	}, &NilController{})
	app := fiber.New()
	reg := provider.NewRegistry(app, "/test")
	_ = reg.Register(router)
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			baseRequest := fmt.Sprintf(`{"token":"%s", "userId":"1", "currency":"USD"}`, test.userToken)
			req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(baseRequest))
			req.Header.Add("Authorization", "Basic pelle")
			req.Header.Add("Content-Type", test.contentType)
			resp, err := app.Test(req)

			b, _ := io.ReadAll(resp.Body)
			assert.NoError(tt, err)
			assert.Equal(tt, test.expectedStatus, resp.StatusCode)
			if test.expectedResponse == "" {
				assert.Empty(tt, b)
			} else {
				assert.JSONEq(tt, test.expectedResponse, string(b))
			}
		})
	}
}

type NilController struct{}

func (nc *NilController) Auth(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) Stake(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) Payout(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) Refund(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) PromoBuyin(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) PromoSettle(_ *fiber.Ctx) error {
	return nil
}
func (nc *NilController) PromoRefund(_ *fiber.Ctx) error {
	return nil
}
