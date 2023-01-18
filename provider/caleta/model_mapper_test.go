package caleta

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"
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
		name        string
		request     *WalletbetRequestObject
		want        *pam.AddTransactionRequest
		dateCompare dateAssert
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.WITHDRAW,
					Provider:              "Caleta",
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.WITHDRAW,
					Provider:              "Caleta",
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
			_, res, err := betTransactionMapper(context.TODO(), tt.request)(dummyAmtReader)

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
		name        string
		request     *WalletbetRequestObject
		want        *pam.AddTransactionRequest
		dateCompare dateAssert
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.PROMOWITHDRAW,
					Provider:              "Caleta",
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.PROMOWITHDRAW,
					Provider:              "Caleta",
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
			_, res, err := promoBetTransactionMapper(context.TODO(), tt.request)(dummyAmtReader)

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
		name        string
		request     *TransactionwinRequestObject
		want        *pam.AddTransactionRequest
		dateCompare dateAssert
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.DEPOSIT,
					Provider:              "Caleta",
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.DEPOSIT,
					Provider:              "Caleta",
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
			_, res, err := winTransactionMapper(context.TODO(), tt.request)(dummyAmtReader)

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
		name        string
		request     *TransactionwinRequestObject
		want        *pam.AddTransactionRequest
		dateCompare dateAssert
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.PROMODEPOSIT,
					Provider:              "Caleta",
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.PROMODEPOSIT,
					Provider:              "Caleta",
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
			_, res, err := promoWinTransactionMapper(context.TODO(), tt.request)(dummyAmtReader)

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
		name        string
		request     *WalletrollbackRequestObject
		want        *pam.AddTransactionRequest
		dateCompare dateAssert
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   now,
					TransactionType:       pam.PROMOCANCEL,
					Provider:              "Caleta",
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
					Provider:     "Caleta",
				},
				Body: pam.Transaction{
					TransactionDateTime:   time.Now(),
					TransactionType:       pam.PROMOCANCEL,
					Provider:              "Caleta",
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
			_, res, err := cancelTransactionMapper(context.TODO(), tt.request, &pam.Session{Currency: "EUR"}, pam.PROMOCANCEL)(dummyAmtReader)

			tt.dateCompare(t, tt.want.Body.TransactionDateTime, res.Body.TransactionDateTime)

			// date already compared now reset them
			res.Body.TransactionDateTime = now
			tt.want.Body.TransactionDateTime = now

			assert.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}
