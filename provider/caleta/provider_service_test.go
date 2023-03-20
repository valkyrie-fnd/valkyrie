package caleta

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
	"github.com/valyala/fasthttp"
)

var (
	providerConf = configs.ProviderConf{
		URL: "http://caleta-test",
		Auth: map[string]any{
			"operator_id": "oid",
			"signing_key": testingPrivateKey,
		},
		ProviderSpecific: map[string]any{
			"game_launch_type": "request",
		},
	}
	request = &provider.GameLaunchRequest{
		Currency:       "USD",
		ProviderGameID: "game-id",
		PlayerID:       "1",
		Country:        "SE",
		Language:       "sv",
		LaunchConfig: map[string]interface{}{
			"lobby_url":      "lobby_url",
			"sub_partner_id": "sub_partner_id",
			"deposit_url":    "deposit_url",
		},
	}
	headers = &provider.GameLaunchHeaders{SessionKey: "123"}
)

type mockAPIClient struct {
	rest.HTTPClient
	getRoundTransactionsFn func(ctx context.Context, gameRoundID string) (*transactionResponse, error)
	requestGameLaunchFn    func(ctx context.Context, body GameUrlBody) (*InlineResponse200, error)
	getGameRoundRenderFn   func(ctx context.Context, gameRoundID string) (*gameRoundRenderResponse, error)
}

func (api *mockAPIClient) getRoundTransactions(ctx context.Context, gameRoundID string) (*transactionResponse, error) {
	return api.getRoundTransactionsFn(ctx, gameRoundID)
}

func (api *mockAPIClient) requestGameLaunch(ctx context.Context, body GameUrlBody) (*InlineResponse200, error) {
	return api.requestGameLaunchFn(ctx, body)
}

func (api *mockAPIClient) getGameRoundRender(ctx context.Context, gameRoundID, casinoID string) (*gameRoundRenderResponse, error) {
	return api.getGameRoundRenderFn(ctx, gameRoundID)
}

func TestStaticUrlGameLaunch(t *testing.T) {
	type args struct {
		req     *provider.GameLaunchRequest
		headers *provider.GameLaunchHeaders
	}

	type tests struct {
		name   string
		args   args
		config configs.ProviderConf
		want   string
		e      error
	}

	var gameLaunchTests = []tests{
		{
			name: "successful game launch",
			config: configs.ProviderConf{
				URL: "https://staging.the-rgs.com",
				Auth: map[string]any{
					"operator_id": "valkyrie",
				},
				ProviderSpecific: map[string]any{
					"game_launch_type": "static",
				},
			},
			args: args{
				req:     request,
				headers: headers,
			},
			want: "https://staging.the-rgs.com/open_game?country=SE&currency=USD&deposit_url=deposit_url&game_code=game-id&lang=sv&lobby_url=lobby_url&operator_id=valkyrie&sub_partner_id=sub_partner_id&token=123&user=1",
			e:    nil,
		},
	}

	for _, test := range gameLaunchTests {
		t.Run(test.name, func(t *testing.T) {
			s, err := NewCaletaService(&mockAPIClient{}, test.config)
			assert.NoError(t, err)
			result, err := s.GameLaunch(nil, test.args.req, test.args.headers)
			if test.e != nil {
				assert.EqualError(t, err, test.e.Error())
			}
			assert.Equal(t, test.want, result)
		})
	}
}

func TestRequestingGameLaunch(t *testing.T) {
	type args struct {
		req     *provider.GameLaunchRequest
		headers *provider.GameLaunchHeaders
	}

	type tests struct {
		name         string
		args         args
		gameLaunchFn func(ctx context.Context, body GameUrlBody) (*InlineResponse200, error)
		config       configs.ProviderConf
		want         string
		e            error
	}

	var gameLaunchTests = []tests{
		{
			name: "successful game launch",
			gameLaunchFn: func(ctx context.Context, body GameUrlBody) (*InlineResponse200, error) {
				return &InlineResponse200{Url: testutils.Ptr("valid-game-url")}, nil
			},
			config: providerConf,
			args: args{
				req:     request,
				headers: headers,
			},
			want: "valid-game-url",
			e:    nil,
		},
		{
			name: "successful game launch without signing_key",
			gameLaunchFn: func(ctx context.Context, body GameUrlBody) (*InlineResponse200, error) {
				return &InlineResponse200{Url: testutils.Ptr("valid-game-url")}, nil
			},
			config: configs.ProviderConf{
				URL: "http://caleta-test",
				ProviderSpecific: map[string]any{
					"game_launch_type": "request",
				},
				Auth: map[string]any{
					"operator_id": "oid",
				},
			},
			args: args{
				req:     request,
				headers: headers,
			},
			want: "valid-game-url",
			e:    nil,
		},
		{
			name: "error post request",
			gameLaunchFn: func(ctx context.Context, body GameUrlBody) (*InlineResponse200, error) {
				return nil, fmt.Errorf("post error")
			},
			config: providerConf,
			args: args{
				req:     request,
				headers: headers,
			},
			want: "",
			e:    errors.New("post error"),
		},
		{
			name: "error url missing from response",
			gameLaunchFn: func(ctx context.Context, body GameUrlBody) (*InlineResponse200, error) {
				return &InlineResponse200{}, nil
			},
			config: providerConf,
			args: args{
				req:     request,
				headers: headers,
			},
			want: "",
			e:    errors.New("url missing from response"),
		},
	}

	for _, test := range gameLaunchTests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			s, err := NewCaletaService(&mockAPIClient{requestGameLaunchFn: test.gameLaunchFn}, test.config)
			assert.NoError(t, err)
			result, err := s.GameLaunch(ctx, test.args.req, test.args.headers)
			if test.e != nil {
				assert.EqualError(t, err, test.e.Error())
			}
			assert.Equal(t, test.want, result)
		})
	}
}

