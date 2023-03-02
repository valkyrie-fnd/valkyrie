package redtiger_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	redtiger_bd "github.com/valkyrie-fnd/valkyrie-stubs/backdoors/redtiger"

	"github.com/valkyrie-fnd/valkyrie/provider/redtiger"
)

// RGI client
type RGIClient struct {
	ProviderURL string
	APIKey      string
	backdoorURL string
	timeout     time.Duration
	baseRequest redtiger.BaseRequest
}

func NewRGIClient(providerURL, key, backdoorURL string) *RGIClient {
	return &RGIClient{
		ProviderURL: providerURL,
		APIKey:      key,
		backdoorURL: backdoorURL,
		timeout:     2 * time.Second,
	}
}

func (api *RGIClient) SetupSession(currency string) error {

	res, err := api.createSession(currency)
	if err != nil {
		return err
	}

	api.baseRequest = redtiger.BaseRequest{
		Token:    res.Token,
		UserID:   res.UserID,
		Casino:   "Kongo",
		Currency: currency,
		IP:       "1.2.3.4",
	}

	return nil
}

func (api *RGIClient) SetSessionToken(token string) error {
	api.baseRequest.Token = token

	return nil
}

func (api *RGIClient) createSession(currency string) (*redtiger_bd.Result, error) {
	req := redtiger_bd.SessionRequest{
		Currency: &currency,
	}

	a := fiber.Post(fmt.Sprintf("%s/redtiger/session", api.backdoorURL)).
		Timeout(api.timeout).
		JSON(&req)

	var resp redtiger_bd.SessionResponse
	if c, b, errs := a.Struct(&resp); c != fiber.StatusOK {
		return nil, errors.Join(append(errs, fmt.Errorf("redtiger/session request failed: %s", b))...)
	} else if !resp.Success {
		return nil, fmt.Errorf("redtiger/session request failed")
	}

	return &resp.Result, nil
}

func (api *RGIClient) ResetSession() error {
	a := fiber.Post(fmt.Sprintf("%s/datastore/session/reset", api.backdoorURL))

	if c, _, errs := a.Bytes(); c != fiber.StatusOK {
		var errStr string
		for _, e := range errs {
			errStr = fmt.Sprintf("%s - %s", errStr, e.Error())
		}
		return fmt.Errorf("Error resetting session: %s", errStr)
	}
	return nil
}

func (api *RGIClient) Auth() (*redtiger.AuthResponseWrapper, error) {

	req := redtiger.AuthRequest{
		BaseRequest: redtiger.BaseRequest{
			Token:    api.baseRequest.Token,
			Casino:   api.baseRequest.Casino,
			IP:       api.baseRequest.IP,
			Currency: api.baseRequest.Currency,
			UserID:   api.baseRequest.UserID,
		},
		Channel:   "APP",
		Affiliate: "",
	}

	path := "/auth"

	a := fiber.Post(fmt.Sprintf("%s%s", api.ProviderURL, path)).Add("Authorization", api.APIKey).
		Timeout(api.timeout).
		JSON(req)

	var result redtiger.AuthResponse
	resp := redtiger.AuthResponseWrapper{
		Result: result,
	}

	status, b, err := a.Struct(&resp)

	if status != fiber.StatusOK {
		return nil, fmt.Errorf("redtiger%s request failed with status [%v]: %s, err: %s", path, status, b, err)
	} else if err != nil {
		return nil, fmt.Errorf("redtiger%s request failed with error and status [%v]: %s", path, status, err)
	}

	api.baseRequest.Token = resp.Result.Token

	return &resp, nil
}

func (api *RGIClient) Stake(gameKey, transID, roundID string, amt, amtPromo redtiger.Money, promo *redtiger.Promo) (*redtiger.StakeResponseWrapper, error) {
	req := redtiger.StakeRequest{
		BaseRequest: api.baseRequest,
		Transaction: redtiger.TransactionStake{
			ID:         transID,
			Stake:      amt,
			StakePromo: amtPromo,
			Details: redtiger.StakeDetails{
				Game:    amt,
				Jackpot: toMoney(0.0),
			},
		},
		Round: redtiger.Round{
			ID:     roundID,
			Starts: true,
			Ends:   false,
		},
		Game: redtiger.Game{
			Key: gameKey,
		},
	}
	if promo != nil {
		req.Promo = *promo
	}

	path := "/stake"

	a := fiber.Post(api.ProviderURL+path).Add("Authorization", api.APIKey).
		Timeout(api.timeout).
		JSON(req)

	var resp redtiger.StakeResponseWrapper
	status, b, err := a.Struct(&resp)
	return &resp, wrapError(path, status, b, err, resp.Error)
}

func (api *RGIClient) PromoBuyin(gameKey, transID, roundID string, amt, amtPromo redtiger.Money) (*redtiger.StakeResponseWrapper, error) {
	req := redtiger.StakeRequest{
		BaseRequest: api.baseRequest,
		Transaction: redtiger.TransactionStake{
			ID:         transID,
			Stake:      amt,
			StakePromo: amtPromo,
			Details: redtiger.StakeDetails{
				Game:    amt,
				Jackpot: toMoney(0.0),
			},
		},
		Round: redtiger.Round{
			ID:     roundID,
			Starts: true,
			Ends:   false,
		},
		Game: redtiger.Game{
			Key: gameKey,
		},
	}
	path := "/promo/buyin"

	a := fiber.Post(api.ProviderURL+path).Add("Authorization", api.APIKey).
		Timeout(api.timeout).
		JSON(req)

	var resp redtiger.StakeResponseWrapper
	status, b, err := a.Struct(&resp)

	return &resp, wrapError(path, status, b, err, resp.Error)
}

