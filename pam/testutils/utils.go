package testutils

import (
	"github.com/shopspring/decimal"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func NewFloatAmount(val float64) pam.Amount {
	return pam.Amount(decimal.NewFromFloat(val))
}

// Ptr returns the pointer to an argument, useful for string literals.
func Ptr[T any](t T) *T {
	return &t
}
