package evolution

import (
	"github.com/shopspring/decimal"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

// Amount alias for use in evolution integration providing
// primarily custom wire format
type Amount pam.Amt

var ZeroAmount = Amount(pam.Zero)

func (a *Amount) toAmt() pam.Amt {
	return pam.Amt(*a)
}

func (a Amount) Equal(b Amount) bool {
	return decimal.Decimal(a).Equal(decimal.Decimal(b))
}

func fromPamAmount(a *pam.Amount) Amount {
	return Amount(a.ToAmt())
}

func (m *Amount) MarshalJSON() ([]byte, error) {
	return []byte(decimal.Decimal(*m).StringFixed(6)), nil
}

func (m *Amount) UnmarshalJSON(data []byte) error {
	d := decimal.Decimal(*m)
	err := d.UnmarshalJSON(data)
	*m = Amount(d)
	return err
}
