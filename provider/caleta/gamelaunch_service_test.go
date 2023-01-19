package caleta

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/rest"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

var (
	providerConf = configs.ProviderConf{
		URL: "http://caleta-test",
		Auth: map[string]any{
			"operator_id": "oid",
			"signing_key": testingPrivateKey,
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
					"game_launch_type": "Static",
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
			s, err := NewCaletaService(test.config, mockClient{})
			assert.NoError(t, err)
			result, err := s.GameLaunch(nil, test.args.req, test.args.headers)
			if test.e != nil {
				assert.EqualError(t, err, test.e.Error())
			}
			assert.Equal(t, test.want, result)
		})
	}
}

type mockClient struct {
	rest.HTTPClientJSONInterface
	PostJSONFunc func(ctx context.Context, req *rest.HTTPRequest, resp any) error
}

func (m mockClient) PostJSON(ctx context.Context, req *rest.HTTPRequest, resp any) error {
	return m.PostJSONFunc(ctx, req, resp)
}

func TestRequestingGameLaunch(t *testing.T) {
	type args struct {
		req     *provider.GameLaunchRequest
		headers *provider.GameLaunchHeaders
	}

	type tests struct {
		name   string
		args   args
		postFn func(ctx context.Context, req *rest.HTTPRequest, resp any) error
		config configs.ProviderConf
		want   string
		e      error
	}

	var gameLaunchTests = []tests{
		{
			name: "successful game launch",
			postFn: func(_ context.Context, req *rest.HTTPRequest, resp any) error {
				assert.Equal(t, "http://caleta-test/api/game/url", req.URL)
				assert.NotEmpty(t, req.Headers["X-Auth-Signature"])
				assert.NotNil(t, req.Body)

				reflect.ValueOf(resp).
					Elem().
					Set(reflect.ValueOf(InlineResponse200{Url: testutils.Ptr("valid-game-url")}))
				return nil
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
			postFn: func(_ context.Context, req *rest.HTTPRequest, resp any) error {
				assert.Empty(t, req.Headers["X-Auth-Signature"])

				reflect.ValueOf(resp).
					Elem().
					Set(reflect.ValueOf(InlineResponse200{Url: testutils.Ptr("valid-game-url")}))
				return nil
			},
			config: configs.ProviderConf{
				URL: "http://caleta-test",
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
			postFn: func(_ context.Context, _ *rest.HTTPRequest, _ any) error {
				return errors.New("post error")
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
			postFn: func(_ context.Context, _ *rest.HTTPRequest, resp any) error {
				reflect.ValueOf(resp).
					Elem().
					Set(reflect.ValueOf(InlineResponse200{}))
				return nil
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
			s, err := NewCaletaService(test.config, mockClient{PostJSONFunc: test.postFn})
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
	service, _ := NewCaletaService(providerConf, nil)
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

type mockSigner struct {
	SignFunc func([]byte) ([]byte, error)
}

func (m mockSigner) Sign(payload []byte) ([]byte, error) {
	return m.SignFunc(payload)
}

func Test_authHeaderSigner(t *testing.T) {
	expectedSignature := "signed"
	headerSigner := authHeaderSigner{signer: mockSigner{SignFunc: func(_ []byte) ([]byte, error) {
		return []byte(expectedSignature), nil
	}}}

	headers := map[string]string{}
	err := headerSigner.sign(&GameUrlBody{}, headers)
	assert.NoError(t, err)

	assert.Equal(t, expectedSignature, headers["X-Auth-Signature"])
}

func Test_authHeaderSignerError(t *testing.T) {
	headerSigner := authHeaderSigner{signer: mockSigner{SignFunc: func(_ []byte) ([]byte, error) {
		return nil, errors.New("sign error")
	}}}

	headers := map[string]string{}
	err := headerSigner.sign(&GameUrlBody{}, headers)
	assert.Error(t, err, "sign error")
	assert.Empty(t, headers)
}
