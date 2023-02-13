package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/creasty/defaults"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/application/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/httpclient"
)

func TestTimeout(t *testing.T) {
	type args struct {
		sleepTimeout     time.Duration
		httpServerConfig configs.HTTPServerConfig
		httpClientConfig configs.HTTPClientConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success with default timeouts",
			args: args{
				sleepTimeout:     0 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{},
				httpClientConfig: configs.HTTPClientConfig{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "handler sleep doesn't count towards server read timeout",
			args: args{
				sleepTimeout: 20 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{
					ReadTimeout: 10 * time.Millisecond,
				},
				httpClientConfig: configs.HTTPClientConfig{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "handler sleep doesn't count towards server write timeout",
			args: args{
				sleepTimeout: 20 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{
					WriteTimeout: 10 * time.Millisecond,
				},
				httpClientConfig: configs.HTTPClientConfig{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "handler sleep doesn't count towards server idle timeout",
			args: args{
				sleepTimeout: 20 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{
					IdleTimeout: 10 * time.Millisecond,
				},
				httpClientConfig: configs.HTTPClientConfig{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "client read timeouts",
			args: args{
				sleepTimeout:     20 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{},
				httpClientConfig: configs.HTTPClientConfig{
					ReadTimeout: 10 * time.Millisecond,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "handler sleep doesn't count towards client write timeout",
			args: args{
				sleepTimeout:     20 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{},
				httpClientConfig: configs.HTTPClientConfig{
					WriteTimeout: 10 * time.Millisecond,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "handler sleep doesn't count towards client idle timeout",
			args: args{
				sleepTimeout:     20 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{},
				httpClientConfig: configs.HTTPClientConfig{
					IdleTimeout: 10 * time.Millisecond,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "client request timeout",
			args: args{
				sleepTimeout:     20 * time.Millisecond,
				httpServerConfig: configs.HTTPServerConfig{},
				httpClientConfig: configs.HTTPClientConfig{
					RequestTimeout: 10 * time.Millisecond,
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valkyrieConfig := configs.ValkyrieConfig{
				Logging: configs.LogConfig{
					Level: "fatal",
				},
				Providers: []configs.ProviderConf{},
				Pam: configs.PamConf{
					"name": "generic",
				},
				HTTPServer: tt.args.httpServerConfig,
				HTTPClient: tt.args.httpClientConfig,
			}
			// init the default timeouts on structs
			err := defaults.Set(&valkyrieConfig.HTTPServer)
			assert.NoError(t, err)
			err = defaults.Set(&valkyrieConfig.HTTPClient)
			assert.NoError(t, err)

			// configure valkyrie to listen on free ports
			providerPort, _ := testutils.GetFreePort()
			operatorPort, _ := testutils.GetFreePort()
			valkyrieConfig.HTTPServer.ProviderAddress = fmt.Sprintf("localhost:%d", providerPort)
			valkyrieConfig.HTTPServer.OperatorAddress = fmt.Sprintf("localhost:%d", operatorPort)
			valkyrie := NewValkyrie(context.TODO(), &valkyrieConfig)

			// add timeout handler used in test
			type result struct {
				Status int
			}
			valkyrie.provider.Get("/timeout/:sleep", func(ctx *fiber.Ctx) error {
				sleep, err := ctx.ParamsInt("sleep")
				assert.NoError(t, err)
				time.Sleep(time.Duration(sleep) * time.Millisecond)
				return ctx.JSON(&result{1})
			})

			valkyrie.Start()
			client := httpclient.Create(valkyrieConfig.HTTPClient)

			req := &httpclient.HTTPRequest{
				Headers: map[string]string{"Accept": "application/json"},
				URL:     fmt.Sprintf("http://localhost:%d/timeout/%d", providerPort, tt.args.sleepTimeout.Milliseconds()),
			}
			resp := &result{}

			// make a request towards /timeout/:sleep handler and check for timeout
			err = client.GetJSON(context.TODO(), req, resp)
			if !tt.wantErr(t, err) {
				assert.Equal(t, 1, resp.Status)
			}

			valkyrie.Stop()
		})
	}

}
