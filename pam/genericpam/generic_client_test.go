package genericpam

import (
	"context"
	"reflect"
	"testing"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"

	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/rest"

	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	GetFunc      func(ctx context.Context, req *rest.HTTPRequest, resp any) error
	GetJSONFunc  func(ctx context.Context, req *rest.HTTPRequest, resp any) error
	PostJSONFunc func(ctx context.Context, req *rest.HTTPRequest, resp any) error
	PutJSONFunc  func(ctx context.Context, req *rest.HTTPRequest, resp any) error
}

func (m mockClient) GetJSON(ctx context.Context, req *rest.HTTPRequest, resp any) error {
	return m.GetJSONFunc(ctx, req, resp)
}
func (m mockClient) Get(ctx context.Context, req *rest.HTTPRequest, resp any) error {
	return m.GetFunc(ctx, req, resp)
}
func (m mockClient) PostJSON(ctx context.Context, req *rest.HTTPRequest, resp any) error {
	return m.PostJSONFunc(ctx, req, resp)
}

func (m mockClient) PutJSON(ctx context.Context, req *rest.HTTPRequest, resp any) error {
	return m.PutJSONFunc(ctx, req, resp)
}

func TestGenericPam_RefreshSession(t *testing.T) {
	var token = "tok"
	var expectedRequest = pam.RefreshSessionRequest{
		Params: pam.RefreshSessionParams{
			Provider:     "prov",
			XPlayerToken: "token",
		},
	}
	var expectedErrorResponse = pam.SessionResponse{
		Error: &pam.PamError{
			Code:    pam.PAMERRSESSIONNOTFOUND,
			Message: "test",
		},
		Status: "ERROR",
	}
	var expectedResponse = pam.SessionResponse{
		Session: &pam.Session{
			Country:  "SE",
			Currency: "SEK",
			Language: "sv",
			PlayerId: "1",
			Token:    token,
		},
		Status: "OK",
	}
	type fields struct {
		baseURL string
		apiKey  string
		rest    rest.HTTPClientJSONInterface
	}
	tests := []struct {
		name    string
		fields  fields
		mapper  pam.RefreshSessionRequestMapper
		wantErr error
		want    *pam.Session
	}{
		{
			name: "successful refresh session",
			fields: fields{
				"base",
				"key",
				mockClient{PutJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "base/players/session", req.URL)

					assert.Equal(t, expectedRequest.Params.XPlayerToken, req.Headers["X-Player-Token"])
					assert.Equal(t, "Bearer key", req.Headers["Authorization"])

					assert.Equal(t, expectedRequest.Params.Provider, req.Query["provider"])

					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.RefreshSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			want: expectedResponse.Session,
		},
		{
			name: "error refresh session",
			fields: fields{
				"base",
				"key",
				mockClient{PutJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					return assert.AnError
				}},
			},
			mapper: func() (context.Context, pam.RefreshSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "http client error",
				ValkErrorCode: pam.ValkErrUndefined,
				OrigError:     assert.AnError,
			},
		},
		{
			name: "error refresh session mapper",
			fields: fields{
				"base",
				"key",
				nil,
			},
			mapper: func() (context.Context, pam.RefreshSessionRequest, error) {
				return context.Background(), pam.RefreshSessionRequest{}, assert.AnError
			},
			wantErr: assert.AnError,
		},
		{
			name: "error refresh session response body",
			fields: fields{
				"base",
				"key",
				mockClient{PutJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedErrorResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.RefreshSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "PAM_ERR_SESSION_NOT_FOUND test",
				ValkErrorCode: pam.ValkErrOpSessionNotFound,
				OrigError:     expectedErrorResponse.Error,
			},
		},
		{
			name: "error refresh session response body nil session",
			fields: fields{
				"base",
				"key",
				mockClient{PutJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(pam.SessionResponse{}))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.RefreshSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "nil entity",
				ValkErrorCode: pam.ValkErrUndefined,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GenericPam{
				baseURL: tt.fields.baseURL,
				apiKey:  tt.fields.apiKey,
				rest:    tt.fields.rest,
			}
			tok, err := c.RefreshSession(tt.mapper)

			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			if tok != nil {
				assert.Equal(t, tt.want, tok)
			}
		})
	}
}

func TestGenericPam_GetBalance(t *testing.T) {
	var expectedRequest = pam.GetBalanceRequest{
		Params: pam.GetBalanceParams{
			Provider:     "prov",
			XPlayerToken: "token",
		},
		PlayerID: "foo",
	}
	var expectedErrorResponse = pam.BalanceResponse{
		Error: &pam.PamError{
			Code:    pam.PAMERRACCNOTFOUND,
			Message: "test",
		},
		Status: "ERROR",
	}
	var expectedResponse = pam.BalanceResponse{
		Balance: &pam.Balance{
			BonusAmount: pam.ZeroAmount,
			CashAmount:  pam.ZeroAmount,
		},
		Status: "OK",
	}
	type fields struct {
		baseURL string
		apiKey  string
		rest    rest.HTTPClientJSONInterface
	}
	tests := []struct {
		name    string
		fields  fields
		mapper  pam.GetBalanceRequestMapper
		wantErr error
		want    *pam.Balance
	}{
		{
			name: "successful get balance",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "base/players/foo/balance", req.URL)

					assert.Equal(t, expectedRequest.Params.XPlayerToken, req.Headers["X-Player-Token"])
					assert.Equal(t, "Bearer key", req.Headers["Authorization"])

					assert.Equal(t, expectedRequest.Params.Provider, req.Query["provider"])

					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetBalanceRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			want: expectedResponse.Balance,
		},
		{
			name: "error get balance",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					return assert.AnError
				}},
			},
			mapper: func() (context.Context, pam.GetBalanceRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "http client error",
				ValkErrorCode: pam.ValkErrUndefined,
				OrigError:     assert.AnError,
			},
		},
		{
			name: "error get balance mapper",
			fields: fields{
				"base",
				"key",
				nil,
			},
			mapper: func() (context.Context, pam.GetBalanceRequest, error) {
				return context.Background(), pam.GetBalanceRequest{}, assert.AnError
			},
			wantErr: assert.AnError,
		},
		{
			name: "error get balance response body",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedErrorResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetBalanceRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "PAM_ERR_ACC_NOT_FOUND test",
				ValkErrorCode: pam.ValkErrOpAccountNotFound,
				OrigError:     expectedErrorResponse.Error,
			},
		},
		{
			name: "error get balance response body nil balance",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(pam.BalanceResponse{}))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetBalanceRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "nil entity",
				ValkErrorCode: pam.ValkErrUndefined,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GenericPam{
				baseURL: tt.fields.baseURL,
				apiKey:  tt.fields.apiKey,
				rest:    tt.fields.rest,
			}
			ar, err := c.GetBalance(tt.mapper)

			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			if ar != nil {
				assert.Equal(t, tt.want, ar)
			}
		})
	}
}

