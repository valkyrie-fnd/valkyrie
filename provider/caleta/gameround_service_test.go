package caleta

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/rest"
	"github.com/valyala/fasthttp"
)

func Test_Requesting_Gameround_Render_Page(t *testing.T) {
	sut, _ := NewCaletaService(configs.ProviderConf{}, mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
		r := resp.(*gameRoundRenderResponse)
		url := "successUrl"
		r.Url = &url
		return nil
	}})
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, _ := sut.GetGameRoundRender(ctx, "gameRoundId")
	assert.Equal(t, "successUrl", res)
}

func Test_Requesting_Gameround_Render_Page_error_missing_url(t *testing.T) {
	sut, _ := NewCaletaService(configs.ProviderConf{}, mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
		return nil
	}})
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, err := sut.GetGameRoundRender(ctx, "gameRoundId")
	assert.Equal(t, "", res)
	assert.EqualError(t, err, "HTTP 400: 0: ")
}

func Test_Requesting_Gameround_Render_Page_error_from_response(t *testing.T) {
	sut, _ := NewCaletaService(configs.ProviderConf{}, mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
		r := resp.(*gameRoundRenderResponse)
		r.Code = 100
		r.Message = "Bad Stuff"
		return nil
	}})
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, err := sut.GetGameRoundRender(ctx, "gameRoundId")
	assert.Equal(t, "", res)
	assert.EqualError(t, err, "HTTP 400: 100: Bad Stuff")
}

func Test_Requesting_Gameround_Render_Page_error_posting(t *testing.T) {
	sut, _ := NewCaletaService(configs.ProviderConf{}, mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
		return fmt.Errorf("Some network error")
	}})
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, err := sut.GetGameRoundRender(ctx, "gameRoundId")
	assert.Equal(t, "", res)
	assert.EqualError(t, err, "Some network error")
}
