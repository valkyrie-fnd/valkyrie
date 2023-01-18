package pam

import (
	"bytes"
	"encoding/gob"

	"github.com/shopspring/decimal"
)

// Amt is then internal alias for the Decimal type and
// it is injected into generated model for the Generic PAM.
// NB: use `Amt` internally instead of the generated pam.Amount
type Amt decimal.Decimal

const decimalPlaces = 6

func (a Amt) String() string {
	return decimal.Decimal(a).String()
}

func (a Amt) Equal(b Amt) bool {
	return decimal.Decimal(a).Equal(decimal.Decimal(b))
}

var Zero Amt = Amt(decimal.Zero)

var ZeroAmount = Amount(decimal.Zero)

func (m *Amount) ToAmt() Amt {
	return Amt(*m)
}

func (m *Amount) Copy() Amount {
	return Amount(decimal.Decimal(*m).Copy())
}

func (m *Amount) Add(v Amount) Amount {
	return Amount(decimal.Decimal(*m).Add(decimal.Decimal(v)))
}

func (m *Amount) Sub(v Amount) Amount {
	return Amount(decimal.Decimal(*m).Sub(decimal.Decimal(v)))
}

// MarshalJSON provides custom marshall of the Amount type
func (m Amount) MarshalJSON() ([]byte, error) {
	return []byte(decimal.Decimal(m).StringFixed(decimalPlaces)), nil
}

// UnmarshalJSON provides custom unmarshal of the Amount
func (m *Amount) UnmarshalJSON(data []byte) error {
	d := decimal.Decimal(*m)
	err := d.UnmarshalJSON(data)
	*m = Amount(d)
	return err
}

// GobDecode provides custom decoding since gob halts on type aliases
func (m *Amount) GobDecode(bs []byte) error {
	buf := bytes.NewBuffer(bs)
	dec := gob.NewDecoder(buf)
	var str string
	err := dec.Decode(&str)
	if err != nil {
		return err
	}

	val, err := decimal.NewFromString(str)
	if err != nil {
		return err
	}

	*m = Amount(val)
	return nil
}

// GobDecode provides custom encoding since gob halts on type aliases
func (m Amount) GobEncode() ([]byte, error) {
	var sink bytes.Buffer
	enc := gob.NewEncoder(&sink)
	err := enc.Encode(m.ToAmt().String())
	return sink.Bytes(), err
}
