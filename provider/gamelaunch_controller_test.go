package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

type GameLaunchWant struct {
	resBody     string
	err         string
	status      int
	contentType string
}
type GameLaunchTestData struct {
	req          GameLaunchRequest
	headers      GameLaunchHeaders
	contentType  string
	conf         configs.ProviderConf
	gameLaunchFn func(gr *GameLaunchRequest, h *GameLaunchHeaders) (string, error)
	want         GameLaunchWant
}

var gameLaunchTests = []GameLaunchTestData{
	{
		req:         GameLaunchRequest{},
		headers:     GameLaunchHeaders{},
		contentType: "",
		conf:        configs.ProviderConf{},
		want: GameLaunchWant{
			resBody:     "",
			err:         "\"Unprocessable Entity\"",
			status:      400,
			contentType: "application/json",
		},
	}, {
		req: GameLaunchRequest{
			Currency:       "sek",
			ProviderGameID: "1",
			PlayerID:       "1",
		},
		headers:     GameLaunchHeaders{},
		contentType: "application/json",
		conf:        configs.ProviderConf{},
		want: GameLaunchWant{
			resBody:     "",
			err:         "{\"SessionKey\":\"Key: 'GameLaunchHeaders.SessionKey' Error:Field validation for 'SessionKey' failed on the 'required' tag\"}",
			status:      400,
			contentType: "application/json",
		},
	}, {
		req: GameLaunchRequest{
			Currency:       "sek",
			ProviderGameID: "1",
			PlayerID:       "1",
		},
		headers: GameLaunchHeaders{
			SessionKey: "123",
		},
		contentType: "application/json",
		conf:        configs.ProviderConf{},
		gameLaunchFn: func(gr *GameLaunchRequest, h *GameLaunchHeaders) (string, error) {
			return "", errors.New("GameLaunchError")
		},
		want: GameLaunchWant{
			resBody:     "",
			err:         "\"GameLaunchError\"",
			status:      400,
			contentType: "application/json",
		},
	},
	{
		req: GameLaunchRequest{
			Currency:       "sek",
			ProviderGameID: "1",
			PlayerID:       "1",
		},
		headers: GameLaunchHeaders{
			SessionKey: "123",
		},
		contentType: "application/json",
		conf:        configs.ProviderConf{},
		gameLaunchFn: func(gr *GameLaunchRequest, h *GameLaunchHeaders) (string, error) {
			return "", rest.NewHTTPError(401, "Unauthorized")
		},
		want: GameLaunchWant{
			resBody:     "",
			err:         "HTTP 401: Unauthorized",
			status:      401,
			contentType: "text/plain; charset=utf-8",
		},
	},
	{
		req: GameLaunchRequest{
			Currency:       "sek",
			ProviderGameID: "1",
			PlayerID:       "1",
		},
		headers: GameLaunchHeaders{
			SessionKey: "123",
		},
		contentType:  "application/json",
		conf:         configs.ProviderConf{},
		gameLaunchFn: func(gr *GameLaunchRequest, h *GameLaunchHeaders) (string, error) { return "SomeLaunchUrl", nil },
		want: GameLaunchWant{
			resBody:     "{\"gameUrl\":\"SomeLaunchUrl\"}",
			err:         "",
			status:      200,
			contentType: "application/json",
		},
	},
}

type ProviderServiceMock struct {
	gameLaunchFn func(gr *GameLaunchRequest, h *GameLaunchHeaders) (string, error)
}

func (gs ProviderServiceMock) GameLaunch(_ *fiber.Ctx, gr *GameLaunchRequest, h *GameLaunchHeaders) (string, error) {
	if gs.gameLaunchFn != nil {
		return gs.gameLaunchFn(gr, h)
	}
	return "", nil
}
func (gs ProviderServiceMock) GetGameRoundRender(*fiber.Ctx, string) (string, error) {
	return "", fmt.Errorf("Not Available")
}

func TestGameLaunch(t *testing.T) {
	for _, test := range gameLaunchTests {
		testApp := fiber.New()
		ctrl := NewGameLaunchController(ProviderServiceMock{test.gameLaunchFn})
		testApp.Post("/gamelaunch", ctrl.GameLaunchEndpoint)
		body, _ := json.Marshal(test.req)
		req := httptest.NewRequest(http.MethodPost, "/gamelaunch", bytes.NewBuffer(body))
		if test.contentType != "" {
			req.Header.Set("Content-Type", test.contentType)
		}
		if test.headers.SessionKey != "" {
			req.Header.Set("X-Player-Token", test.headers.SessionKey)
		}

		resp, _ := testApp.Test(req, -1)
		assert.Equal(t, test.want.status, resp.StatusCode, "wrong status code")
		if test.want.err != "" {
			responseBody, _ := io.ReadAll(resp.Body)
			assert.Equal(t, test.want.err, string(responseBody), "Error message incorrect")
		}
		if test.want.resBody != "" {
			responseBody, _ := io.ReadAll(resp.Body)
			assert.Equal(t, test.want.resBody, string(responseBody), "Response body is incorrect")
		}
		if test.want.contentType != "" {
			contentType := resp.Header.Get("Content-Type")
			assert.Equal(t, test.want.contentType, contentType)
		}
	}
}
