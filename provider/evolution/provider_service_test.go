package evolution

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
	"github.com/valkyrie-fnd/valkyrie/valkhttp"
)

type MockClient struct {
	valkhttp.HTTPClient
	PostJSONFunc func(ctx context.Context, req *valkhttp.HTTPRequest, resp any) error
	GetFunc      func(ctx context.Context, req *valkhttp.HTTPRequest, resp *[]byte) error
}

func (m MockClient) Post(ctx context.Context, p valkhttp.Parser, req *valkhttp.HTTPRequest, resp any) error {
	return m.PostJSONFunc(ctx, req, resp)
}

func (m MockClient) Get(ctx context.Context, p valkhttp.Parser, req *valkhttp.HTTPRequest, resp any) error {
	return m.GetFunc(ctx, req, resp.(*[]byte))
}

type fields struct {
	Auth   AuthConf
	C      *configs.ProviderConf
	Client valkhttp.HTTPClient
}

func TestGameLaunchService_GameLaunch(t *testing.T) {
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)
	type args struct {
		ctx *fiber.Ctx
		g   *provider.GameLaunchRequest
		h   *provider.GameLaunchHeaders
	}
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
				Client: MockClient{PostJSONFunc: func(ctx context.Context, req *valkhttp.HTTPRequest, resp any) error {
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
				Client: MockClient{PostJSONFunc: func(ctx context.Context, req *valkhttp.HTTPRequest, resp any) error {
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
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)
	type args struct {
		c *fiber.Ctx
		g *provider.GameRoundRenderRequest
	}
	a := args{
		c: c,
		g: &provider.GameRoundRenderRequest{
			GameRoundID: "123",
		},
	}
	ac := AuthConf{
		CasinoToken: "casinoToken",
		CasinoKey:   "casinoKey",
	}
	conf := &configs.ProviderConf{
		URL: "evo-base-url.com",
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCode int
		wantBody string
		wantErr  error
	}{
		{
			name: "Returns ok with body of response",
			args: a,
			fields: fields{
				Auth: ac,
				C:    conf,
				Client: MockClient{
					GetFunc: func(ctx context.Context, req *valkhttp.HTTPRequest, resp *[]byte) error {
						*resp = []byte("ResponseBody")
						return nil
					},
				},
			},
			wantCode: 200,
			wantBody: "ResponseBody",
			wantErr:  nil,
		},
		{
			name: "Pass gameRoundId to request",
			args: a,
			fields: fields{
				Auth: ac,
				C:    conf,
				Client: MockClient{
					GetFunc: func(ctx context.Context, req *valkhttp.HTTPRequest, resp *[]byte) error {
						assert.Equal(t, "123", req.Query["gameId"])
						return nil
					},
				},
			},
			wantCode: 200,
			wantBody: "",
			wantErr:  nil,
		},
		{
			name: "Pass base64 encoded casinoKey:casinoToken as basic Auth header",
			args: a,
			fields: fields{
				Auth: ac,
				C:    conf,
				Client: MockClient{
					GetFunc: func(ctx context.Context, req *valkhttp.HTTPRequest, resp *[]byte) error {
						encoded := base64.StdEncoding.EncodeToString(([]byte("casinoKey:casinoToken")))
						assert.Equal(t, fmt.Sprintf("Basic %s", encoded), req.Headers["Authorization"])
						return nil
					},
				},
			},
			wantCode: 200,
			wantBody: "",
			wantErr:  nil,
		},
		{
			name: "Returns error from service call",
			args: a,
			fields: fields{
				Auth: ac,
				C:    conf,
				Client: MockClient{
					GetFunc: func(ctx context.Context, req *valkhttp.HTTPRequest, resp *[]byte) error {
						return fmt.Errorf("Failed request")
					},
				},
			},
			wantCode: 400,
			wantBody: "",
			wantErr:  fmt.Errorf("Failed request"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			sut := EvoService{
				Auth:   test.fields.Auth,
				Conf:   test.fields.C,
				Client: test.fields.Client,
			}
			resCode, err := sut.GetGameRoundRender(test.args.c, *test.args.g)
			assert.Equal(tt, test.wantCode, resCode)
			if test.wantErr != nil {
				assert.Error(tt, test.wantErr, err)
			}
			if test.wantBody != "" {
				assert.Equal(tt, test.wantBody, string(test.args.c.Response().Body()))
			}
		})
	}
}
