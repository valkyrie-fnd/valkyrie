package caleta

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

var hundredFiftyAndFifty = 15050000
var sekConst = SEK

func Test_Balance(t *testing.T) {
	playerID := "luke"
	tests := []struct {
		name      string
		sessionFn func() (*pam.Session, error)
		balanceFn func() (*pam.Balance, error)
		request   WalletbalanceRequestObject
		want      WalletbalanceResponseObject
		wantErr   error
	}{
		{
			"Return error from session",
			func() (*pam.Session, error) { return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrAuth} },
			nil,
			WalletbalanceRequestObject{
				Body: &WalletBalanceBody{
					Token: "123",
				},
			},
			Walletbalance200JSONResponse{
				Status: RSERRORINVALIDTOKEN,
			},
			nil,
		},
		{
			"Return error from Balance",
			func() (*pam.Session, error) { return &pam.Session{}, nil },
			func() (*pam.Balance, error) { return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrUndefined} },
			WalletbalanceRequestObject{
				Body: &WalletBalanceBody{
					Token: "123",
				},
			},
			Walletbalance200JSONResponse{
				Status: RSERRORUNKNOWN,
			},
			nil,
		},
		{
			"Return Balance response",
			func() (*pam.Session, error) { return &pam.Session{Currency: "SEK"}, nil },
			func() (*pam.Balance, error) { return &pam.Balance{CashAmount: testutils.NewFloatAmount(150.5)}, nil },
			WalletbalanceRequestObject{
				Body: &WalletBalanceBody{
					RequestUuid:  "123abc",
					Token:        "123",
					SupplierUser: playerID,
				},
			},
			Walletbalance200JSONResponse{
				Status:      RSOK,
				Balance:     &hundredFiftyAndFifty,
				Currency:    &sekConst,
				RequestUuid: "123abc",
				User:        &playerID,
			},
			nil,
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamstub := pamStub{}
			pamstub.sessionFn = test.sessionFn
			pamstub.balanceFn = test.balanceFn
			sut := NewWalletService(&pamstub, nil)

			resp, err := sut.Walletbalance(ctx, test.request)
			assert.Equal(tt, test.want, resp)
			assert.Equal(tt, test.wantErr, err)
		})
	}
}
func Test_Check(t *testing.T) {
	newToken := "newToken123"
	tests := []struct {
		name      string
		refreshFn func() (*pam.Session, error)
		request   WalletcheckRequestObject
		want      WalletcheckResponseObject
		wantErr   error
	}{
		{
			"Return error when call refreshsession fails",
			func() (*pam.Session, error) { return nil, fmt.Errorf("refresh error") },
			WalletcheckRequestObject{
				Body: &WalletCheckBody{
					Token: "123",
				},
			},
			nil,
			fmt.Errorf("refresh error"),
		},
		{
			"returns updated token",
			func() (*pam.Session, error) { return &pam.Session{Token: "newToken123"}, nil },
			WalletcheckRequestObject{
				Body: &WalletCheckBody{
					Token: "123",
				},
			},
			Walletcheck200JSONResponse{
				Token: &newToken,
			},
			nil,
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamstub := pamStub{}
			pamstub.refreshSessionFn = test.refreshFn
			sut := NewWalletService(&pamstub, nil)

			resp, err := sut.Walletcheck(ctx, test.request)
			assert.Equal(tt, test.want, resp)
			assert.Equal(tt, test.wantErr, err)
		})
	}
}

