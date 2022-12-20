package vplugin

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/pam"
)

func TestVPlugin_GobCodecs(t *testing.T) {
	testCases := []struct {
		desc   string
		source any
		target any
	}{
		{
			desc:   "gob Amount",
			source: testutils.NewFloatAmount(123.123),
			target: testutils.NewFloatAmount(123.123),
		},
		{
			desc: "gob GetBalanceRequest",
			source: pam.GetBalanceRequest{
				PlayerID: "apa",
				Params: pam.GetBalanceParams{
					Provider: "prov1der",
				},
			},
			target: pam.GetBalanceRequest{},
		},
		{
			desc: "gob BalanceResponse",
			source: pam.BalanceResponse{
				Status: pam.OK,
				Balance: &pam.Balance{
					BonusAmount: testutils.NewFloatAmount(123),
				},
			},
			target: pam.BalanceResponse{},
		},
		{
			desc: "gob RefreshSessionRequest",
			source: pam.RefreshSessionRequest{
				Params: pam.RefreshSessionParams{
					Provider:     "prov1der",
					XPlayerToken: "assksksksksks",
				},
			},
			target: pam.RefreshSessionRequest{},
		},
		{
			desc: "gob Session",
			source: pam.SessionResponse{
				Status: pam.OK,
				Session: &pam.Session{
					Country:  "ye",
					PlayerId: "id",
				},
			},
			target: pam.Session{},
		},
		{
			desc: "gob GetTransactionsRequest",
			source: pam.GetTransactionsRequest{
				PlayerID: "id",
				Params: pam.GetTransactionsParams{
					Provider:       "prov",
					ProviderBetRef: utils.Ptr("ref"),
				},
			},
			target: pam.GetTransactionsRequest{},
		},
		{
			desc: "gob AddTransactionResponse",
			source: pam.AddTransactionResponse{
				TransactionResult: &pam.TransactionResult{
					TransactionId: utils.Ptr("ett"),
				},
			},
			target: pam.AddTransactionResponse{},
		},
		{
			desc: "gob AddTransactionRequest",
			source: pam.AddTransactionRequest{
				PlayerID: "id",
				Params: pam.AddTransactionParams{
					XPlayerToken: "token",
				},
				Body: pam.Transaction{
					CashAmount: testutils.NewFloatAmount(123),
				},
			},
			target: pam.AddTransactionRequest{},
		},
		{
			desc: "gob GetGameRoundRequest",
			source: pam.GetGameRoundRequest{
				ProviderRoundID: "...",
			},
			target: pam.GetGameRoundRequest{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var sink bytes.Buffer
			enc := gob.NewEncoder(&sink)
			err := enc.Encode(&tC.source)
			assert.NoError(t, err)

			dec := gob.NewDecoder(&sink)
			err = dec.Decode(&tC.target)

			if assert.NoError(t, err) {
				assert.Equal(t, tC.source, tC.target)
			}
		})
	}
}
