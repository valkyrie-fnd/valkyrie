package evolution

import (
	"testing"

	"github.com/valkyrie-fnd/valkyrie/pam/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

var brokenBalanceFn = func() (*pam.Balance, error) { return nil, assert.AnError }
var brokenTransFn = func() ([]pam.Transaction, error) { return nil, assert.AnError }
var brokenAddTransFn = func() (*pam.TransactionResult, error) { return nil, assert.AnError }
var oneGrandBalanceFn = func() (*pam.Balance, error) {
	return &pam.Balance{
		CashAmount:  testutils.NewFloatAmount(1000),
		BonusAmount: pam.ZeroAmount,
	}, nil
}

func fetchTransFn(tt pam.TransactionType, ref string) func() ([]pam.Transaction, error) {
	return func() ([]pam.Transaction, error) {
		trans := pam.Transaction{
			TransactionType:       tt,
			CashAmount:            testutils.NewFloatAmount(11),
			ProviderBetRef:        &ref,
			ProviderTransactionId: "trans_1",
		}
		return []pam.Transaction{trans}, nil
	}
}

var newTransFn = func() (*pam.TransactionResult, error) {
	return &pam.TransactionResult{
		TransactionId: strPtr("trans2"),
		Balance:       &pam.Balance{CashAmount: testutils.NewFloatAmount(1000), BonusAmount: pam.ZeroAmount}}, nil
}

func TestProviderService_Debit(t *testing.T) {
	pamstub := pamStub{}
	svc := NewService(&pamstub)

	tests := []struct {
		name    string
		transFn func() (*pam.TransactionResult, error)
		want    *StandardResponse
		wantErr error
	}{
		{
			name:    "happy path",
			transFn: newTransFn,
			want: &StandardResponse{
				Status:  "OK",
				Balance: amountFromFloat(1000),
				Bonus:   ZeroAmount,
			},
			wantErr: nil,
		},
		{
			name:    "everything broken",
			transFn: brokenAddTransFn,
			wantErr: ProviderError{message: assert.AnError.Error()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pamstub.addTransFn = tt.transFn

			resp, err := svc.Debit(DebitRequest{
				RequestBase: RequestBase{
					UserID: "user_1",
					SID:    "sid_X",
				},
				Currency: "SEK",
				Transaction: Transaction{
					ID:     "ext_1",
					RefID:  "ref_A",
					Amount: amountFromFloat(10.01),
				},
				Game: Game{
					ID: "round_1",
					Details: GameDetails{
						Table: GameTable{
							ID: "BJ_1",
						},
					},
				},
			})

			assert.Equal(t, tt.want, resp)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &ProviderError{})
			}
		})
	}

}