func TestGenericPam_GetTransactions(t *testing.T) {
	var providerTransactionID = "123"
	var providerBetRef = "321"
	var expectedRequestBetRef = pam.GetTransactionsRequest{
		PlayerID: "1",
		Params: pam.GetTransactionsParams{
			Provider:       "prov",
			XPlayerToken:   "token",
			ProviderBetRef: &providerBetRef,
		},
	}
	var expectedRequestTransactionID = pam.GetTransactionsRequest{
		PlayerID: "1",
		Params: pam.GetTransactionsParams{
			Provider:              "prov",
			XPlayerToken:          "token",
			ProviderTransactionId: &providerTransactionID,
		},
	}
	var expectedErrorResponse = pam.GetTransactionsResponse{
		Error: &pam.PamError{
			Code:    pam.PAMERRTRANSNOTFOUND,
			Message: "test",
		},
		Status: "ERROR",
	}
	var expectedResponse = pam.GetTransactionsResponse{
		Transactions: &[]pam.Transaction{
			{
				BonusAmount:           pam.ZeroAmount,
				CashAmount:            pam.ZeroAmount,
				Currency:              "SEK",
				ProviderBetRef:        &providerBetRef,
				ProviderTransactionId: providerTransactionID,
				TransactionDateTime:   pam.Timestamp{},
				TransactionType:       "DEPOSIT",
			},
		},
		Status: "OK",
	}
	var expectedEmptyResponse = pam.GetTransactionsResponse{
		Transactions: &[]pam.Transaction{},
		Status:       "OK",
	}

	type fields struct {
		baseURL string
		apiKey  string
		rest    rest.HTTPClientJSONInterface
	}
	tests := []struct {
		name    string
		fields  fields
		mapper  pam.GetTransactionsRequestMapper
		wantErr error
		want    []pam.Transaction
	}{
		{
			name: "successful get transaction by providerBetRef",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "base/players/1/transactions", req.URL)

					assert.Equal(t, expectedRequestBetRef.Params.XPlayerToken, req.Headers["X-Player-Token"])
					assert.Equal(t, "Bearer key", req.Headers["Authorization"])

					assert.Equal(t, expectedRequestBetRef.Params.Provider, req.Query["provider"])
					assert.Equal(t, *expectedRequestBetRef.Params.ProviderBetRef, req.Query["providerBetRef"])

					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetTransactionsRequest, error) {
				return context.Background(), expectedRequestBetRef, nil
			},
			want: *expectedResponse.Transactions,
		},
		{
			name: "successful get transaction by providerTransactionId",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "base/players/1/transactions", req.URL)

					assert.Equal(t, expectedRequestTransactionID.Params.XPlayerToken, req.Headers["X-Player-Token"])
					assert.Equal(t, "Bearer key", req.Headers["Authorization"])

					assert.Equal(t, expectedRequestTransactionID.Params.Provider, req.Query["provider"])
					assert.Equal(t, *expectedRequestTransactionID.Params.ProviderTransactionId, req.Query["providerTransactionId"])

					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetTransactionsRequest, error) {
				return context.Background(), expectedRequestTransactionID, nil
			},
			want: *expectedResponse.Transactions,
		},
		{
			name: "error get transaction empty",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedEmptyResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetTransactionsRequest, error) {
				return context.Background(), expectedRequestTransactionID, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "No transactions",
				ValkErrorCode: pam.ValkErrOpTransNotFound,
			},
		},
		{
			name: "error get transaction",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					return assert.AnError
				}},
			},
			mapper: func() (context.Context, pam.GetTransactionsRequest, error) {
				return context.Background(), expectedRequestBetRef, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "http client error",
				ValkErrorCode: pam.ValkErrUndefined,
				OrigError:     assert.AnError,
			},
		},
		{
			name: "error get transaction mapper",
			fields: fields{
				"base",
				"key",
				nil,
			},
			mapper: func() (context.Context, pam.GetTransactionsRequest, error) {
				return context.Background(), pam.GetTransactionsRequest{}, assert.AnError
			},
			wantErr: assert.AnError,
		},
		{
			name: "error get transaction response body",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedErrorResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetTransactionsRequest, error) {
				return context.Background(), expectedRequestBetRef, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "PAM_ERR_TRANS_NOT_FOUND test",
				ValkErrorCode: pam.ValkErrOpTransNotFound,
				OrigError:     expectedErrorResponse.Error,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GenericPam{
				baseURL: tt.fields.baseURL,
				apiKey:  tt.fields.apiKey,
				rest:    tt.fields.rest,
			}
			tr, err := c.GetTransactions(tt.mapper)

			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			if tr != nil {
				assert.Equal(t, tt.want, tr)
			}
		})
	}
}

