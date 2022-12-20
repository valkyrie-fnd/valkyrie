package caleta

import (
	"github.com/shopspring/decimal"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

const iMulti = 100000

var dMulti = decimal.NewFromInt(iMulti)

func fromPamAmount(a pam.Amount) *int {
	f, _ := decimal.Decimal(a).Float64()
	i := int(f * iMulti)
	return &i
}

func toPamAmount(i *int) pam.Amount {
	amt := decimal.NewFromInt(int64(*i)).Div(dMulti)
	return pam.Amount(amt)
}
