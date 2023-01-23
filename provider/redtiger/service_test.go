package redtiger

import (
	"testing"
	"time"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

var (
	country  = "SE"
	language = "sv"
)

func TestAuth(t *testing.T) {
	tests := []struct {
		name      string
		refreshFn func() (*pam.Session, error)
		balanceFn func() (*pam.Balance, error)
		req       AuthRequest
		want      *AuthResponseWrapper
		wantErr   *ErrorResponse
	}{
		{
			"Return error if refreshing of session fails",
			func() (*pam.Session, error) {
				return nil, pam.ValkyrieError{
					ErrMsg:        "fail",
					ValkErrorCode: pam.ValkErrAuth,
				}
			},
			func() (*pam.Balance, error) {
				return &pam.Balance{
					BonusAmount: pam.ZeroAmount,
					CashAmount:  testutils.NewFloatAmount(666),
				}, nil
			},
			AuthRequest{},
			nil,
			&ErrorResponse{Success: false, Error: Error{
				Message: "Not authorized",
				Code:    301,
			}},
		},
		{
			"Return error if GetBalance fails",
			func() (*pam.Session, error) {
				return &pam.Session{
					Country:  country,
					Currency: "SEK",
					Language: language,
					PlayerId: "999",
					Token:    "token",
				}, nil
			},
			func() (*pam.Balance, error) {
				return nil, pam.ValkyrieError{
					ErrMsg:        "fail",
					ValkErrorCode: pam.ValkErrUndefined,
				}
			},
			AuthRequest{},
			nil,
			&ErrorResponse{Success: false, Error: Error{
				Message: "Failed to Auth: Code: 0, Msg: fail",
				Code:    201,
			}},
		},
		{
			"Return success if all goes well",
			func() (*pam.Session, error) {
				return &pam.Session{
					Country:  country,
					Currency: "SEK",
					Language: language,
					PlayerId: "999",
					Token:    "abc123",
				}, nil
			},
			func() (*pam.Balance, error) {
				return &pam.Balance{
					BonusAmount: pam.ZeroAmount,
					CashAmount:  testutils.NewFloatAmount(666),
				}, nil
			},
			AuthRequest{
				BaseRequest: BaseRequest{
					UserID: "999",
					Casino: "kasino",
				},
			},
			&AuthResponseWrapper{
				Success: true,
				Result: AuthResponse{
					BaseResponse: BaseResponse{
						Token:    "abc123",
						Currency: "SEK",
					},
					UserID:   "999",
					Casino:   "kasino",
					Country:  country,
					Language: language,
					Balance: Balance{
						Cash:  Money(testutils.NewFloatAmount(666)),
						Bonus: zeroMoney(),
					},
				},
			},
			nil,
		},
		{
			"Return success even if request does not have UserId",
			func() (*pam.Session, error) {
				return &pam.Session{
					Country:  country,
					Currency: "SEK",
					Language: language,
					PlayerId: "999",
					Token:    "abc123",
				}, nil
			},
			func() (*pam.Balance, error) {
				return &pam.Balance{
					BonusAmount: pam.ZeroAmount,
					CashAmount:  testutils.NewFloatAmount(666),
				}, nil
			},
			AuthRequest{
				BaseRequest: BaseRequest{
					Casino: "kasino",
				},
			},
			&AuthResponseWrapper{
				Success: true,
				Result: AuthResponse{
					BaseResponse: BaseResponse{
						Token:    "abc123",
						Currency: "SEK",
					},
					UserID:   "999",
					Casino:   "kasino",
					Country:  "SE",
					Language: "sv",
					Balance: Balance{
						Cash:  Money(testutils.NewFloatAmount(666)),
						Bonus: zeroMoney(),
					},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamStub := pamStub{}
			sut := NewService(&pamStub)
			pamStub.balanceFn = test.balanceFn
			pamStub.refreshSessionFn = test.refreshFn

			resp, err := sut.Auth(test.req)
			assert.Equal(tt, test.want, resp)
			assert.Equal(tt, test.wantErr, err)
		})
	}
}

func TestPayout(t *testing.T) {
	tests := []struct {
		name             string
		addTransactionFn func() (*pam.TransactionResult, error)
		balanceFn        func() (*pam.Balance, error)
		req              PayoutRequest
		want             *PayoutResponseWrapper
		wantErr          *ErrorResponse
	}{
		{
			"Return error if transaction fails",
			func() (*pam.TransactionResult, error) {
				return nil, pam.ValkyrieError{
					ErrMsg:        "Transaction Failed",
					ValkErrorCode: pam.ValkErrUndefined,
				}
			},
			nil,
			PayoutRequest{},
			nil,
			&ErrorResponse{
				Success: false,
				Error: Error{
					Message: "Code: 0, Msg: Transaction Failed",
					Code:    201,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamStub := pamStub{}
			sut := NewService(&pamStub)
			pamStub.balanceFn = test.balanceFn
			pamStub.addTransFn = test.addTransactionFn

			resp, err := sut.Payout(test.req)
			assert.Equal(tt, test.want, resp)
			assert.Equal(tt, test.wantErr, err)
		})
	}
}

func TestRefund(t *testing.T) {
	tests := []struct {
		name             string
		addTransactionFn func() (*pam.TransactionResult, error)
		balanceFn        func() (*pam.Balance, error)
		req              RefundRequest
		want             *RefundResponseWrapper
		wantErr          *ErrorResponse
	}{
		{
			"Return error if transaction fails",
			func() (*pam.TransactionResult, error) {
				return nil, pam.ValkyrieError{
					ErrMsg:        "Transaction Failed",
					ValkErrorCode: pam.ValkErrUndefined,
				}
			},
			nil,
			RefundRequest{},
			nil,
			&ErrorResponse{
				Success: false,
				Error: Error{
					Message: "Code: 0, Msg: Transaction Failed",
					Code:    201,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamStub := pamStub{}
			sut := NewService(&pamStub)
			pamStub.balanceFn = test.balanceFn
			pamStub.addTransFn = test.addTransactionFn

			resp, err := sut.Refund(test.req)
			assert.Equal(tt, test.want, resp)
			assert.Equal(tt, test.wantErr, err)
		})
	}
}

func TestStake(t *testing.T) {
	tests := []struct {
		name             string
		addTransactionFn func() (*pam.TransactionResult, error)
		balanceFn        func() (*pam.Balance, error)
		getTransactionFn func() ([]pam.Transaction, error)
		getGameRound     func() (*pam.GameRound, error)
		req              StakeRequest
		want             *StakeResponseWrapper
		wantErr          *ErrorResponse
	}{
		{
			"No previous transaction but gameround has ended cause error",
			nil,
			nil,
			func() ([]pam.Transaction, error) {
				return nil, pam.ValkyrieError{ErrMsg: "Could not find transaction", ValkErrorCode: pam.ValkErrOpTransNotFound}
			},
			func() (*pam.GameRound, error) {
				aMinuteAgo := time.Now().Add(-time.Minute)
				return &pam.GameRound{
					EndTime:         &aMinuteAgo,
					ProviderRoundId: "RoundXYZ",
				}, nil
			},
			StakeRequest{
				BaseRequest: BaseRequest{},
				Transaction: TransactionStake{
					ID: "TranId123",
				},
				Round: Round{
					ID: "RoundXYZ",
				},
			},
			nil,
			&ErrorResponse{
				Success: false,
				Error:   Error{Message: "Additional bet on game round not allowed", Code: 200},
			},
		},
		{
			"No Previous transaction exist, but gameround has not ended, works",
			func() (*pam.TransactionResult, error) {
				transID := "TranId123"
				balance := pam.Balance{
					CashAmount:  pam.ZeroAmount,
					BonusAmount: pam.ZeroAmount,
				}
				return &pam.TransactionResult{TransactionId: &transID, Balance: &balance}, nil
			},
			func() (*pam.Balance, error) {
				return &pam.Balance{
					BonusAmount: pam.ZeroAmount,
					CashAmount:  pam.ZeroAmount,
				}, nil
			},
			func() ([]pam.Transaction, error) {
				return nil, pam.ValkyrieError{ErrMsg: "Could not find transaction", ValkErrorCode: pam.ValkErrOpTransNotFound}
			},
			func() (*pam.GameRound, error) {
				return &pam.GameRound{
					ProviderRoundId: "RoundXYZ",
				}, nil
			},
			StakeRequest{
				BaseRequest: BaseRequest{},
				Transaction: TransactionStake{
					ID:    "TranId123",
					Stake: zeroMoney(),
				},
				Round: Round{
					ID: "RoundXYZ",
				},
			},
			&StakeResponseWrapper{
				Response: Response{
					Success: true,
				},
				Result: StakeResponse{
					BaseResponse: BaseResponse{},
					ID:           "TranId123",
					Stake: Balance{
						Cash:  zeroMoney(),
						Bonus: zeroMoney(),
					},
					Balance: Balance{
						Cash:  zeroMoney(),
						Bonus: zeroMoney(),
					},
				},
			},
			nil,
		},
		{
			"Transactions exist, works still to add new ones/The addTransaction is idempotent",
			func() (*pam.TransactionResult, error) {
				transID := "TranId123"
				balance := pam.Balance{
					CashAmount:  pam.ZeroAmount,
					BonusAmount: pam.ZeroAmount,
				}
				return &pam.TransactionResult{TransactionId: &transID, Balance: &balance}, nil
			},
			func() (*pam.Balance, error) {
				return &pam.Balance{}, nil
			},
			func() ([]pam.Transaction, error) {
				return []pam.Transaction{{
					ProviderTransactionId: "Doesn't have to be the same id",
				}}, nil
			},
			func() (*pam.GameRound, error) {
				return &pam.GameRound{
					ProviderRoundId: "RoundXYZ",
				}, nil
			},
			StakeRequest{
				BaseRequest: BaseRequest{},
				Transaction: TransactionStake{
					ID:    "TranId123",
					Stake: zeroMoney(),
				},
				Round: Round{
					ID: "RoundXYZ",
				},
			},
			&StakeResponseWrapper{
				Response: Response{
					Success: true,
				},
				Result: StakeResponse{
					BaseResponse: BaseResponse{},
					ID:           "TranId123",
					Stake: Balance{
						Cash:  zeroMoney(),
						Bonus: zeroMoney(),
					},
					Balance: Balance{
						Cash:  zeroMoney(),
						Bonus: zeroMoney(),
					},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			pamStub := pamStub{}
			sut := NewService(&pamStub)
			pamStub.balanceFn = test.balanceFn
			pamStub.getGameRoundFn = test.getGameRound
			pamStub.getTransFn = test.getTransactionFn
			pamStub.addTransFn = test.addTransactionFn

			resp, err := sut.Stake(test.req)
			assert.Equal(tt, test.want, resp)
			assert.Equal(tt, test.wantErr, err)
		})
	}
}

// Test_UnknownUserResponse verifies that potential IDOR issues are not propagated by Valkyrie. It's assumed
// that PAM's handle that properly but lets make sure it actually works. Ergo: unknown user issues should
// result in invalid session and nothing else.
func Test_UnknownUserResponse(t *testing.T) {
	pamStub := pamStub{}
	svc := NewService(&pamStub)
	testCases := []pam.ValkErrorCode{pam.ValkErrOpUserNotFound, pam.ValkErrOpSessionNotFound}
	for _, tt := range testCases {

		pamStub.refreshSessionFn = func() (*pam.Session, error) {
			return nil, pam.ErrorWrapper("a", tt, assert.AnError)
		}

		svc.Auth(AuthRequest{})
		_, err := svc.Auth(AuthRequest{})
		assert.NotNil(t, err)
		assert.Equalf(t, NotAuthorized, err.Error.Code, "expected NotAuthorized on %s", tt)
	}
}

type pamStub struct {
	pam.PamClient
	balanceFn        func() (*pam.Balance, error)
	refreshSessionFn func() (*pam.Session, error)
	sessionFn        func() (*pam.Session, error)
	addTransFn       func() (*pam.TransactionResult, error)
	getTransFn       func() ([]pam.Transaction, error)
	getGameRoundFn   func() (*pam.GameRound, error)
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
