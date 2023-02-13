package evolution

import (
	"context"
	"testing"
	"time"

	"github.com/valkyrie-fnd/valkyrie/pam/testutils"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

var s = &ProviderService{ctx: context.Background()}

var noRoundingRounder pam.AmountRounder = func(amt pam.Amt) (*pam.Amount, error) {
	a := pam.Amount(amt)
	return &a, nil
}

func Test_balanceRequestMapper(t *testing.T) {
	type args struct {
		r RequestBase
	}
	tests := []struct {
		name string
		args args
		want pam.GetBalanceRequest
	}{
		{
			name: "balance request should map to a valid balance request",
			args: args{
				r: RequestBase{
					SID:    "asdf",
					UserID: "boo",
					UUID:   "uuid-0a",
				},
			},
			want: pam.GetBalanceRequest{
				Params: pam.GetBalanceParams{
					Provider:       ProviderName,
					XPlayerToken:   "asdf",
					XCorrelationID: "uuid-0a",
				},
				PlayerID: "boo",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.balanceRequestMapper(tt.args.r)()
			assert.Nil(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_refreshSessionRequest(t *testing.T) {
	type args struct {
		r CheckRequest
	}
	tests := []struct {
		name string
		args args
		want pam.RefreshSessionRequest
	}{
		{
			name: "Valid check request should map to a valid balance request",
			args: args{
				r: CheckRequest{
					RequestBase: RequestBase{
						SID:    "asdf",
						UserID: "boo",
						UUID:   "uuid-0a",
					},
				},
			},
			want: pam.RefreshSessionRequest{
				Params: pam.RefreshSessionParams{
					Provider:       ProviderName,
					XPlayerToken:   "asdf",
					XCorrelationID: "uuid-0a",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.refreshSessionRequestMapper(tt.args.r)()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_debitRequestMapper(t *testing.T) {
	providerGameID := "table_id"
	providerRoundID := "evo_round_id_1"
	transTime := time.Now()
	type args struct {
		r DebitRequest
	}
	tests := []struct {
		name string
		args args
		want *pam.AddTransactionRequest
	}{
		{
			name: "Transaction ref and id mapped to external fields",
			args: args{
				r: DebitRequest{
					Currency: "EUR",
					Transaction: Transaction{
						ID:     "evo_trans_id",
						RefID:  "evo_ref",
						Amount: amountFromFloat(123.123),
					},
					RequestBase: RequestBase{
						SID:    "sessXXX",
						UserID: "player1",
					},
					Game: Game{
						ID: "evo_round_id_1",
						Details: GameDetails{
							Table: GameTable{
								ID: "table_id",
							},
						},
					},
				},
			},
			want: &pam.AddTransactionRequest{
				PlayerID: "player1",
				Params: pam.AddTransactionParams{
					Provider:     ProviderName,
					XPlayerToken: "sessXXX",
				},
				Body: pam.AddTransactionJSONRequestBody{
					TransactionType:       pam.WITHDRAW,
					CashAmount:            testutils.NewFloatAmount(123.123),
					BonusAmount:           pam.ZeroAmount,
					PromoAmount:           pam.ZeroAmount,
					Currency:              "EUR",
					ProviderTransactionId: "evo_trans_id",
					ProviderBetRef:        strPtr("evo_ref"),
					ProviderGameId:        &providerGameID,
					ProviderRoundId:       &providerRoundID,
					TransactionDateTime:   transTime,
					Provider:              ProviderName,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.debitRequestMapper(tt.args.r, transTime)(noRoundingRounder)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
			assert.Equal(t, transTime, res.Body.TransactionDateTime, "transaction time is not provided by evo")
		})
	}
}

func Test_creditTransRequestMapper(t *testing.T) {
	providerGameID := "table_id"
	providerRoundID := "evo_round_id_1"
	transTime := time.Now()
	type args struct {
		r CreditRequest
	}
	tests := []struct {
		name string
		args args
		want *pam.AddTransactionRequest
	}{
		{
			name: "credit request transaction ref and id mapped",
			args: args{
				r: CreditRequest{
					Currency: "EUR",
					Transaction: Transaction{
						ID:     "evo_trans_id",
						RefID:  "evo_ref",
						Amount: amountFromFloat(123.123),
					},
					RequestBase: RequestBase{
						SID:    "sessXXX",
						UserID: "player1",
					},
					Game: Game{
						ID: "evo_round_id_1",
						Details: GameDetails{
							Table: GameTable{
								ID: "table_id",
							},
						},
					},
				},
			},
			want: &pam.AddTransactionRequest{
				PlayerID: "player1",
				Params: pam.AddTransactionParams{
					Provider:     ProviderName,
					XPlayerToken: "sessXXX",
				},
				Body: pam.AddTransactionJSONRequestBody{
					TransactionType:       pam.DEPOSIT,
					CashAmount:            testutils.NewFloatAmount(123.123),
					BonusAmount:           pam.ZeroAmount,
					PromoAmount:           pam.ZeroAmount,
					Currency:              "EUR",
					ProviderTransactionId: "evo_trans_id",
					ProviderBetRef:        strPtr("evo_ref"),
					ProviderGameId:        &providerGameID,
					ProviderRoundId:       &providerRoundID,
					TransactionDateTime:   transTime,
					Provider:              ProviderName,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.creditTransRequestMapper(tt.args.r, transTime)(noRoundingRounder)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
			assert.Equal(t, transTime, res.Body.TransactionDateTime, "transaction time is not provided by evo")
		})
	}
}

func Test_cancelTransRequestMapper(t *testing.T) {
	providerGameID := "table_id"
	providerRoundID := "evo_round_id_1"
	transTime := time.Now()
	type args struct {
		r CancelRequest
	}
	tests := []struct {
		name string
		args args
		want *pam.AddTransactionRequest
	}{
		{
			name: "cancel request transaction id and ref mapped",
			args: args{
				r: CancelRequest{
					Currency: "EUR",
					Transaction: Transaction{
						ID:     "evo_trans_id",
						RefID:  "evo_ref",
						Amount: amountFromFloat(123.123),
					},
					RequestBase: RequestBase{
						SID:    "sessXXX",
						UserID: "player1",
					},
					Game: Game{
						ID: "evo_round_id_1",
						Details: GameDetails{
							Table: GameTable{
								ID: "table_id",
							},
						},
					},
				},
			},
			want: &pam.AddTransactionRequest{
				PlayerID: "player1",
				Params: pam.AddTransactionParams{
					Provider:     ProviderName,
					XPlayerToken: "sessXXX",
				},
				Body: pam.AddTransactionJSONRequestBody{
					TransactionType:       pam.CANCEL,
					CashAmount:            testutils.NewFloatAmount(123.123),
					BonusAmount:           pam.ZeroAmount,
					PromoAmount:           pam.ZeroAmount,
					Currency:              "EUR",
					ProviderTransactionId: "evo_trans_id",
					ProviderBetRef:        strPtr("evo_ref"),
					ProviderGameId:        &providerGameID,
					ProviderRoundId:       &providerRoundID,
					TransactionDateTime:   transTime,
					Provider:              ProviderName,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.cancelTransRequestMapper(tt.args.r, transTime)(noRoundingRounder)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
			assert.Equal(t, transTime, res.Body.TransactionDateTime, "transaction time is not provided by evo")
		})
	}
}

func Test_promoPayoutTransRequestMapper(t *testing.T) {
	providerGameID := "table_id"
	providerRoundID := "evo_round_id_1"
	transTime := time.Now()
	type args struct {
		r PromoPayoutRequest
	}
	tests := []struct {
		name string
		args args
		want *pam.AddTransactionRequest
	}{
		{
			name: "promo payout request transaction id mapped",
			args: args{
				r: PromoPayoutRequest{
					Currency: "EUR",
					PromoTransaction: PromoTransaction{
						ID:     "evo_trans_id",
						Amount: amountFromFloat(123.123),
					},
					RequestBase: RequestBase{
						SID:    "sessXXX",
						UserID: "player1",
					},
					Game: Game{
						ID: "evo_round_id_1",
						Details: GameDetails{
							Table: GameTable{
								ID: "table_id",
							},
						},
					},
				},
			},
			want: &pam.AddTransactionRequest{
				PlayerID: "player1",
				Params: pam.AddTransactionParams{
					Provider:     ProviderName,
					XPlayerToken: "sessXXX",
				},
				Body: pam.AddTransactionJSONRequestBody{
					TransactionType:       pam.PROMODEPOSIT,
					CashAmount:            testutils.NewFloatAmount(123.123),
					BonusAmount:           pam.ZeroAmount,
					PromoAmount:           pam.ZeroAmount,
					Currency:              "EUR",
					ProviderTransactionId: "evo_trans_id",
					ProviderGameId:        &providerGameID,
					ProviderRoundId:       &providerRoundID,
					TransactionDateTime:   transTime,
					Provider:              ProviderName,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.promoPayoutTransRequestMapper(tt.args.r, transTime)(noRoundingRounder)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
			assert.Equal(t, transTime, res.Body.TransactionDateTime, "transaction time is not provided by evo")
		})
	}
}

func Test_findTransForCreditRequestMapper(t *testing.T) {
	type args struct {
		r CreditRequest
	}
	tests := []struct {
		name string
		args args
		want pam.GetTransactionsRequest
	}{
		{
			name: "map credit request to get transaction",
			args: args{
				r: CreditRequest{
					RequestBase: RequestBase{
						SID:    "asdf",
						UserID: "boo",
						UUID:   "uuid-0a",
					},
					Transaction: Transaction{RefID: "refid"},
				},
			},
			want: pam.GetTransactionsRequest{
				PlayerID: "boo",
				Params: pam.GetTransactionsParams{
					Provider:       ProviderName,
					XPlayerToken:   "asdf",
					ProviderBetRef: strPtr("refid"),
					XCorrelationID: "uuid-0a",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.findTransForCreditRequestMapper(tt.args.r)()
			assert.Nil(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_findTransForCancelRequestMapper(t *testing.T) {
	type args struct {
		r CancelRequest
	}
	tests := []struct {
		name string
		args args
		want pam.GetTransactionsRequest
	}{
		{
			name: "map cancel request to get transaction",
			args: args{
				r: CancelRequest{
					RequestBase: RequestBase{
						SID:    "asdf",
						UserID: "boo",
						UUID:   "uuid-0a",
					},
					Transaction: Transaction{RefID: "refid"},
				},
			},
			want: pam.GetTransactionsRequest{
				PlayerID: "boo",
				Params: pam.GetTransactionsParams{
					Provider:       ProviderName,
					XPlayerToken:   "asdf",
					ProviderBetRef: strPtr("refid"),
					XCorrelationID: "uuid-0a",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := s.findTransForCancelRequestMapper(tt.args.r)()
			assert.Nil(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

// TestRoundingError checks that rounding errors are will be properly handled
func TestRoundingError(t *testing.T) {
	failingRounder := func(amt pam.Amt) (*pam.Amount, error) { return nil, assert.AnError }

	_, _, err := s.debitRequestMapper(DebitRequest{}, time.Now())(failingRounder)
	assert.Error(t, err, "debitRequestMapper should fail on rounding errors")

	_, _, err = s.creditTransRequestMapper(CreditRequest{}, time.Now())(failingRounder)
	assert.Error(t, err, "creditTransRequestMapper should fail on rounding errors")

	_, _, err = s.cancelTransRequestMapper(CancelRequest{}, time.Now())(failingRounder)
	assert.Error(t, err, "cancelTransRequestMapper should fail on rounding errors")

	_, _, err = s.promoPayoutTransRequestMapper(PromoPayoutRequest{}, time.Now())(failingRounder)
	assert.Error(t, err, "promoPayoutTransRequestMapper should fail on rounding errors")
}

func strPtr(str string) *string {
	return &str
}
