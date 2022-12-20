package evolution

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func TestProviderController_Balance(t *testing.T) {

	tests := []struct {
		name             string
		method           string
		requestBody      string
		stubFn           func() (*StandardResponse, error)
		stubCheckFn      func() (*CheckResponse, error)
		expectedResponse string
	}{
		{
			name:   "Balance example from specs",
			method: "/balance",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"game": null,
				"currency":"EUR",
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
			stubFn: func() (*StandardResponse, error) {
				return &StandardResponse{
					Status:  "OK",
					Balance: amountFromFloat(999.35),
					Bonus:   ZeroAmount,
					UUID:    "ce186440-ed92-11e3-ac10-0800200c9a66",
				}, nil
			},
			expectedResponse: `{
				"status":"OK",
				"balance":999.35,
				"bonus":0.00,
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
		{
			name:   "Unknown error",
			method: "/balance",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"game": null,
				"currency":"EUR"
			}`,
			stubFn: func() (*StandardResponse, error) {
				return nil, errors.New("oops, balance blew up")
			},
			expectedResponse: `{
				"status":"UNKNOWN_ERROR",
				"bonus":0.00,
				"balance":0.00
			}`,
		},
		{
			name:   "Check example from specs",
			method: "/check",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"channel":{
						"type":"P"
						},
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
			stubCheckFn: func() (*CheckResponse, error) {
				return &CheckResponse{
					Status: "OK",
					SID:    "sid-parameter-from-UserAuthentication-call",
					UUID:   "ce186440-ed92-11e3-ac10-0800200c9a66",
				}, nil
			},
			expectedResponse: `{
				"status":"OK",
				"sid":"sid-parameter-from-UserAuthentication-call",
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
		{
			name:   "Unknown error",
			method: "/check",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"game": null,
				"currency":"EUR"
			}`,
			stubCheckFn: func() (*CheckResponse, error) {
				return nil, errors.New("oops, check blew up")
			},
			expectedResponse: `{
				"status":"UNKNOWN_ERROR",
				"bonus":0.00,
				"balance":0.00
			}`,
		},
		{
			name:   "Debit example from specs",
			method: "/debit",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"currency":"EUR",
				"game":{
					"id":"7kfwqku4jb4mtas1n4k4irqa",
					"type":"blackjack",
					"details" : {
						"table" : {
						"id" : "aaabbbcccdddeee111",
						"vid" : "aaabbbcccdddeee111"
						}
					}
				},
				"transaction":{
					"id":"9AotBIvi23",
					"refId":"1459zzz",
					"amount":1.556179
					},
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
			stubFn: func() (*StandardResponse, error) {
				return &StandardResponse{
					Status:  "OK",
					Balance: amountFromFloat(999.35),
					Bonus:   amountFromFloat(1.0),
					UUID:    "ce186440-ed92-11e3-ac10-0800200c9a66",
				}, nil
			},
			expectedResponse: `{
				"status":"OK",
				"balance":999.35,
				"bonus":1.00,
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
		{
			name:   "Credit example from specs",
			method: "/credit",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"currency":"EUR",
				"game":{
					"id":"7kfwqku4jb4mtas1n4k4irqa",
					"type":"blackjack",
					"details" : {
						"table" : {
						"id" : "aaabbbcccdddeee111",
						"vid" : "aaabbbcccdddeee111"
						}
					}
				},
				"transaction":{
					"id":"K4FOqh17v0",
					"refId":"1459zzz",
					"amount":1.556179
					},
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
			stubFn: func() (*StandardResponse, error) {
				return &StandardResponse{
					Status:  "OK",
					Balance: amountFromFloat(999.35),
					Bonus:   amountFromFloat(1.0),
					UUID:    "ce186440-ed92-11e3-ac10-0800200c9a66",
				}, nil
			},
			expectedResponse: `{
				"status":"OK",
				"balance":999.35,
				"bonus":1.00,
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
		{
			name:   "Cancel example from specs",
			method: "/cancel",
			requestBody: `{
				"transaction": {
					"id": "9AotBIvi23",
					"refId": "1459zzz",
					"amount": 1.556179
				},
				"sid": "sid-parameter-from-UserAuthentication-call",
				"userId": "euID-parameter-from-UserAuthentication-call",
				"uuid": "ce186440-ed92-11e3-ac10-0800200c9a66",
				"currency": "EUR",
				"game": {
					"id": "7kfwqku4jb4mtas1n4k4irqa",
					"type": "blackjack",
					"details": {
					"table": {
						"id": "aaabbbcccdddeee111",
						"vid": "aaabbbcccdddeee111"
					}
					}
				}
			}`,
			stubFn: func() (*StandardResponse, error) {
				return &StandardResponse{
					Status:  "OK",
					Balance: amountFromFloat(999.35),
					Bonus:   amountFromFloat(1.0),
					UUID:    "ce186440-ed92-11e3-ac10-0800200c9a66",
				}, nil
			},
			expectedResponse: `{
				"status":"OK",
				"balance":999.35,
				"bonus":1.00,
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
		{
			name:   "Rejected cancel",
			method: "/cancel",
			requestBody: `{
				"transaction": {
					"id": "9AotBIvi23",
					"refId": "1459zzz",
					"amount": 1.556179
				},
				"sid": "sid-parameter-from-UserAuthentication-call",
				"userId": "euID-parameter-from-UserAuthentication-call",
				"uuid": "xxx",
				"currency": "EUR",
				"game": {
					"id": "7kfwqku4jb4mtas1n4k4irqa",
					"type": "blackjack",
					"details": {
					"table": {
						"id": "aaabbbcccdddeee111",
						"vid": "aaabbbcccdddeee111"
					}
					}
				}
			}`,
			stubFn: func() (*StandardResponse, error) {
				return nil, toProviderError(pam.ValkyrieError{ValkErrorCode: pam.ValkErrAlreadySettled}, "xxx", amountFromFloat(9.9), amountFromFloat(1.1))
			},
			expectedResponse: `{
				"status":"BET_ALREADY_SETTLED",
				"balance":9.9,
				"bonus":1.1,
				"uuid":"xxx"
			}`,
		},
		{
			name:   "Promo payout example from specs",
			method: "/promo_payout",
			requestBody: `{
				"sid": "sid-parameter-from-UserAuthentication-call",
				"userId": "euID-parameter-from-UserAuthentication-call",
				"currency": "EUR",
				"game": null,
				"promoTransaction": {
					"type": "FreeRoundPlayableSpent",
					"id": "TD1459zzz",
					"amount": 1.556179,
					"voucherId": "d82ce074-bd5a-11eb-8529-0242ac130003",
					"remainingRounds": 0
				},
				"uuid": "ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
			stubFn: func() (*StandardResponse, error) {
				return &StandardResponse{
					Status:  "OK",
					Balance: amountFromFloat(999.35),
					Bonus:   amountFromFloat(1.0),
					UUID:    "ce186440-ed92-11e3-ac10-0800200c9a66",
				}, nil
			},
			expectedResponse: `{
				"status":"OK",
				"balance":999.35,
				"bonus":1.00,
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
	}

	serviceStub := stubService{}
	controller := NewProviderController(&serviceStub)

	app := fiber.New()
	app.Post("check", func(c *fiber.Ctx) error {
		return controller.Check(c)
	})
	app.Post("balance", func(c *fiber.Ctx) error {
		return controller.Balance(c)
	})
	app.Post("debit", func(c *fiber.Ctx) error {
		return controller.Debit(c)
	})
	app.Post("credit", func(c *fiber.Ctx) error {
		return controller.Debit(c)
	})
	app.Post("cancel", func(c *fiber.Ctx) error {
		return controller.Cancel(c)
	})
	app.Post("promo_payout", func(c *fiber.Ctx) error {
		return controller.PromoPayout(c)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceStub.fn = tt.stubFn
			serviceStub.checkFn = tt.stubCheckFn

			req := httptest.NewRequest(http.MethodPost, tt.method, strings.NewReader(tt.requestBody))
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			resp, err := app.Test(req)
			require.NoError(t, err)
			body, _ := io.ReadAll(resp.Body)
			assert.JSONEq(t, tt.expectedResponse, string(body))
		})
	}
}

func Test_request_validation(t *testing.T) {
	serviceStub := stubService{
		fn: func() (*StandardResponse, error) {
			return nil, errors.New("should not get here")
		},
	}
	controller := NewProviderController(&serviceStub)
	app := fiber.New()
	app.Post("check", func(c *fiber.Ctx) error {
		return controller.Check(c)
	})
	app.Post("balance", func(c *fiber.Ctx) error {
		return controller.Balance(c)
	})
	app.Post("debit", func(c *fiber.Ctx) error {
		return controller.Debit(c)
	})
	app.Post("credit", func(c *fiber.Ctx) error {
		return controller.Debit(c)
	})
	app.Post("cancel", func(c *fiber.Ctx) error {
		return controller.Cancel(c)
	})
	app.Post("promo_payout", func(c *fiber.Ctx) error {
		return controller.PromoPayout(c)
	})

	tests := []struct {
		name        string
		method      string
		requestBody string
	}{
		{
			name:   "Broken balance request",
			method: "/balance",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
		{
			name:   "broken debit",
			method: "/debit",
			requestBody: `{
				"sid":"sid-parameter-from-UserAuthentication-call",
				"userId":"euID-parameter-from-UserAuthentication-call",
				"currency":"E",
				"game":{
					"id":"7kfwqku4jb4mtas1n4k4irqa",
					"type":"blackjack",
					"details" : {
						"table" : {
						"id" : "aaabbbcccdddeee111",
						"vid" : "aaabbbcccdddeee111"
						}
					}
				},
				"transaction":{
					"id":"9AotBIvi23",
					"refId":"1459zzz",
					"amount":1.556179
					},
				"uuid":"ce186440-ed92-11e3-ac10-0800200c9a66"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.method, strings.NewReader(tt.requestBody))
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			resp, _ := app.Test(req)

			require.Equal(t, 400, resp.StatusCode)
			// body, _ := io.ReadAll(resp.Body)
		})
	}
}

type stubService struct {
	fn      func() (*StandardResponse, error)
	checkFn func() (*CheckResponse, error)
}

func (s *stubService) Check(CheckRequest) (*CheckResponse, error) {
	return s.checkFn()
}

func (s *stubService) Balance(BalanceRequest) (*StandardResponse, error) {
	return s.fn()
}

func (s *stubService) Credit(CreditRequest) (*StandardResponse, error) {
	return s.fn()
}

func (s *stubService) Debit(DebitRequest) (*StandardResponse, error) {
	return s.fn()
}

func (s *stubService) Cancel(CancelRequest) (*StandardResponse, error) {
	return s.fn()
}

func (s *stubService) PromoPayout(PromoPayoutRequest) (*StandardResponse, error) {
	return s.fn()
}

func (s *stubService) WithContext(context.Context) Service {
	return s
}
