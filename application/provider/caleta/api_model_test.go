package caleta

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_transactionResponse(t *testing.T) {
	expectedTimestamp, _ := time.Parse(time.RFC3339Nano, "2023-01-17T13:55:26.551Z")

	tests := []struct {
		name     string
		response string
		want     transactionResponse
	}{
		{
			name: "No transactions found",
			response: `
			{
				"code": 1016,
				"message": "Invalid Round"
			}`,
			want: transactionResponse{Code: 1016, Message: "Invalid Round"},
		},
		{
			name: "Transactions found",
			response: `
			{
				"round_id": "CG-909",
				"transactions": [
				  {
					"round_id": 909,
					"txn_uuid": "txn-uuid-0",
					"created_time": "2023-01-17T13:55:26.551Z",
					"closed_time": "2023-01-17T13:55:26.552Z",
					"amount": 200000,
					"payload": {
					  "bet": "Base",
					  "round": "CG-909",
					  "token": "token-0",
					  "amount": 200000,
					  "game_id": 121,
					  "is_free": false,
					  "currency": "GBP",
					  "game_code": "game-code",
					  "request_uuid": "req-uuid-0",
					  "round_closed": false,
					  "supplier_user": "6",
					  "transaction_uuid": "txn-uuid-0",
					  "jackpot_contribution": 2000
					}
				  },
				  {
					"round_id": 909,
					"txn_uuid": "txn-uuid-1",
					"created_time": "2023-01-17T13:55:26.553Z",
					"closed_time": "2023-01-17T13:55:26.554Z",
					"amount": 3250000,
					"payload": {
					  "bet": "zero",
					  "round": "CG-909",
					  "token": "token-1",
					  "amount": 3250000,
					  "game_id": 121,
					  "is_free": false,
					  "currency": "GBP",
					  "game_code": "game-code",
					  "request_uuid": "req-uuid-1",
					  "round_closed": true,
					  "supplier_user": "6",
					  "transaction_uuid": "txn-uuid-1",
					  "reference_transaction_uuid": "txn-uuid-0"
					}
				  }
				]
			}`,
			want: transactionResponse{
				RoundID: "CG-909",
				RoundTransactions: &[]roundTransaction{
					{
						RoundID:     909,
						TxnUUID:     "txn-uuid-0",
						CreatedTime: expectedTimestamp,
						ClosedTime:  expectedTimestamp.Add(1 * time.Millisecond),
						Amount:      200000,
						Payload: payload{
							Bet:                 "Base",
							Round:               "CG-909",
							Token:               "token-0",
							Amount:              200000,
							GameID:              121,
							IsFree:              false,
							Currency:            "GBP",
							GameCode:            "game-code",
							RequestUUID:         "req-uuid-0",
							RoundClosed:         false,
							SupplierUser:        "6",
							TransactionUUID:     "txn-uuid-0",
							JackpotContribution: 2000,
						},
					},
					{
						RoundID:     909,
						TxnUUID:     "txn-uuid-1",
						CreatedTime: expectedTimestamp.Add(2 * time.Millisecond),
						ClosedTime:  expectedTimestamp.Add(3 * time.Millisecond),
						Amount:      3250000,
						Payload: payload{
							Bet:                      "zero",
							Round:                    "CG-909",
							Token:                    "token-1",
							Amount:                   3250000,
							GameID:                   121,
							IsFree:                   false,
							Currency:                 "GBP",
							GameCode:                 "game-code",
							RequestUUID:              "req-uuid-1",
							RoundClosed:              true,
							SupplierUser:             "6",
							TransactionUUID:          "txn-uuid-1",
							ReferenceTransactionUUID: "txn-uuid-0",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			var resp transactionResponse
			err := json.Unmarshal([]byte(test.response), &resp)
			assert.NoError(t, err)
			assert.Equal(tt, test.want, resp)
		})
	}
}
