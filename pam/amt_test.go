package pam

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAmount_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       Amount
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"marshal",
			floatAmount(1.234567),
			[]byte("1.234567"),
			assert.NoError,
		},
		{
			"marshal negative number",
			floatAmount(-47),
			[]byte("-47.000000"),
			assert.NoError,
		},
		{
			"marshal rounds excess decimals",
			floatAmount(1.23456789),
			[]byte("1.234568"),
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if !tt.wantErr(t, err, "MarshalJSON()") {
				return
			}
			assert.Equalf(t, tt.want, got, "MarshalJSON()")
		})
	}
}

func TestAmount_ToAmt(t *testing.T) {
	amount := floatAmount(1.23)

	amt := amount.ToAmt()

	assert.Equal(t, "1.23", amt.String())
}

func TestAmount_Equal(t *testing.T) {
	assert.True(t, floatAmt(1.23).Equal(floatAmt(1.23)))

	assert.False(t, floatAmt(1.23).Equal(floatAmt(3.21)))
}

func TestAmount_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		want    Amount
		data    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"unmarshal",
			floatAmount(1.234567),
			[]byte("1.234567"),
			assert.NoError,
		},
		{
			"unmarshal negative number",
			floatAmount(-47),
			[]byte("-47.000000"),
			assert.NoError,
		},
		{
			"unmarshal keeps excess decimals",
			floatAmount(1.23456789),
			[]byte("1.23456789"),
			assert.NoError,
		},
		{
			"unmarshal invalid number",
			Amount{},
			[]byte("foo"),
			assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Amount{}
			if !tt.wantErr(t, m.UnmarshalJSON(tt.data), fmt.Sprintf("UnmarshalJSON(%v)", tt.data)) {
				assert.True(t, tt.want.ToAmt().Equal(m.ToAmt()))
			}
		})
	}
}

func TestAmount_Copy(t *testing.T) {
	amount := floatAmount(1.23)

	otherAmount := amount.Copy()

	assert.False(t, &amount == &otherAmount, "pointers should be different")
	assert.True(t, amountEqual(amount, otherAmount), "values should be equal")
}

func TestAmount_Add(t *testing.T) {
	amount := floatAmount(1.5)
	amount = amount.Add(floatAmount(1.5))

	assert.True(t, amountEqual(amount, floatAmount(3.0)), "values should be equal")
}

func floatAmount(val float64) Amount {
	return Amount(decimal.NewFromFloat(val))
}

func amountEqual(a, b Amount) bool {
	return decimal.Decimal(a).Equal(decimal.Decimal(b))
}