func Test_Bet(t *testing.T) {
	playerID := "Player1"
	tests := []struct {
		name                string
		sessionFn           func() (*pam.Session, error)
		addTranFn           func() (*pam.TransactionResult, error)
		transactionHandling pam.TransactionHandling
		request             WalletbetRequestObject
		want                WalletbetResponseObject
	}{
		{
			"Return error from session",
			func() (*pam.Session, error) { return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrAuth} },
			nil,
			pam.PROVIDER,
			WalletbetRequestObject{
				Body: &WalletBetBody{
					Token: "123",
				},
			},
			Walletbet200JSONResponse{
				Status: RSERRORINVALIDTOKEN,
			},
		},
		{
			"Return error from addTransaction",
			func() (*pam.Session, error) { return &pam.Session{}, nil },
			func() (*pam.TransactionResult, error) {
				return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrOpBetNotAllowed}
			},
			pam.PROVIDER,
			WalletbetRequestObject{
				Body: &WalletBetBody{
					Token: "123",
				},
			},
			Walletbet200JSONResponse{
				Status: RSERRORUSERDISABLED,
			},
		},
		{
			"Return RS_ERROR_INVALID_GAME from ValkErrOpRoundNotFound error in addTransaction",
			func() (*pam.Session, error) { return &pam.Session{}, nil },
			func() (*pam.TransactionResult, error) {
				return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrOpRoundNotFound}
			},
			pam.PROVIDER,
			WalletbetRequestObject{
				Body: &WalletBetBody{
					Token: "123",
				},
			},
			Walletbet200JSONResponse{
				Status: RSERRORINVALIDGAME,
			},
		},
		{
			"Return balance response when successful",
			func() (*pam.Session, error) { return &pam.Session{Currency: "SEK"}, nil },
			func() (*pam.TransactionResult, error) {
				tranID := "tranId"
				return &pam.TransactionResult{
					Balance:       &pam.Balance{CashAmount: testutils.NewFloatAmount(150.5)},
					TransactionId: &tranID}, nil
			},
			pam.PROVIDER,
			WalletbetRequestObject{
				Body: &WalletBetBody{
					Token:        "123",
					RequestUuid:  "uuid123",
					SupplierUser: playerID,
				},
			},
			Walletbet200JSONResponse{
				Status:      RSOK,
				Balance:     &hundredFiftyAndFifty,
				Currency:    &sekConst,
				RequestUuid: "uuid123",
				User:        &playerID,
			},
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamstub := pamStub{}
			pamstub.sessionFn = test.sessionFn
			pamstub.addTransFn = test.addTranFn
			pamstub.getTransactionHandlingFn = func() pam.TransactionHandling { return test.transactionHandling }
			sut := NewWalletService(&pamstub, nil)

			resp, err := sut.Walletbet(ctx, test.request)
			assert.Nil(tt, err, "Error should always be nil")
			assert.Equal(tt, test.want, resp)
		})
	}
}

func Test_Win(t *testing.T) {
	playerID := "Player2"
	tests := []struct {
		name                string
		sessionFn           func() (*pam.Session, error)
		addTranFn           func() (*pam.TransactionResult, error)
		transactionHandling pam.TransactionHandling
		request             TransactionwinRequestObject
		want                TransactionwinResponseObject
	}{
		{
			"Return error from session",
			func() (*pam.Session, error) { return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrAuth} },
			nil,
			pam.PROVIDER,
			TransactionwinRequestObject{
				Body: &WalletWinBody{
					Token: "123",
				},
			},
			Transactionwin200JSONResponse{
				Status: RSERRORINVALIDTOKEN,
			},
		},
		{
			"Return error from addTransaction",
			func() (*pam.Session, error) { return &pam.Session{}, nil },
			func() (*pam.TransactionResult, error) {
				return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrOpBetNotAllowed}
			},
			pam.PROVIDER,
			TransactionwinRequestObject{
				Body: &WalletWinBody{
					Token: "123",
				},
			},
			Transactionwin200JSONResponse{
				Status: RSERRORUSERDISABLED,
			},
		},
		{
			"Return balance response when successful",
			func() (*pam.Session, error) { return &pam.Session{Currency: "SEK"}, nil },
			func() (*pam.TransactionResult, error) {
				tranID := "tranId"
				return &pam.TransactionResult{
					Balance:       &pam.Balance{CashAmount: testutils.NewFloatAmount(150.5)},
					TransactionId: &tranID}, nil
			},
			pam.PROVIDER,
			TransactionwinRequestObject{
				Body: &WalletWinBody{
					Token:        "123",
					RequestUuid:  "uuid123",
					SupplierUser: playerID,
				},
			},
			Transactionwin200JSONResponse{
				Status:      RSOK,
				Balance:     &hundredFiftyAndFifty,
				Currency:    &sekConst,
				RequestUuid: "uuid123",
				User:        &playerID,
			},
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamstub := pamStub{}
			pamstub.sessionFn = test.sessionFn
			pamstub.addTransFn = test.addTranFn
			pamstub.getTransactionHandlingFn = func() pam.TransactionHandling { return test.transactionHandling }
			api := &mockAPIClient{getRoundTransactionsFn: func(ctx context.Context, gameRoundID string) (*transactionResponse, error) {
				return &transactionResponse{}, nil
			}}
			sut := NewWalletService(&pamstub, api)

			resp, err := sut.Transactionwin(ctx, test.request)
			assert.Nil(tt, err, "Error should always be nil")
			assert.Equal(tt, test.want, resp)
		})
	}
}