func TestGenericPam_AddTransaction(t *testing.T) {
	var transactionID = "123"
	var expectedRequest = &pam.AddTransactionRequest{
		PlayerID: "1",
		Params: pam.AddTransactionParams{
			Provider:     "prov",
			XPlayerToken: "token",
		},
		Body: pam.AddTransactionJSONRequestBody{
			BonusAmount:         pam.ZeroAmount,
			CashAmount:          testutils.NewFloatAmount(1),
			Currency:            "SEK",
			TransactionDateTime: pam.Timestamp{},
			TransactionType:     "DEPOSIT",
		},
	}
	var expectedErrorResponse = pam.AddTransactionResponse{
		Error: &pam.PamError{
			Code:    pam.PAMERRTRANSNOTFOUND,
			Message: "test",
		},
		Status: "ERROR",
	}
	var expectedErrorBalanceResponse = pam.AddTransactionResponse{
		Error: &pam.PamError{
			Code:    pam.PAMERRTRANSNOTFOUND,
			Message: "test",
		},
		Status: "ERROR",
		TransactionResult: &pam.TransactionResult{
			Balance: &pam.Balance{},
		},
	}
	var expectedResponse = pam.AddTransactionResponse{
		Status: "OK",
		TransactionResult: &pam.TransactionResult{
			Balance:       &pam.Balance{},
			TransactionId: &transactionID,
		},
	}

	type fields struct {
		baseURL string
		apiKey  string
		rest    rest.HTTPClientJSONInterface
	}
	tests := []struct {
		name    string
		fields  fields
		mapper  pam.AddTransactionRequestMapper
		wantErr error
		want    *string
	}{
		{
			name: "successful post transaction",
			fields: fields{
				"base",
				"key",
				mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "base/players/1/transactions", req.URL)

					assert.Equal(t, expectedRequest.Params.XPlayerToken, req.Headers["X-Player-Token"])
					assert.Equal(t, "Bearer key", req.Headers["Authorization"])

					assert.Equal(t, expectedRequest.Params.Provider, req.Query["provider"])

					assert.Equal(t, &expectedRequest.Body, req.Body)

					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			mapper: func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			want: &transactionID,
		},
		{
			name: "error post transaction",
			fields: fields{
				"base",
				"key",
				mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					return assert.AnError
				}},
			},
			mapper: func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "http client error",
				ValkErrorCode: pam.ValkErrUndefined,
				OrigError:     assert.AnError,
			},
		},
		{
			name: "error post transaction mapper",
			fields: fields{
				"base",
				"key",
				nil,
			},
			mapper: func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
				return context.Background(), &pam.AddTransactionRequest{}, assert.AnError
			},
			wantErr: assert.AnError,
		},
		{
			name: "error post transaction response body",
			fields: fields{
				"base",
				"key",
				mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedErrorResponse))
					return nil
				}},
			},
			mapper: func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "PAM_ERR_TRANS_NOT_FOUND test",
				ValkErrorCode: pam.ValkErrOpTransNotFound,
				OrigError:     expectedErrorResponse.Error,
			},
		},
		{
			name: "error post transaction response body nil result",
			fields: fields{
				"base",
				"key",
				mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(pam.AddTransactionResponse{}))
					return nil
				}},
			},
			mapper: func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "nil entity",
				ValkErrorCode: pam.ValkErrUndefined,
			},
		},
		{
			name: "error post transaction response body with balance",
			fields: fields{
				"base",
				"key",
				mockClient{PostJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedErrorBalanceResponse))
					return nil
				}},
			},
			mapper: func(pam.AmountRounder) (context.Context, *pam.AddTransactionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "PAM_ERR_TRANS_NOT_FOUND test",
				ValkErrorCode: pam.ValkErrOpTransNotFound,
				OrigError:     expectedErrorBalanceResponse.Error,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GenericPam{
				baseURL: tt.fields.baseURL,
				apiKey:  tt.fields.apiKey,
				rest:    tt.fields.rest,
			}
			tResponse, err := c.AddTransaction(tt.mapper)

			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			if tResponse != nil && tResponse.TransactionId != nil {
				assert.Equal(t, transactionID, *tResponse.TransactionId)
			}
		})
	}
}

