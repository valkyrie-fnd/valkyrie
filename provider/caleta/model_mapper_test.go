package caleta

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/pam"
)

var dummyAmtReader pam.AmountRounder = func(amt pam.Amt) (*pam.Amount, error) { return utils.Ptr(pam.Amount(amt)), nil }

var now = time.Now().UTC()
var nowMsgTst = MsgTimestamp(now)

type dateAssert func(*testing.T, time.Time, time.Time) bool

var equalDate dateAssert = func(t *testing.T, expected time.Time, actual time.Time) bool {
	return assert.Equal(t, expected, actual)
}
var laterDate dateAssert = func(t *testing.T, start time.Time, actual time.Time) bool {
	return assert.GreaterOrEqual(t, actual, start)
}

func Test_betTransactionMapper(t *testing.T) {

	tests := []struct {
		name              string
		request           *WalletbetRequestObject
		roundTransactions *[]roundTransaction
		want              *pam.AddTransactionRequest
		dateCompare       dateAssert
	}{
		{
			name: "basic",
			request: &WalletbetRequestObject{
				Params: WalletbetParams{
					XMsgTimestamp: &nowMsgTst,
				},
				Body: &WalletBetBody{
					Amount:          1,
					Currency:        "EUR",
					GameCode:        "a",
					TransactionUuid: "id",
					Token:           "tkn",
					Round:           "r1",
				},
			},
			dateCompare: equalDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.WITHDRAW,
					Provider:              "caleta",
					CashAmount:            toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("id"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
		{
			name: "without date",
			request: &WalletbetRequestObject{
				Params: WalletbetParams{},
				Body: &WalletBetBody{
					Amount:          1,
					Currency:        "EUR",
					GameCode:        "a",
					TransactionUuid: "id",
					Token:           "tkn",
					Round:           "r1",
				},
			},
			dateCompare: laterDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.WITHDRAW,
					Provider:              "caleta",
					CashAmount:            toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("id"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := betTransactionMapper(context.TODO(), tt.request, tt.roundTransactions)(dummyAmtReader)

			tt.dateCompare(t, tt.want.Body.TransactionDateTime, res.Body.TransactionDateTime)

			// date already compared now reset them
			res.Body.TransactionDateTime = now
			tt.want.Body.TransactionDateTime = now

			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_promoBetTransactionMapper(t *testing.T) {
	tests := []struct {
		name              string
		request           *WalletbetRequestObject
		roundTransactions *[]roundTransaction
		want              *pam.AddTransactionRequest
		dateCompare       dateAssert
	}{
		{
			name: "basic",
			request: &WalletbetRequestObject{
				Params: WalletbetParams{
					XMsgTimestamp: &nowMsgTst,
				},
				Body: &WalletBetBody{
					Amount:          1,
					Currency:        "EUR",
					GameCode:        "a",
					TransactionUuid: "id",
					Token:           "tkn",
					Round:           "r1",
				},
			},
			dateCompare: equalDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.PROMOWITHDRAW,
					Provider:              "caleta",
					PromoAmount:           toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("id"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
		{
			name: "without date",
			request: &WalletbetRequestObject{
				Params: WalletbetParams{},
				Body: &WalletBetBody{
					Amount:          1,
					Currency:        "EUR",
					GameCode:        "a",
					TransactionUuid: "id",
					Token:           "tkn",
					Round:           "r1",
				},
			},
			dateCompare: laterDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.PROMOWITHDRAW,
					Provider:              "caleta",
					PromoAmount:           toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("id"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := promoBetTransactionMapper(context.TODO(), tt.request, tt.roundTransactions)(dummyAmtReader)

			tt.dateCompare(t, tt.want.Body.TransactionDateTime, res.Body.TransactionDateTime)

			// date already compared now reset them
			res.Body.TransactionDateTime = now
			tt.want.Body.TransactionDateTime = now

			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_winTransactionMapper(t *testing.T) {
	tests := []struct {
		name              string
		request           *TransactionwinRequestObject
		roundTransactions *[]roundTransaction
		want              *pam.AddTransactionRequest
		dateCompare       dateAssert
	}{
		{
			name: "basic",
			request: &TransactionwinRequestObject{
				Params: TransactionwinParams{
					XMsgTimestamp: &nowMsgTst,
				},
				Body: &WalletWinBody{
					Amount:                   1,
					Currency:                 "EUR",
					GameCode:                 "a",
					TransactionUuid:          "id",
					ReferenceTransactionUuid: "ref",
					Token:                    "tkn",
					Round:                    "r1",
				},
			},
			dateCompare: equalDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.DEPOSIT,
					Provider:              "caleta",
					CashAmount:            toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("ref"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
		{
			name: "without date",
			request: &TransactionwinRequestObject{
				Params: TransactionwinParams{},
				Body: &WalletWinBody{
					Amount:                   1,
					Currency:                 "EUR",
					GameCode:                 "a",
					TransactionUuid:          "id",
					ReferenceTransactionUuid: "ref",
					Token:                    "tkn",
					Round:                    "r1",
				},
			},
			dateCompare: laterDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.DEPOSIT,
					Provider:              "caleta",
					CashAmount:            toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("ref"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := winTransactionMapper(context.TODO(), tt.request, tt.roundTransactions)(dummyAmtReader)

			tt.dateCompare(t, tt.want.Body.TransactionDateTime, res.Body.TransactionDateTime)

			// date already compared now reset them
			res.Body.TransactionDateTime = now
			tt.want.Body.TransactionDateTime = now

			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_promoWinTransactionMapper(t *testing.T) {
	tests := []struct {
		name              string
		request           *TransactionwinRequestObject
		roundTransactions *[]roundTransaction
		want              *pam.AddTransactionRequest
		dateCompare       dateAssert
	}{
		{
			name: "basic",
			request: &TransactionwinRequestObject{
				Params: TransactionwinParams{
					XMsgTimestamp: &nowMsgTst,
				},
				Body: &WalletWinBody{
					Amount:                   1,
					Currency:                 "EUR",
					GameCode:                 "a",
					TransactionUuid:          "id",
					ReferenceTransactionUuid: "ref",
					Token:                    "tkn",
					Round:                    "r1",
				},
			},
			dateCompare: equalDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.PROMODEPOSIT,
					Provider:              "caleta",
					PromoAmount:           toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("ref"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
		{
			name: "without date",
			request: &TransactionwinRequestObject{
				Params: TransactionwinParams{},
				Body: &WalletWinBody{
					Amount:                   1,
					Currency:                 "EUR",
					GameCode:                 "a",
					TransactionUuid:          "id",
					ReferenceTransactionUuid: "ref",
					Token:                    "tkn",
					Round:                    "r1",
				},
			},
			dateCompare: laterDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.PROMODEPOSIT,
					Provider:              "caleta",
					PromoAmount:           toPamAmount(1),
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("ref"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := promoWinTransactionMapper(context.TODO(), tt.request, tt.roundTransactions)(dummyAmtReader)

			tt.dateCompare(t, tt.want.Body.TransactionDateTime, res.Body.TransactionDateTime)

			// date already compared now reset them
			res.Body.TransactionDateTime = now
			tt.want.Body.TransactionDateTime = now

			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_cancelTransactionMapper(t *testing.T) {

	tests := []struct {
		name              string
		request           *WalletrollbackRequestObject
		roundTransactions *[]roundTransaction
		want              *pam.AddTransactionRequest
		dateCompare       dateAssert
	}{
		{
			name: "basic",
			request: &WalletrollbackRequestObject{
				Params: WalletrollbackParams{
					XMsgTimestamp: &nowMsgTst,
				},
				Body: &WalletRollbackBody{
					GameCode:                 "a",
					TransactionUuid:          "id",
					ReferenceTransactionUuid: "ref",
					Token:                    "tkn",
					Round:                    "r1",
					RoundClosed:              true,
				},
			},
			dateCompare: equalDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.PROMOCANCEL,
					Provider:              "caleta",
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("ref"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(true),
				},
			},
		},
		{
			name: "without date",
			request: &WalletrollbackRequestObject{
				Params: WalletrollbackParams{},
				Body: &WalletRollbackBody{
					GameCode:                 "a",
					TransactionUuid:          "id",
					ReferenceTransactionUuid: "ref",
					Token:                    "tkn",
					Round:                    "r1",
				},
			},
			dateCompare: laterDate,
			want: &pam.AddTransactionRequest{
				Params: pam.AddTransactionParams{
					XPlayerToken: "tkn",
					Provider:     "caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.PROMOCANCEL,
					Provider:              "caleta",
					Currency:              "EUR",
					ProviderTransactionId: "id",
					ProviderBetRef:        utils.Ptr("ref"),
					ProviderGameId:        utils.Ptr("a"),
					ProviderRoundId:       utils.Ptr("r1"),
					IsGameOver:            utils.Ptr(false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, res, err := cancelTransactionMapper(context.TODO(), tt.request, &pam.Session{Currency: "EUR"}, pam.PROMOCANCEL, tt.roundTransactions)(dummyAmtReader)

			tt.dateCompare(t, tt.want.Body.TransactionDateTime, res.Body.TransactionDateTime)

			// date already compared now reset them
			res.Body.TransactionDateTime = now
			tt.want.Body.TransactionDateTime = now

			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_roundTransactionsMapper(t *testing.T) {
	tests := []struct {
		name  string
		input *[]roundTransaction
		want  *[]pam.RoundTransaction
	}{
		{
			name:  "no transactions",
			input: nil,
			want:  nil,
		},
		{
			name: "one transaction",
			input: &[]roundTransaction{
				{
					ID:          303,
					CreatedTime: now,
					ClosedTime:  now.Add(1 + time.Second),
					TxnUUID:     "txn-uuid",
					Payload: payload{
						Bet:                      "zero",
						Round:                    "CG-303",
						Token:                    "token",
						Currency:                 "EUR",
						GameCode:                 "gc",
						RequestUUID:              "req-uuid",
						SupplierUser:             "supp-usr",
						TransactionUUID:          "trans-uuid",
						ReferenceTransactionUUID: testutils.Ptr("ref-trans-uuid"),
						GameID:                   1,
						Amount:                   200000,
						RoundClosed:              true,
						IsFree:                   false,
					},
					RoundID:      101,
					TxnType:      0,
					Status:       909,
					CacheEntryID: 606,
					Amount:       200000,
				},
			},
			want: &[]pam.RoundTransaction{
				{
					ProviderTransactionId: utils.Ptr("trans-uuid"),
					CashAmount:            utils.Ptr(toPamAmount(200000)),
					IsGameOver:            utils.Ptr(true),
					TransactionDateTime:   utils.Ptr(now),
					ProviderBetRef:        utils.Ptr("ref-trans-uuid"),
					TransactionType:       pam.DEPOSIT,
					BetCode:               utils.Ptr("zero"),
				},
			},
		},
		{
			name: "three transactions",
			input: &[]roundTransaction{
				{
					ID:          303,
					CreatedTime: now,
					ClosedTime:  now,
					TxnUUID:     "txn-uuid-0",
					Payload: payload{
						Bet:                 "Base",
						Round:               "CG-303",
						Token:               "token",
						Currency:            "EUR",
						GameCode:            "gc",
						RequestUUID:         "req-uuid-0",
						SupplierUser:        "supp-usr",
						TransactionUUID:     "txn-uuid-0",
						GameID:              1,
						JackpotContribution: 2000,
						Amount:              200000,
						RoundClosed:         false,
						IsFree:              false,
					},
					RoundID:      101,
					TxnType:      0,
					Status:       909,
					CacheEntryID: 606,
					Amount:       200000,
				},
				{
					ID:          303,
					CreatedTime: now.Add(1 * time.Second),
					ClosedTime:  now.Add(1 * time.Second),
					TxnUUID:     "txn-uuid-1",
					Payload: payload{
						Bet:                 "Extra Ball",
						Round:               "CG-303",
						Token:               "token",
						Currency:            "EUR",
						GameCode:            "gc",
						RequestUUID:         "req-uuid-1",
						SupplierUser:        "supp-usr",
						TransactionUUID:     "txn-uuid-1",
						GameID:              1,
						JackpotContribution: 3000,
						Amount:              300000,
						RoundClosed:         false,
						IsFree:              false,
					},
					RoundID:      101,
					TxnType:      0,
					Status:       909,
					CacheEntryID: 606,
					Amount:       300000,
				},
				{
					ID:          303,
					CreatedTime: now.Add(2 * time.Second),
					ClosedTime:  now.Add(2 * time.Second),
					TxnUUID:     "txn-uuid-2",
					Payload: payload{
						Bet:                      "zero",
						Round:                    "CG-303",
						Token:                    "token",
						Currency:                 "EUR",
						GameCode:                 "gc",
						RequestUUID:              "req-uuid-2",
						SupplierUser:             "supp-usr",
						TransactionUUID:          "txn-uuid-2",
						ReferenceTransactionUUID: testutils.Ptr("txn-uuid-0"),
						GameID:                   1,
						Amount:                   200000,
						RoundClosed:              true,
						IsFree:                   false,
					},
					RoundID:      101,
					TxnType:      0,
					Status:       909,
					CacheEntryID: 606,
					Amount:       200000,
				},
			},
			want: &[]pam.RoundTransaction{
				{
					ProviderTransactionId: utils.Ptr("txn-uuid-0"),
					CashAmount:            utils.Ptr(toPamAmount(200000)),
					IsGameOver:            utils.Ptr(false),
					TransactionDateTime:   utils.Ptr(now),
					JackpotContribution:   utils.Ptr(toPamAmount(2000)),
					TransactionType:       pam.WITHDRAW,
					BetCode:               utils.Ptr("Base"),
				},
				{
					ProviderTransactionId: utils.Ptr("txn-uuid-1"),
					CashAmount:            utils.Ptr(toPamAmount(300000)),
					IsGameOver:            utils.Ptr(false),
					TransactionDateTime:   utils.Ptr(now.Add(1 * time.Second)),
					JackpotContribution:   utils.Ptr(toPamAmount(3000)),
					TransactionType:       pam.WITHDRAW,
					BetCode:               utils.Ptr("Extra Ball"),
				},
				{
					ProviderTransactionId: utils.Ptr("txn-uuid-2"),
					CashAmount:            utils.Ptr(toPamAmount(200000)),
					IsGameOver:            utils.Ptr(true),
					TransactionDateTime:   utils.Ptr(now.Add(2 * time.Second)),
					ProviderBetRef:        utils.Ptr("txn-uuid-0"),
					TransactionType:       pam.DEPOSIT,
					BetCode:               utils.Ptr("zero"),
				},
			},
		},
		{
			name: "cancel transaction",
			input: &[]roundTransaction{
				{
					ID:          303,
					CreatedTime: now,
					ClosedTime:  now.Add(1 + time.Second),
					TxnUUID:     "txn-uuid",
					Payload: payload{
						Round:                    "CG-303",
						Token:                    "token",
						GameCode:                 "gc",
						RequestUUID:              "req-uuid",
						SupplierUser:             "supp-usr",
						TransactionUUID:          "trans-uuid",
						ReferenceTransactionUUID: testutils.Ptr("ref-trans-uuid"),
						GameID:                   1,
						Amount:                   200000,
						RoundClosed:              true,
						IsFree:                   false,
					},
					RoundID:      101,
					TxnType:      0,
					Status:       909,
					CacheEntryID: 606,
					Amount:       200000,
				},
			},
			want: &[]pam.RoundTransaction{
				{
					ProviderTransactionId: utils.Ptr("trans-uuid"),
					CashAmount:            utils.Ptr(toPamAmount(200000)),
					IsGameOver:            utils.Ptr(true),
					TransactionDateTime:   utils.Ptr(now),
					ProviderBetRef:        utils.Ptr("ref-trans-uuid"),
					TransactionType:       pam.CANCEL,
				},
			},
		},
		{
			name: "unknown transaction type",
			input: &[]roundTransaction{
				{
					ID:          303,
					CreatedTime: now,
					ClosedTime:  now.Add(1 + time.Second),
					TxnUUID:     "txn-uuid",
					Payload: payload{
						Round:                    "CG-303",
						Token:                    "token",
						GameCode:                 "gc",
						Currency:                 "EUR",
						RequestUUID:              "req-uuid",
						SupplierUser:             "supp-usr",
						TransactionUUID:          "trans-uuid",
						ReferenceTransactionUUID: testutils.Ptr("ref-trans-uuid"),
						GameID:                   1,
						Amount:                   200000,
						RoundClosed:              true,
						IsFree:                   false,
					},
					RoundID:      101,
					TxnType:      0,
					Status:       909,
					CacheEntryID: 606,
					Amount:       200000,
				},
			},
			want: new([]pam.RoundTransaction),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := roundTransactionsMapper(tt.input)
			assert.Equal(t, tt.want, res)
		})
	}
}