func Test_Rollback(t *testing.T) {
	playerID := "Player3"
	tests := []struct {
		name                string
		sessionFn           func() (*pam.Session, error)
		addTranFn           func() (*pam.TransactionResult, error)
		transactionHandling pam.TransactionHandling
		request             WalletrollbackRequestObject
		want                WalletrollbackResponseObject
	}{
		{
			"Return error from session",
			func() (*pam.Session, error) { return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrAuth} },
			nil,
			pam.PROVIDER,
			WalletrollbackRequestObject{
				Body: &WalletRollbackBody{
					Token: "123",
				},
			},
			Walletrollback200JSONResponse{
				Status: RSERRORINVALIDTOKEN,
			},
		},
		{
			"Return OK when error from addTransaction",
			func() (*pam.Session, error) { return &pam.Session{Currency: "SEK"}, nil },
			func() (*pam.TransactionResult, error) {
				return &pam.TransactionResult{Balance: &pam.Balance{
					CashAmount: testutils.NewFloatAmount(150.5),
				}}, pam.ValkyrieError{ValkErrorCode: pam.ValkErrOpBetNotAllowed}
			},
			pam.PROVIDER,
			WalletrollbackRequestObject{
				Body: &WalletRollbackBody{
					Token:       "123",
					RequestUuid: "uuid123",
					User:        &playerID,
				},
			},
			Walletrollback200JSONResponse{
				Status:      RSOK,
				Balance:     &hundredFiftyAndFifty,
				Currency:    &sekConst,
				RequestUuid: "uuid123",
				User:        &playerID,
			},
		},
		{
			"Return OK when error and nil balance from addTransaction",
			func() (*pam.Session, error) { return &pam.Session{Currency: "SEK"}, nil },
			func() (*pam.TransactionResult, error) {
				return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrOpBetNotAllowed}
			},
			pam.PROVIDER,
			WalletrollbackRequestObject{
				Body: &WalletRollbackBody{
					Token:       "123",
					RequestUuid: "uuid123",
					User:        &playerID,
				},
			},
			Walletrollback200JSONResponse{
				Status:      RSOK,
				Balance:     testutils.Ptr(0),
				Currency:    &sekConst,
				RequestUuid: "uuid123",
				User:        &playerID,
			},
		},
		{
			"Return balance response when successful",
			func() (*pam.Session, error) { return &pam.Session{Currency: "SEK"}, nil },
			func() (*pam.TransactionResult, error) {
				tranID := "tranId"
				return &pam.TransactionResult{
					Balance:       &pam.Balance{CashAmount: testutils.NewFloatAmount(150.5)},
					TransactionId: &tranID}, nil
			},
			pam.PROVIDER,
			WalletrollbackRequestObject{
				Body: &WalletRollbackBody{
					Token:       "123",
					RequestUuid: "uuid123",
					User:        &playerID,
				},
			},
			Walletrollback200JSONResponse{
				Status:      RSOK,
				Balance:     &hundredFiftyAndFifty,
				Currency:    &sekConst,
				RequestUuid: "uuid123",
				User:        &playerID,
			},
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamstub := pamStub{getTransactionHandlingFn: func() pam.TransactionHandling { return test.transactionHandling }}
			pamstub.sessionFn = test.sessionFn
			pamstub.addTransFn = test.addTranFn
			api := &mockAPIClient{getRoundTransactionsFn: func(ctx context.Context, gameRoundID string) (*transactionResponse, error) {
				return &transactionResponse{}, nil
			}}
			sut := NewWalletService(&pamstub, api)

			resp, err := sut.Walletrollback(ctx, test.request)
			assert.Nil(tt, err, "Error should always be nil")
			assert.Equal(tt, test.want, resp)
		})
	}
}