func TestGenericPam_GetGameRound(t *testing.T) {
	var providerRoundID = "123"
	var expectedRequest = pam.GetGameRoundRequest{
		PlayerID:        "1",
		ProviderRoundID: providerRoundID,
		Params: pam.GetGameRoundParams{
			Provider:     "prov",
			XPlayerToken: "token",
		},
	}
	var expectedErrorResponse = pam.GameRoundResponse{
		Error: &pam.PamError{
			Code:    pam.PAMERRROUNDNOTFOUND,
			Message: "test",
		},
		Status: "ERROR",
	}
	var expectedResponse = pam.GameRoundResponse{
		Gameround: &pam.GameRound{
			EndTime:         nil,
			ProviderGameId:  "game",
			ProviderRoundId: providerRoundID,
			StartTime:       pam.Timestamp{},
		},
		Status: "OK",
	}

	type fields struct {
		baseURL string
		apiKey  string
		rest    rest.HTTPClientJSONInterface
	}
	tests := []struct {
		name    string
		fields  fields
		mapper  pam.GetGameRoundRequestMapper
		wantErr error
		want    pam.GameRound
	}{
		{
			name: "successful get gameround",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "base/players/1/gamerounds/"+expectedResponse.Gameround.ProviderRoundId, req.URL)

					assert.Equal(t, expectedRequest.Params.XPlayerToken, req.Headers["X-Player-Token"])
					assert.Equal(t, "Bearer key", req.Headers["Authorization"])

					assert.Equal(t, expectedRequest.Params.Provider, req.Query["provider"])

					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetGameRoundRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			want: *expectedResponse.Gameround,
		},
		{
			name: "error get gameround",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					return assert.AnError
				}},
			},
			mapper: func() (context.Context, pam.GetGameRoundRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "http client error",
				ValkErrorCode: pam.ValkErrUndefined,
				OrigError:     assert.AnError,
			},
		},
		{
			name: "error get gameround mapper",
			fields: fields{
				"base",
				"key",
				nil,
			},
			mapper: func() (context.Context, pam.GetGameRoundRequest, error) {
				return context.Background(), pam.GetGameRoundRequest{}, assert.AnError
			},
			wantErr: assert.AnError,
		},
		{
			name: "error get gameround response body",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedErrorResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetGameRoundRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "PAM_ERR_ROUND_NOT_FOUND test",
				ValkErrorCode: pam.ValkErrOpRoundNotFound,
				OrigError:     expectedErrorResponse.Error,
			},
		},
		{
			name: "error get gameround response body nil gameround",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(pam.GameRoundResponse{}))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetGameRoundRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "nil entity",
				ValkErrorCode: pam.ValkErrUndefined,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GenericPam{
				baseURL: tt.fields.baseURL,
				apiKey:  tt.fields.apiKey,
				rest:    tt.fields.rest,
			}
			g, err := c.GetGameRound(tt.mapper)

			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			if g != nil {
				assert.Equal(t, tt.want, *g)
			}
		})
	}
}

