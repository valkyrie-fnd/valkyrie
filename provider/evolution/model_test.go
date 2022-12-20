package evolution

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func amountFromFloat(val float64) Amount {
	return Amount(decimal.NewFromFloat(val))
}

func TestAmountMarshall(t *testing.T) {
	trans := Transaction{
		ID:     "transID",
		RefID:  "refID",
		Amount: amountFromFloat(123.123456),
	}

	bs, err := json.Marshal(&trans)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"id":"transID", "refId":"refID", "amount": 123.123456}`, string(bs))

	var t2 Transaction
	err = json.Unmarshal(bs, &t2)

	assert.NoError(t, err)

	assert.Equal(t, trans, t2)
}