func (api *RGIClient) Payout(gameKey, transID, roundID string, amt, promoAmt, gameAmt redtiger.Money, jackpotAmt redtiger.JackpotMoney) (*redtiger.PayoutResponseWrapper, error) {
	req := redtiger.PayoutRequest{
		BaseRequest: api.baseRequest,
		Transaction: redtiger.TransactionPayout{
			ID:          transID,
			Payout:      amt,
			PayoutPromo: promoAmt,
			Details: redtiger.PayoutDetails{
				Game:    gameAmt,
				Jackpot: jackpotAmt,
			},
			Sources: redtiger.Sources{
				Lines:    amt,
				Features: amt,
			},
		},
		Round: redtiger.Round{
			ID:     roundID,
			Starts: false,
			Ends:   true,
		},
		Game: redtiger.Game{
			Type: "slot",
			Key:  gameKey,
		},
		Retry: false,
	}

	path := "/payout"

	a := fiber.Post(api.ProviderURL+path).Add("Authorization", api.APIKey).
		Timeout(api.timeout).
		JSON(req)

	var resp redtiger.PayoutResponseWrapper
	status, b, err := a.Struct(&resp)
	return &resp, wrapError(path, status, b, err, resp.Error)
}

func (api *RGIClient) PromoSettle(gameKey, transID, roundID string, amt, promoAmt, gameAmt redtiger.Money, jackpotAmt redtiger.JackpotMoney) (*redtiger.PayoutResponseWrapper, error) {
	req := redtiger.PayoutRequest{
		BaseRequest: api.baseRequest,
		Transaction: redtiger.TransactionPayout{
			ID:          transID,
			Payout:      amt,
			PayoutPromo: promoAmt,
			Details: redtiger.PayoutDetails{
				Game:    gameAmt,
				Jackpot: jackpotAmt,
			},
			Sources: redtiger.Sources{
				Lines:    amt,
				Features: amt,
			},
		},
		Round: redtiger.Round{
			ID:     roundID,
			Starts: false,
			Ends:   true,
		},
		Game: redtiger.Game{
			Type: "slot",
			Key:  gameKey,
		},
		Retry: false,
	}

	path := "/promo/settle"

	a := fiber.Post(api.ProviderURL+path).Add("Authorization", api.APIKey).
		Timeout(api.timeout).
		JSON(req)

	var resp redtiger.PayoutResponseWrapper
	status, b, err := a.Struct(&resp)
	return &resp, wrapError(path, status, b, err, resp.Error)
}

func (api *RGIClient) Refund(transID, gameID, roundID string, amt redtiger.Money) (*redtiger.RefundResponseWrapper, error) {
	req := redtiger.RefundRequest{
		BaseRequest: api.baseRequest,
		Transaction: redtiger.TransactionStake{
			ID:         transID,
			Stake:      amt,
			StakePromo: toMoney(0.0),
			Details: redtiger.StakeDetails{
				Game:    amt,
				Jackpot: toMoney(0.0),
			},
		},
		Round: redtiger.Round{
			ID:     roundID,
			Starts: true,
			Ends:   false,
		},
		Game: redtiger.Game{
			Type: "slot",
			Key:  gameID,
		},
	}

	path := "/refund"

	a := fiber.Post(api.ProviderURL+path).Add("Authorization", api.APIKey).
		Timeout(api.timeout).
		JSON(req)

	var resp redtiger.RefundResponseWrapper
	status, b, err := a.Struct(&resp)
	return &resp, wrapError(path, status, b, err, resp.Error)
}

func (api *RGIClient) PromoRefund(transID, gameID, roundID string, amt redtiger.Money) (*redtiger.RefundResponseWrapper, error) {
	req := redtiger.RefundRequest{
		BaseRequest: api.baseRequest,
		Transaction: redtiger.TransactionStake{
			ID:         transID,
			Stake:      amt,
			StakePromo: amt,
			Details: redtiger.StakeDetails{
				Game:    amt,
				Jackpot: toMoney(0.0),
			},
		},
		Round: redtiger.Round{
			ID:     roundID,
			Starts: true,
			Ends:   false,
		},
		Game: redtiger.Game{
			Type: "slot",
			Key:  gameID,
		},
	}

	path := "/promo/refund"

	a := fiber.Post(api.ProviderURL+path).Add("Authorization", api.APIKey).
		Timeout(api.timeout).
		JSON(req)

	var resp redtiger.RefundResponseWrapper

	status, b, err := a.Struct(&resp)

	return &resp, wrapError(path, status, b, err, resp.Error)
}

func wrapError(path string, code int, b []byte, err []error, rtErr *redtiger.Error) error {
	if code != fiber.StatusOK {
		return fmt.Errorf("redtiger%s request failed with status [%v]: %s, Error: %s", path, code, b, err)
	} else if err != nil {
		return fmt.Errorf("redtiger%s request failed with status [%v]: %s, Error: %s", path, code, b, err)
	}
	if rtErr != nil {
		return fmt.Errorf("redtiger%s request failed with status [%v]: %s", path, code, rtErr.Message)
	}

	return nil
}
