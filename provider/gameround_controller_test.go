package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valkyrie-fnd/valkyrie/valkhttp"
)

type gameRenderWant struct {
	status         int
	contentType    string
	locationHeader string
	body           string
	err            string
}
type gameRenderTestData struct {
	id              string
	gameRoundRender func(c *fiber.Ctx, req GameRoundRenderRequest) (int, error)
	want            gameRenderWant
}

var gameRenderTests = []gameRenderTestData{
	{
		id: "abc123",
		gameRoundRender: func(c *fiber.Ctx, req GameRoundRenderRequest) (int, error) {
			c.Set("Location", "redirectUrl")
			return 302, nil
		},
		want: gameRenderWant{
			status:         302,
			contentType:    "text/plain; charset=utf-8",
			locationHeader: "redirectUrl",
			err:            "",
			body:           "",
		},
	},
	{
		id: "abc123",
		gameRoundRender: func(c *fiber.Ctx, req GameRoundRenderRequest) (int, error) {
			return 400, valkhttp.HTTPError{Message: "Wrong Id Maybe", Code: 400}
		},
		want: gameRenderWant{
			status:         400,
			contentType:    "text/html",
			locationHeader: "",
			body:           "Wrong Id Maybe",
			err:            "",
		},
	},
	{
		id: "abc123",
		gameRoundRender: func(c *fiber.Ctx, req GameRoundRenderRequest) (int, error) {
			return 500, fmt.Errorf("SomeOtherError")
		},
		want: gameRenderWant{
			status:         500,
			contentType:    "text/plain; charset=utf-8",
			locationHeader: "",
			body:           "",
			err:            "SomeOtherError",
		},
	},
}

type gameRoundRenderService struct {
	gameRoundRender func(c *fiber.Ctx, req GameRoundRenderRequest) (int, error)
}

func (gs gameRoundRenderService) GameLaunch(_ *fiber.Ctx, gr *GameLaunchRequest, h *GameLaunchHeaders) (string, error) {
	return "", fmt.Errorf("Not Available")
}
func (gs gameRoundRenderService) GetGameRoundRender(c *fiber.Ctx, req GameRoundRenderRequest) (int, error) {
	if gs.gameRoundRender != nil {
		return gs.gameRoundRender(c, req)
	}
	return 200, nil
}

func TestGameRoundRender(t *testing.T) {
	for _, test := range gameRenderTests {
		testApp := fiber.New()
		ctrl := NewGameRoundController(gameRoundRenderService{test.gameRoundRender})
		testApp.Get("/gameRender/:gameRoundId", ctrl.GetGameRoundEndpoint)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/gameRender/%s", test.id), nil)

		resp, _ := testApp.Test(req, -1)
		assert.Equal(t, test.want.status, resp.StatusCode, "wrong status code")
		if test.want.err != "" {
			responseBody, _ := io.ReadAll(resp.Body)
			assert.Equal(t, test.want.err, string(responseBody), "Error message incorrect")
		}
		if test.want.contentType != "" {
			contentType := resp.Header.Get("Content-Type")
			assert.Equal(t, test.want.contentType, contentType)
		}
		if test.want.locationHeader != "" {
			location := resp.Header.Get("Location")
			assert.Equal(t, test.want.locationHeader, location)
		}
		if test.want.body != "" {
			responseBody, _ := io.ReadAll(resp.Body)
			assert.Equal(t, test.want.body, string(responseBody), "Response body is incorrect")
		}
	}
}