func Test_getLaunchConfig(t *testing.T) {
	launchConfigInput := map[string]interface{}{
		"lobby_url":      "lobby_url",
		"sub_partner_id": "sub_partner_id",
		"deposit_url":    "deposit_url",
	}

	launchConfig, err := getLaunchConfig(launchConfigInput)
	assert.NoError(t, err)
	assert.Equal(t, "lobby_url", launchConfig.LobbyURL)
	assert.Equal(t, "sub_partner_id", launchConfig.SubPartnerID)
	assert.Equal(t, "deposit_url", *launchConfig.DepositURL)
}

func Test_getLaunchConfigEmpty(t *testing.T) {
	launchConfigInput := map[string]interface{}{}

	launchConfig, err := getLaunchConfig(launchConfigInput)

	assert.NoError(t, err)
	assert.Empty(t, launchConfig.LobbyURL)
	assert.Empty(t, launchConfig.SubPartnerID)
	assert.Nil(t, launchConfig.DepositURL)
}

func Test_getLaunchConfigError(t *testing.T) {
	launchConfigInput := map[string]interface{}{
		"lobby_url": 1,
	}

	launchConfig, err := getLaunchConfig(launchConfigInput)

	assert.Error(t, err, "1 error(s) decoding:\n\n* 'lobby_url' expected type 'string', got unconvertible type 'int', value: '1'")
	assert.Nil(t, launchConfig)
}

func Test_getGameUrlBody(t *testing.T) {
	service, err := NewCaletaService(&mockAPIClient{}, providerConf)
	assert.NoError(t, err)
	body, err := service.getGameLaunchBody(request, headers)
	assert.NoError(t, err)

	assert.Equal(t, request.ProviderGameID, body.GameCode)
	assert.Equal(t, request.Country, string(body.Country))
	assert.Equal(t, request.Language, string(body.Lang))
	assert.Equal(t, request.Currency, string(body.Currency))
	assert.Equal(t, request.PlayerID, *body.User)
	assert.Equal(t, headers.SessionKey, *body.Token)
	assert.Equal(t, providerConf.Auth["operator_id"], body.OperatorId)
	assert.Equal(t, request.LaunchConfig["sub_partner_id"], body.SubPartnerId)
	assert.Equal(t, request.LaunchConfig["lobby_url"], body.LobbyUrl)
	assert.Equal(t, request.LaunchConfig["deposit_url"], *body.DepositUrl)
}

func Test_Requesting_Gameround_Render_Page(t *testing.T) {
	service, err := NewCaletaService(&mockAPIClient{getGameRoundRenderFn: func(ctx context.Context, gameRoundID string) (*gameRoundRenderResponse, error) {
		return &gameRoundRenderResponse{InlineResponse200: InlineResponse200{Url: testutils.Ptr("successUrl")}}, nil
	}}, configs.ProviderConf{})
	assert.NoError(t, err)
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, _ := service.GetGameRoundRender(ctx, provider.GameRoundRenderRequest{GameRoundID: "gameRoundId"})
	assert.Equal(t, 302, res)
	locHeader := ctx.Response().Header.Peek("Location")
	assert.Equal(t, "successUrl", string(locHeader))
}

func Test_Requesting_Gameround_Render_Page_error_missing_url(t *testing.T) {
	service, err := NewCaletaService(&mockAPIClient{getGameRoundRenderFn: func(ctx context.Context, gameRoundID string) (*gameRoundRenderResponse, error) {
		return &gameRoundRenderResponse{}, nil
	}}, configs.ProviderConf{})
	assert.NoError(t, err)
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, err := service.GetGameRoundRender(ctx, provider.GameRoundRenderRequest{GameRoundID: "gameRoundId"})
	assert.Equal(t, 400, res)
	assert.EqualError(t, err, "HTTP 400: 0: ")
}

func Test_Requesting_Gameround_Render_Page_error_from_response(t *testing.T) {
	service, err := NewCaletaService(&mockAPIClient{getGameRoundRenderFn: func(ctx context.Context, gameRoundID string) (*gameRoundRenderResponse, error) {
		return &gameRoundRenderResponse{Code: 100, Message: "Bad Stuff"}, nil
	}}, configs.ProviderConf{})
	assert.NoError(t, err)
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, err := service.GetGameRoundRender(ctx, provider.GameRoundRenderRequest{GameRoundID: "gameRoundId"})
	assert.Equal(t, 400, res)
	assert.EqualError(t, err, "HTTP 400: 100: Bad Stuff")
}

func Test_Requesting_Gameround_Render_Page_error_posting(t *testing.T) {
	service, err := NewCaletaService(&mockAPIClient{getGameRoundRenderFn: func(ctx context.Context, gameRoundID string) (*gameRoundRenderResponse, error) {
		return nil, fmt.Errorf("Some network error")
	}}, configs.ProviderConf{})
	assert.NoError(t, err)
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	res, err := service.GetGameRoundRender(ctx, provider.GameRoundRenderRequest{GameRoundID: "gameRoundId"})
	assert.Equal(t, 500, res)
	assert.EqualError(t, err, "Some network error")
}