func TestProviderService_Credit(t *testing.T) {
	type fields struct {
		getTransFn   func() ([]pam.Transaction, error)
		addTransFn   func() (*pam.TransactionResult, error)
		getBalanceFn func() (*pam.Balance, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    CreditRequest
		want    *StandardResponse
		wantErr *ProviderError
	}{
		{
			name: "happy case",
			fields: fields{
				getBalanceFn: oneGrandBalanceFn,
				getTransFn:   fetchTransFn(pam.WITHDRAW, "ref1"),
				addTransFn:   newTransFn,
			},
			want: &StandardResponse{
				Status:  "OK",
				Balance: amountFromFloat(1000),
				Bonus:   ZeroAmount,
			},
		},
		{
			name: "reject credit when bet(same extRefId) is missing",
			fields: fields{
				getBalanceFn: oneGrandBalanceFn,
				getTransFn: func() ([]pam.Transaction, error) {
					return nil, pam.ValkyrieError{
						ValkErrorCode: pam.ValkErrBetNotFound,
					}
				},
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusBetDoesNotExist.code,
					Balance: amountFromFloat(1000),
				},
			},
		},
		{
			name: "reject credit when another credit exists for the same reference",
			fields: fields{
				getBalanceFn: oneGrandBalanceFn,
				getTransFn:   fetchTransFn(pam.DEPOSIT, "ref1"),
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusBetAlreadySettled.code,
					Balance: amountFromFloat(1000),
				},
			},
		},
		{
			name: "reject credit when corresponding debit has been cancelled",
			fields: fields{
				getTransFn:   fetchTransFn(pam.CANCEL, "ref1"),
				getBalanceFn: oneGrandBalanceFn,
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusBetAlreadySettled.code,
					Balance: amountFromFloat(1000),
				},
			},
		},
		{
			name: "get transaction and get balance fails",
			fields: fields{
				getBalanceFn: brokenBalanceFn,
				getTransFn:   brokenTransFn,
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusUnknownError.code,
					Balance: ZeroAmount,
				},
			},
		},
		{
			name: "get transaction works but get balance fails",
			fields: fields{
				getBalanceFn: brokenBalanceFn,
				addTransFn:   newTransFn,
				getTransFn:   fetchTransFn(pam.CANCEL, "ref1"),
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusUnknownError.code,
					Balance: ZeroAmount,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare pam stub with fixed responses
			pamstub := pamStub{
				addTransFn: tt.fields.addTransFn,
				getTransFn: tt.fields.getTransFn,
				balanceFn:  tt.fields.getBalanceFn,
			}
			svc := NewService(&pamstub)

			pamstub.getTransFn = tt.fields.getTransFn

			// perform the actual request
			result, err := svc.Credit(CreditRequest{
				Currency: "EUR",
				Transaction: Transaction{
					RefID: "ref1",
					ID:    "trans2",
				},
			})

			if tt.wantErr != nil {
				var provErr ProviderError
				require.ErrorAs(t, err, &provErr, "expecting a provider error response")
				assert.Equal(t, tt.wantErr.response.Status, provErr.response.Status)
				assert.Equal(t, tt.wantErr.response.Balance, provErr.response.Balance)
			}

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestProviderService_Cancel(t *testing.T) {
	type fields struct {
		getTransFn   func() ([]pam.Transaction, error)
		getBalanceFn func() (*pam.Balance, error)
	}
	tests := []struct {
		name    string
		fields  fields
		want    *StandardResponse
		wantErr *ProviderError
	}{
		{
			name: "happy case",
			fields: fields{
				getTransFn:   fetchTransFn(pam.WITHDRAW, "ref1"),
				getBalanceFn: oneGrandBalanceFn,
			},
			want: &StandardResponse{
				Status:  "OK",
				Balance: amountFromFloat(1000),
				Bonus:   ZeroAmount,
			},
		},
		{
			name: "reject cancel when same extRefId is missing",
			fields: fields{
				getTransFn: func() ([]pam.Transaction, error) {
					return nil, pam.ValkyrieError{
						ValkErrorCode: pam.ValkErrBetNotFound,
					}
				},
				getBalanceFn: oneGrandBalanceFn,
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusBetDoesNotExist.code,
					Balance: amountFromFloat(1000),
				},
			},
		},
		{
			name: "reject cancel when another cancel exists for the same reference",
			fields: fields{
				getTransFn:   fetchTransFn(pam.CANCEL, "ref1"),
				getBalanceFn: oneGrandBalanceFn,
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusBetAlreadySettled.code,
					Balance: amountFromFloat(1000),
				},
			},
		},
		{
			name: "reject cancel when corresponding bet has been settled",
			fields: fields{
				getTransFn:   fetchTransFn(pam.DEPOSIT, "ref1"),
				getBalanceFn: oneGrandBalanceFn,
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusBetAlreadySettled.code,
					Balance: amountFromFloat(1000),
				},
			},
		},
		{
			name: "broken transaction and balance fetch",
			fields: fields{
				getTransFn:   brokenTransFn,
				getBalanceFn: brokenBalanceFn,
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusUnknownError.code,
					Balance: ZeroAmount,
				},
			},
		},
		{
			name: "get transaction works but get balance fails",
			fields: fields{
				getTransFn:   fetchTransFn(pam.DEPOSIT, "ref1"),
				getBalanceFn: brokenBalanceFn,
			},
			want: nil,
			wantErr: &ProviderError{
				response: &StandardResponse{
					Status:  StatusUnknownError.code,
					Balance: ZeroAmount,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare pam stub with fixed responses
			pamstub := pamStub{}
			pamstub.balanceFn = tt.fields.getBalanceFn
			pamstub.addTransFn = newTransFn
			pamstub.getTransFn = tt.fields.getTransFn

			svc := NewService(&pamstub)

			result, err := svc.Cancel(CancelRequest{
				Currency: "EUR",
				Transaction: Transaction{
					RefID: "ref1",
					ID:    "trans2",
				}})

			if tt.wantErr != nil {
				var provErr ProviderError
				require.ErrorAs(t, err, &provErr, "expecting a provider error response")
				assert.Equal(t, tt.wantErr.response.Status, provErr.response.Status)
				assert.Equal(t, tt.wantErr.response.Balance, provErr.response.Balance)
			}

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestProviderService_PromoPayout(t *testing.T) {
	// prepare pam stub with fixed responses
	pamstub := pamStub{}
	pamstub.balanceFn = func() (*pam.Balance, error) {
		return &pam.Balance{
			CashAmount: testutils.NewFloatAmount(1000),
		}, nil
	}

	pamstub.addTransFn = func() (*pam.TransactionResult, error) {
		id := "trans2"
		balance := pam.Balance{
			CashAmount:  testutils.NewFloatAmount(1000),
			BonusAmount: pam.ZeroAmount,
		}
		return &pam.TransactionResult{TransactionId: &id, Balance: &balance}, nil
	}

	svc := NewService(&pamstub)

	type fields struct {
		getTransFn func() ([]pam.Transaction, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    PromoPayoutRequest
		want    *StandardResponse
		wantErr *ProviderError
	}{
		{
			name: "happy case",
			fields: fields{
				getTransFn: func() ([]pam.Transaction, error) {
					return []pam.Transaction{createTrans(pam.WITHDRAW, 11, "ref1", "trans1")}, nil
				},
			},
			args: PromoPayoutRequest{
				Currency: "EUR",
			},
			want: &StandardResponse{
				Status:  "OK",
				Balance: amountFromFloat(1000),
				Bonus:   ZeroAmount,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup stub methods
			pamstub.getTransFn = tt.fields.getTransFn

			result, err := svc.PromoPayout(tt.args)

			if tt.wantErr != nil {
				var provErr ProviderError
				require.ErrorAs(t, err, &provErr, "expecting a provider error response")
				assert.Equal(t, tt.wantErr.response.Status, provErr.response.Status)
				assert.Equal(t, tt.wantErr.response.Balance, provErr.response.Balance)
			}

			assert.Equal(t, tt.want, result)
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

		pamStub.balanceFn = func() (*pam.Balance, error) {
			return nil, pam.ErrorWrapper("a", tt, assert.AnError)
		}

		_, err := svc.Balance(BalanceRequest{})
		assert.Error(t, err)
		var pErr ProviderError
		if assert.ErrorAs(t, err, &pErr) {
			assert.Equal(t, StatusInvalidSID.code, pErr.response.Status)
		}
	}
}

func createTrans(typ pam.TransactionType, amt float64, pRef string, pTransID string) pam.Transaction {
	return pam.Transaction{
		TransactionType:       typ,
		CashAmount:            testutils.NewFloatAmount(amt),
		ProviderBetRef:        &pRef,
		ProviderTransactionId: pTransID,
	}
}

type pamStub struct {
	pam.PamClient
	balanceFn        func() (*pam.Balance, error)
	sessionFn        func() (*pam.Session, error)
	refreshSessionFn func() (*pam.Session, error)
	addTransFn       func() (*pam.TransactionResult, error)
	getTransFn       func() ([]pam.Transaction, error)
	getGameRoundFn   func() (*pam.GameRound, error)
}

func (pam *pamStub) GetSession(_ pam.GetSessionRequestMapper) (*pam.Session, error) {
	return pam.sessionFn()
}

func (pam *pamStub) RefreshSession(_ pam.RefreshSessionRequestMapper) (*pam.Session, error) {
	return pam.refreshSessionFn()
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
