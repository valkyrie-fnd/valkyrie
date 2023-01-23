package evolution

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

type MockClient struct {
	rest.HTTPClientJSONInterface
	PostJSONFunc func(ctx context.Context, req *rest.HTTPRequest, resp any) error
}

func (m MockClient) PostJSON(ctx context.Context, req *rest.HTTPRequest, resp any) error {
	return m.PostJSONFunc(ctx, req, resp)
}

func TestGameLaunchService_GameLaunch(t *testing.T) {
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	var expectedRequest = provider.GameLaunchRequest{
		Currency:       "SEK",
		ProviderGameID: "i",
		PlayerID:       "p",
		Casino:         "c",
		Country:        "SE",
		Language:       "sv",
		SessionIP:      "i",
		LaunchConfig:   map[string]interface{}{},
	}
	var expectedHeaders = provider.GameLaunchHeaders{
		SessionKey: "key",
	}
	var expectedResponse = UserAuthenticationResponse{
		Entry: "entry",
	}

	type fields struct {
		Auth   AuthConf
		C      *configs.ProviderConf
		Client rest.HTTPClientJSONInterface
	}
	type args struct {
		ctx *fiber.Ctx
		g   *provider.GameLaunchRequest
		h   *provider.GameLaunchHeaders
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "successful gamelaunch",
			fields: fields{
				C: &configs.ProviderConf{
					Name: "Evolution",
					URL:  "evo-url",
				},
				Auth: AuthConf{CasinoToken: "casino-token", CasinoKey: "casino-key"},
				Client: MockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "evo-url/ua/v1/casino-key/casino-token", req.URL)
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			args: args{
				ctx: ctx,
				g:   &expectedRequest,
				h:   &expectedHeaders,
			},
			want:    "evo-urlentry",
			wantErr: assert.NoError,
		},
		{
			name: "error gamelaunch",
			fields: fields{
				Auth: AuthConf{CasinoToken: "casino-token", CasinoKey: "casino-key"},
				C: &configs.ProviderConf{
					Name: "Evolution",
					URL:  "evo-url",
				},
				Client: MockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					return assert.AnError
				}},
			},
			args: args{
				ctx: ctx,
				g:   &expectedRequest,
				h:   &expectedHeaders,
			},
			want: "",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := EvoService{
				Auth:   tt.fields.Auth,
				Conf:   tt.fields.C,
				Client: tt.fields.Client,
			}
			got, err := service.GameLaunch(tt.args.ctx, tt.args.g, tt.args.h)
			if !tt.wantErr(t, err, fmt.Sprintf("GameLaunch(%v, %v, %v)", tt.args.ctx, tt.args.g, tt.args.h)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GameLaunch(%v, %v, %v)", tt.args.ctx, tt.args.g, tt.args.h)
		})
	}
}

func TestGameRoundRender(t *testing.T) {
	sut := EvoService{}
	_, err := sut.GetGameRound(nil, "")
	assert.EqualError(t, err, "Not available")
}
