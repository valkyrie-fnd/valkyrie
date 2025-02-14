package pam

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
)

func TestSixDecimalRounder(t *testing.T) {
	tests := []struct {
		name     string
		in       Amt
		expected Amt
		err      assert.ErrorAssertionFunc
	}{
		{"no rounding, no error", floatAmt(123), floatAmt(123), assert.NoError},
		{"dropping insignificant decimals is fine", floatAmt(1.1234560000000), floatAmt(1.123456), assert.NoError},
		{"error when precision is lost", floatAmt(1.1234560000001), floatAmt(1.123456), assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := SixDecimalRounder(tt.in)
			if !tt.err(t, err) {
				assert.True(t, tt.expected.Equal(res.ToAmt()))
			}
		})
	}
}

func floatAmt(val float64) Amt {
	return Amt(decimal.NewFromFloat(val))
}