func Test_getRoundTransactions(t *testing.T) {
	tests := []struct {
		name                   string
		transactionHandling    pam.TransactionHandling
		round                  string
		getRoundTransactionsFn func(ctx context.Context, gameRoundID string) (*transactionResponse, error)
		want                   *[]roundTransaction
	}{
		{
			name:                "Operator transaction handling should not return any transactions",
			transactionHandling: pam.OPERATOR,
			round:               "909",
		},
		{
			name:                "Provider transaction should return round transactions",
			transactionHandling: pam.PROVIDER,
			round:               "909",
			getRoundTransactionsFn: func(ctx context.Context, gameRoundID string) (*transactionResponse, error) {
				id, _ := strconv.Atoi(gameRoundID)
				return &transactionResponse{
					RoundID: gameRoundID,
					RoundTransactions: &[]roundTransaction{
						{
							RoundID: id,
							TxnUUID: "txn-uuid",
						},
					},
				}, nil
			},
			want: &[]roundTransaction{
				{
					RoundID: 909,
					TxnUUID: "txn-uuid",
				},
			},
		},
		{
			name:                "Failed getting transactions",
			transactionHandling: pam.PROVIDER,
			round:               "909",
			getRoundTransactionsFn: func(ctx context.Context, gameRoundID string) (*transactionResponse, error) {
				return nil, fmt.Errorf("Error fetching transactions")
			},
			want: nil,
		},
		{
			name:                "Provider transaction handling but no transactions found",
			transactionHandling: pam.PROVIDER,
			round:               "909",
			getRoundTransactionsFn: func(ctx context.Context, gameRoundID string) (*transactionResponse, error) {
				return &transactionResponse{Code: 1016, Message: "Invalid Round"}, nil
			},
			want: (*[]roundTransaction)(nil),
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamstub := pamStub{getTransactionHandlingFn: func() pam.TransactionHandling { return test.transactionHandling }}
			api := &mockAPIClient{getRoundTransactionsFn: test.getRoundTransactionsFn}
			service := NewWalletService(&pamstub, api)
			resp := service.getRoundTransactions(ctx, test.transactionHandling, test.round)
			assert.Equal(tt, test.want, resp)
		})
	}

}

type pamStub struct {
	pam.PamClient
	balanceFn                func() (*pam.Balance, error)
	refreshSessionFn         func() (*pam.Session, error)
	sessionFn                func() (*pam.Session, error)
	addTransFn               func() (*pam.TransactionResult, error)
	getTransFn               func() ([]pam.Transaction, error)
	getGameRoundFn           func() (*pam.GameRound, error)
	getTransactionHandlingFn func() pam.TransactionHandling
}

func (pam *pamStub) RefreshSession(_ pam.RefreshSessionRequestMapper) (*pam.Session, error) {
	return pam.refreshSessionFn()
}

func (pam *pamStub) GetSession(_ pam.GetSessionRequestMapper) (*pam.Session, error) {
	return pam.sessionFn()
}

func (pam *pamStub) GetBalance(_ pam.GetBalanceRequestMapper) (*pam.Balance, error) {
	return pam.balanceFn()
}

func (pam *pamStub) GetTransactions(_ pam.GetTransactionsRequestMapper) ([]pam.Transaction, error) {
	return pam.getTransFn()
}

func (pam *pamStub) AddTransaction(_ pam.AddTransactionRequestMapper) (*pam.TransactionResult, error) {
	return pam.addTransFn()
}

func (pam *pamStub) GetGameRound(_ pam.GetGameRoundRequestMapper) (*pam.GameRound, error) {
	return pam.getGameRoundFn()
}

func (pam *pamStub) GetTransactionHandling() pam.TransactionHandling {
	return pam.getTransactionHandlingFn()
}
