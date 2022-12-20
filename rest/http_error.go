package rest

import (
	"errors"
	"fmt"
)

// HTTPError error returned from Client
type HTTPError struct {
	Message string
	Code    int
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func NewHTTPError(code int, msg string) error {
	return HTTPError{Message: msg, Code: code}
}

var TimeoutError = errors.New("client request timeout")