func TestGenericPam_GetSession(t *testing.T) {
	var expectedRequest = pam.GetSessionRequest{
		Params: pam.GetSessionParams{
			Provider:     "prov",
			XPlayerToken: "token",
		},
	}
	var expectedErrorResponse = pam.SessionResponse{
		Error: &pam.PamError{
			Code:    pam.PAMERRSESSIONNOTFOUND,
			Message: "test",
		},
		Status: "ERROR",
	}
	var expectedResponse = pam.SessionResponse{
		Session: &pam.Session{
			Country:  "SE",
			Currency: "SEK",
			Language: "sv",
			PlayerId: "123",
			Token:    "token",
		},
		Status: "OK",
	}

	type fields struct {
		baseURL string
		apiKey  string
		rest    rest.HTTPClientJSONInterface
	}
	tests := []struct {
		name    string
		fields  fields
		mapper  pam.GetSessionRequestMapper
		wantErr error
		want    pam.Session
	}{
		{
			name: "successful get session",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					assert.Equal(t, "base/players/session", req.URL)

					assert.Equal(t, expectedRequest.Params.XPlayerToken, req.Headers["X-Player-Token"])
					assert.Equal(t, "Bearer key", req.Headers["Authorization"])

					assert.Equal(t, expectedRequest.Params.Provider, req.Query["provider"])

					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			want: *expectedResponse.Session,
		},
		{
			name: "error get session",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					return assert.AnError
				}},
			},
			mapper: func() (context.Context, pam.GetSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "http client error",
				ValkErrorCode: pam.ValkErrUndefined,
				OrigError:     assert.AnError,
			},
		},
		{
			name: "error get session mapper",
			fields: fields{
				"base",
				"key",
				nil,
			},
			mapper: func() (context.Context, pam.GetSessionRequest, error) {
				return context.Background(), pam.GetSessionRequest{}, assert.AnError
			},
			wantErr: assert.AnError,
		},
		{
			name: "error get session response body",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(expectedErrorResponse))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "PAM_ERR_SESSION_NOT_FOUND test",
				ValkErrorCode: pam.ValkErrOpSessionNotFound,
				OrigError:     expectedErrorResponse.Error,
			},
		},
		{
			name: "error get session response body nil",
			fields: fields{
				"base",
				"key",
				mockClient{GetJSONFunc: func(ctx context.Context, req *rest.HTTPRequest, resp any) error {
					reflect.ValueOf(resp).
						Elem().
						Set(reflect.ValueOf(pam.SessionResponse{}))
					return nil
				}},
			},
			mapper: func() (context.Context, pam.GetSessionRequest, error) {
				return context.Background(), expectedRequest, nil
			},
			wantErr: pam.ValkyrieError{
				ErrMsg:        "nil entity",
				ValkErrorCode: pam.ValkErrUndefined,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GenericPam{
				baseURL: tt.fields.baseURL,
				apiKey:  tt.fields.apiKey,
				rest:    tt.fields.rest,
			}
			session, err := c.GetSession(tt.mapper)

			if err != nil {
				assert.Equal(t, tt.wantErr, err)
			}
			if session != nil {
				assert.Equal(t, tt.want, *session)
			}
		})
	}
}
