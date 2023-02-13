package redtiger

import (
	"encoding/json"
	"testing"

	"github.com/valkyrie-fnd/valkyrie/pam/testutils"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func toMoney(val float64) Money {
	return Money(testutils.NewFloatAmount(val))
}

func toJackpotMoney(val float64) JackpotMoney {
	return JackpotMoney(testutils.NewFloatAmount(val))
}

// rtMoney should be marshaled to fix regex: \d+\.\d{2}
func Test_rtMoney_marshal(t *testing.T) {
	tests := []struct {
		name string
		sut  Balance
		want string
	}{
		{"0 decimals should produce 2 decimals", Balance{toMoney(10), toMoney(11)}, `{"cash": "10.00", "bonus": "11.00"}`},
		{"1 decimals should produce 2 decimals", Balance{toMoney(10.1), toMoney(11.0)}, `{"cash": "10.10", "bonus": "11.00"}`},
		{"2 decimals should produce 2 decimals", Balance{toMoney(10.01), toMoney(11.11)}, `{"cash": "10.01", "bonus": "11.11"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			res, err := json.Marshal(test.sut)
			assert.NoError(tt, err, "should not produce error")
			assert.JSONEq(tt, test.want, string(res))
		})
	}
}

func Test_rtMoney_unmarshal(t *testing.T) {
	tests := []struct {
		name string
		want Balance
		sut  string
	}{
		{
			"zeroth decimals should parse into values",
			Balance{toMoney(10), toMoney(11)},
			`{"cash": "10.00", "bonus": "11.00"}`,
		},
		{
			"tenth decimals should parse into values",
			Balance{toMoney(10.10), toMoney(11.0)},
			`{"cash": "10.10", "bonus": "11.00"}`,
		},
		{
			"hundreds decimals should parse into values",
			Balance{toMoney(10.01), toMoney(11.11)},
			`{"cash": "10.01", "bonus": "11.11"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			var res Balance
			err := json.Unmarshal([]byte(test.sut), &res)
			assert.NoError(tt, err, "should not produce error")
			assert.Equal(tt, test.want.Bonus.string(), res.Bonus.string())
			assert.Equal(tt, test.want.Cash.string(), res.Cash.string())
		})
	}
}

func Test_rtJackpotMoney_marshal(t *testing.T) {
	testSources := []struct {
		name string
		sut  Sources
		want string
	}{
		{
			"0 decimals should produce 6 decimals for jackpotMap",
			Sources{Lines: toMoney(10), Features: toMoney(11), Jackpot: map[string]JackpotMoney{"super": toJackpotMoney(10)}},
			`{"lines": "10.00", "features": "11.00", "jackpot": {"super": "10.000000"}}`,
		},
		{
			"1 decimals should produce 6 decimals for jackpotMap",
			Sources{Lines: toMoney(10.1), Features: toMoney(11.0), Jackpot: map[string]JackpotMoney{"super": toJackpotMoney(10.1)}},
			`{"lines": "10.10", "features": "11.00", "jackpot": {"super": "10.100000"}}`},
	}
	for _, test := range testSources {
		t.Run(test.name, func(tt *testing.T) {
			res, err := json.Marshal(test.sut)
			assert.NoError(tt, err, "should not produce error")
			assert.JSONEq(tt, test.want, string(res))
		})
	}
}

func Test_rtJackpotMoney_unmarshal(t *testing.T) {
	testSources := []struct {
		name string
		want Sources
		sut  string
	}{
		{
			"zeroth decimals should parse into values",
			Sources{Lines: toMoney(10), Features: toMoney(11), Jackpot: map[string]JackpotMoney{"super": toJackpotMoney(10)}},
			`{"lines": "10.00", "features": "11.00", "jackpot": {"super": "10.000000"}}`,
		},
		{
			"tenth decimals should parse into values",
			Sources{Lines: toMoney(10.1), Features: toMoney(11.0), Jackpot: map[string]JackpotMoney{"super": toJackpotMoney(10.1)}},
			`{"lines": "10.10", "features": "11.00", "jackpot": {"super": "10.100000"}}`,
		},
	}
	for _, test := range testSources {
		t.Run(test.name, func(tt *testing.T) {
			var res Sources
			err := json.Unmarshal([]byte(test.sut), &res)
			assert.NoError(tt, err, "should not produce error")
			assert.True(tt, test.want.Jackpot["super"].Equal(res.Jackpot["super"]))
		})
	}
}

func (m Money) string() string {
	return decimal.Decimal(m).String()
}
