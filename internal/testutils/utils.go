// Package testutils contains shared helpers for unit tests and integrations tests in
// the provider scope
package testutils

import (
	"crypto/rand"
	"encoding/base64"
	"net"
	"os"

	"github.com/shopspring/decimal"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

// GetFreePort returns a free open port that is ready to use.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func RandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Ptr returns the pointer to an argument, useful for string literals.
func Ptr[T any](t T) *T {
	return &t
}

func EnvOrDefault(environmentKey, defaultValue string) string {
	if v, found := os.LookupEnv(environmentKey); found {
		return v
	} else {
		return defaultValue
	}
}

func Get[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func NewFloatAmount(val float64) pam.Amount {
	return pam.Amount(decimal.NewFromFloat(val))
}
