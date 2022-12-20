package pam

import (
	"fmt"

	"github.com/shopspring/decimal"
)

var SixDecimalRounder AmountRounder = func(amt Amt) (*Amount, error) {
	orig := decimal.Decimal(amt)
	b := orig.Round(6)
	if !orig.Sub(b).Equal(decimal.Zero) {
		return nil, fmt.Errorf("rounding will result in lost precision %s -> %s", amt, b)
	}
	res := Amount(b)
	return &res, nil
}
