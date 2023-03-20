package caleta

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

var (
	providerConfiguration = configs.ProviderConf{
		URL: "http://caleta-test",
		Auth: map[string]any{
			"operator_id": "oid",
			"signing_key": testingPrivateKey,
		},
	}
)

type mockRestClient struct {
	rest.HTTPClient
	JSONFunc func(ctx context.Context, req *rest.HTTPRequest, resp any) error
}

func (m mockRestClient) Post(ctx context.Context, p rest.Parser, req *rest.HTTPRequest, resp any) error {
	return m.JSONFunc(ctx, req, resp)
}

func (m mockRestClient) Get(ctx context.Context, p rest.Parser, req *rest.HTTPRequest, resp any) error {
	return m.JSONFunc(ctx, req, resp)
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
func Test_requestGameLaunch(t *testing.T) {
	type tests struct {
		name   string
		body   GameUrlBody
		jsonFn func(_ context.Context, req *rest.HTTPRequest, resp any) error
		config configs.ProviderConf
		want   any
		e      error
	}

	var gameLaunchTests = []tests{
		{
			name: "successful game launch",
			jsonFn: func(_ context.Context, req *rest.HTTPRequest, resp any) error {
				assert.Equal(t, "http://caleta-test/api/game/url", req.URL)
				assert.NotEmpty(t, req.Headers["X-Auth-Signature"])
				assert.NotNil(t, req.Body)

				reflect.ValueOf(resp).
					Elem().
					Set(reflect.ValueOf(InlineResponse200{Url: testutils.Ptr("valid-game-url")}))
				return nil
			},
			config: providerConfiguration,
			body:   GameUrlBody{},
			want:   &InlineResponse200{Url: testutils.Ptr("valid-game-url")},
			e:      nil,
		},
		{
			name: "successful game launch without signing_key",
			jsonFn: func(_ context.Context, req *rest.HTTPRequest, resp any) error {
				assert.Empty(t, req.Headers["X-Auth-Signature"])

				reflect.ValueOf(resp).
					Elem().
					Set(reflect.ValueOf(InlineResponse200{Url: testutils.Ptr("valid-game-url")}))
				return nil
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
			body: GameUrlBody{},
			want: &InlineResponse200{Url: testutils.Ptr("valid-game-url")},
			e:    nil,
		},
		{
			name: "error post request",
			jsonFn: func(_ context.Context, _ *rest.HTTPRequest, _ any) error {
				return errors.New("post error")
			},
			config: providerConfiguration,
			want:   &InlineResponse200{},
			e:      errors.New("post error"),
		},
		{
			name: "no error but url missing from response",
			jsonFn: func(_ context.Context, _ *rest.HTTPRequest, resp any) error {
				reflect.ValueOf(resp).
					Elem().
					Set(reflect.ValueOf(InlineResponse200{}))
				return nil
			},
			config: providerConfiguration,
			body:   GameUrlBody{},
			want:   &InlineResponse200{},
			e:      nil,
		},
	}

	for _, test := range gameLaunchTests {
		t.Run(test.name, func(t *testing.T) {
			api, err := NewAPIClient(mockRestClient{JSONFunc: test.jsonFn}, test.config)
			assert.NoError(t, err)
			result, err := api.requestGameLaunch(context.TODO(), test.body)
			if test.e != nil {
				assert.EqualError(t, err, test.e.Error())
			}
			assert.Equal(t, test.want, result)
		})
	}
}

func Test_getGameRoundRender(t *testing.T) {

	type tests struct {
		name        string
		gameRoundID string
		jsonFn      func(_ context.Context, req *rest.HTTPRequest, resp any) error
		config      configs.ProviderConf
		want        any
		e           error
	}

	var gameRenderTests = []tests{
		{
			name:        "get game render page",
			gameRoundID: "909",
			jsonFn: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
				r := resp.(*gameRoundRenderResponse)
				url := "successUrl"
				r.Url = &url
				return nil
			},
			config: providerConfiguration,
			want:   &gameRoundRenderResponse{InlineResponse200: InlineResponse200{Url: testutils.Ptr("successUrl")}},
			e:      nil,
		},
		{
			name:        "get game render page missing url",
			gameRoundID: "909",
			jsonFn: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
				return nil
			},
			config: providerConfiguration,
			want:   &gameRoundRenderResponse{},
			e:      nil,
		},
		{
			name:        "get game render page error from response",
			gameRoundID: "909",
			jsonFn: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
				r := resp.(*gameRoundRenderResponse)
				r.Code = 100
				r.Message = "Bad Stuff"
				return nil
			},
			config: providerConfiguration,
			want:   &gameRoundRenderResponse{Code: 100, Message: "Bad Stuff"},
			e:      nil,
		},
		{
			name:        "get game render page http error",
			gameRoundID: "909",
			jsonFn: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
				return fmt.Errorf("some error")
			},
			config: providerConfiguration,
			want:   &gameRoundRenderResponse{},
			e:      errors.New("some error"),
		},
	}

	for _, test := range gameRenderTests {
		t.Run(test.name, func(t *testing.T) {
			api, err := NewAPIClient(mockRestClient{JSONFunc: test.jsonFn}, test.config)
			assert.NoError(t, err)
			result, err := api.getGameRoundRender(context.TODO(), test.gameRoundID, "")
			if test.e != nil {
				assert.EqualError(t, err, test.e.Error())
			}
			assert.Equal(t, test.want, result)
		})
	}
}

func Test_roundTransactions(t *testing.T) {

	type tests struct {
		name        string
		gameRoundID string
		jsonFn      func(ctx context.Context, req *rest.HTTPRequest, resp any) error
		want        *transactionResponse
		e           error
	}

	var getRoundTransactionsTests = []tests{
		{
			name:        "Successful get of round transactions",
			gameRoundID: "909",
			want: &transactionResponse{
				RoundID: "909",
				RoundTransactions: &[]roundTransaction{
					{
						RoundID: 909,
					}}},
			jsonFn: func(_ context.Context, req *rest.HTTPRequest, resp any) error {
				assert.Equal(t, "http://caleta-test/api/transactions/round", req.URL)
				assert.NotEmpty(t, req.Headers["X-Auth-Signature"])
				assert.Empty(t, req.Query)

				body := req.Body.(transactionRequestBody)
				assert.Equal(t, "909", body.RoundID)

				reflect.ValueOf(resp).
					Elem().
					Set(reflect.ValueOf(transactionResponse{
						RoundID: "909",
						RoundTransactions: &[]roundTransaction{
							{
								RoundID: 909,
							}}}))
				return nil
			},
			e: nil,
		},
	}

	for _, test := range getRoundTransactionsTests {
		t.Run(test.name, func(t *testing.T) {
			api, err := NewAPIClient(mockRestClient{JSONFunc: test.jsonFn}, providerConfiguration)
			assert.NoError(t, err)
			assert.NotNil(t, api)
			tx, err := api.getRoundTransactions(context.TODO(), test.gameRoundID)
			if test.e != nil {
				assert.EqualError(t, err, test.e.Error())
			}
			assert.Equal(t, (*test.want), (*tx))
		})
	}
}
