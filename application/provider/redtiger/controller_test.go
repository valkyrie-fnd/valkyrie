package redtiger

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestBaseRequestValidation(t *testing.T) {
	tests := []struct {
		name                 string
		contentType          string
		reqBody              string
		expectedResponseBody string
	}{
		{
			"No Content type result in error",
			"",
			`{}`,
			`{"success":false,"error":{"message":"Invalid input. err: Unprocessable Entity","code":200}}`,
		},
		{
			"Token is required",
			fiber.MIMEApplicationJSON,
			`{"currency":"SEK"}`,
			`{"success":false,"error":{"message":"Invalid input. err: Key: 'AuthRequest.BaseRequest.Token' Error:Field validation for 'Token' failed on the 'required' tag","code":200}}`,
		},
	}
	serviceStub := CtrlServiceStub{}
	ctrlSut := NewProviderController(&serviceStub)
	app := fiber.New()
	app.Post("auth", func(c *fiber.Ctx) error {
		return ctrlSut.Auth(c)
	})
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(test.reqBody))
			req.Header.Add(fiber.HeaderContentType, test.contentType)
			resp, _ := app.Test(req)

			body, _ := io.ReadAll(resp.Body)
			assert.JSONEq(t, test.expectedResponseBody, string(body))
		})
	}
}

type CtrlServiceStub struct{}

func (CtrlServiceStub) Auth(AuthRequest) (*AuthResponseWrapper, *ErrorResponse) {
	return nil, nil
}
func (CtrlServiceStub) Stake(StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
	return nil, nil
}
func (CtrlServiceStub) Payout(PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
	return nil, nil
}
func (CtrlServiceStub) Refund(RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
	return nil, nil
}
func (CtrlServiceStub) PromoBuyin(StakeRequest) (*StakeResponseWrapper, *ErrorResponse) {
	return nil, nil
}
func (CtrlServiceStub) PromoSettle(PayoutRequest) (*PayoutResponseWrapper, *ErrorResponse) {
	return nil, nil
}
func (CtrlServiceStub) PromoRefund(RefundRequest) (*RefundResponseWrapper, *ErrorResponse) {
	return nil, nil
}
func (s CtrlServiceStub) WithContext(context.Context) Service {
	return s
}
