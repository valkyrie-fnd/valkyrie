// Very thin wrapper around http client doing requests to the operator backend
// https://studio.evolutiongaming.com/api/userauthentication/docs/v2/index.html
package evolution_test

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/provider/evolution"
)

// Client calling licensee RGI
type EvoRGIClient struct {
	url       string
	sidURL    string
	reqBase   evolution.RequestBase
	authToken string
}

func NewEvo(baseURL, sidURL, token string) *EvoRGIClient {
	return &EvoRGIClient{baseURL, sidURL, evolution.RequestBase{
		UUID: uuid.NewString(),
	}, token}
}

func (c *EvoRGIClient) String() string {
	return fmt.Sprintf("Session [%s], UserId [%s]", c.reqBase.SID, c.reqBase.UserID)
}

func (c *EvoRGIClient) SID(userID string, channel rune) (*evolution.CheckResponse, error) {
	req := evolution.CheckRequest{
		Channel: struct {
			Type string `json:"type"`
		}{string(channel)},
		RequestBase: evolution.RequestBase{
			UserID: userID,
			SID:    testutils.RandomString(3),
			UUID:   uuid.NewString(),
		},
	}

	a := fiber.Post(c.sidURL + "/evolution/sid").
		QueryString(fmt.Sprintf("authToken=%s", c.authToken)).
		Timeout(2 * time.Second).
		JSON(req)
	var resp evolution.CheckResponse
	status, b, err := a.Struct(&resp)

	if status != fiber.StatusOK {
		return nil, fmt.Errorf("evo/sid request failed with status [%v]: %s", status, err)
	} else if err != nil {
		return nil, testutils.Stack(err, fmt.Errorf("evo/sid request failed: %s", b))
	}

	// Store the session info in the client for subsequent interaction
	c.reqBase = req.RequestBase
	// Update Sid if provided in the response
	if resp.SID != "" {
		c.reqBase.SID = resp.SID
	}

	return &resp, nil
}

func (c *EvoRGIClient) Check(r evolution.CheckRequest) (*evolution.CheckResponse, error) {
	a := fiber.Post(c.url + "/check").
		QueryString(fmt.Sprintf("authToken=%s", c.authToken)).
		Timeout(2 * time.Second).
		JSON(&r)
	var resp evolution.CheckResponse
	status, b, err := a.Struct(&resp)

	if status != fiber.StatusOK {
		return nil, fmt.Errorf("evo/check request failed with status [%v]: %s", status, err)
	} else if err != nil {
		return nil, testutils.Stack(err, fmt.Errorf("evo/check request failed: %s", b))
	}

	// Store the session info in the client for subsequent interaction
	c.reqBase = r.RequestBase
	// Update Sid if provided in the response
	if resp.SID != "" {
		c.reqBase.SID = resp.SID
	}

	return &resp, nil
}

func (c *EvoRGIClient) Balance(curr string) (*evolution.StandardResponse, error) {
	r := evolution.BalanceRequest{
		RequestBase: c.reqBase,
		Currency:    curr,
	}
	return post(c.url, "/balance", c.authToken, &r)
}

func (c *EvoRGIClient) Debit(curr string, game evolution.Game, trans evolution.Transaction) (*evolution.StandardResponse, error) {
	r := evolution.DebitRequest{
		RequestBase: c.reqBase,
		Currency:    curr,
		Game:        game,
		Transaction: trans,
	}
	return post(c.url, "/debit", c.authToken, &r)
}

func (c *EvoRGIClient) Credit(curr string, game evolution.Game, trans evolution.Transaction) (*evolution.StandardResponse, error) {
	r := evolution.CreditRequest{
		RequestBase: c.reqBase,
		Currency:    curr,
		Game:        game,
		Transaction: trans,
	}
	return post(c.url, "/credit", c.authToken, &r)
}

func (c *EvoRGIClient) Cancel(curr string, game evolution.Game, trans evolution.Transaction) (*evolution.StandardResponse, error) {
	r := evolution.DebitRequest{
		RequestBase: c.reqBase,
		Currency:    curr,
		Game:        game,
		Transaction: trans,
	}
	return post(c.url, "/cancel", c.authToken, &r)
}

func (c *EvoRGIClient) PromoPayout(curr string, game evolution.Game, trans evolution.PromoTransaction) (*evolution.StandardResponse, error) {
	r := evolution.PromoPayoutRequest{
		RequestBase:      c.reqBase,
		Currency:         curr,
		Game:             game,
		PromoTransaction: trans,
	}
	return post(c.url, "/promo_payout", c.authToken, &r)
}

// Sugared method for "placing bet" the Evo way
func (c *EvoRGIClient) placeBet(amt float64, tableID, transID, gameRoundID, refID string, checks ...resultCheck) string {
	// For duplicates and cancels - use the same transId as for the original transaction
	if transID == "" {
		transID = testutils.RandomString(5)
	}
	r, err := sendTrans(amt, tableID, gameRoundID, transID, refID, c.Debit)
	for _, check := range checks {
		check(r, err)
	}
	return transID
}

// Sugared method for "cancel bet" the Evo way
func (c *EvoRGIClient) cancelBet(amt float64, tableID, gameRoundID, transID, refID string, checks ...resultCheck) {
	r, err := sendTrans(amt, tableID, gameRoundID, transID, refID, c.Cancel)
	for _, check := range checks {
		check(r, err)
	}
}

// Sugared method for "End Game" the Evo way
func (c *EvoRGIClient) settleBet(payout float64, tableID, gameRoundID, refID string, checks ...resultCheck) string {
	transID := testutils.RandomString(5)
	r, err := sendTrans(payout, tableID, gameRoundID, transID, refID, c.Credit)
	for _, check := range checks {
		check(r, err)
	}
	return transID
}

// Sugared method for "End Game" the Evo way
func (c *EvoRGIClient) promoPayout(payout float64, tableID, transID, gameRoundID string, checks ...resultCheck) string {
	if transID == "" {
		transID = testutils.RandomString(5)
	}

	game := evolution.Game{
		ID:   gameRoundID,
		Type: "blackjack",
		Details: evolution.GameDetails{
			Table: evolution.GameTable{
				ID: tableID,
			},
		},
	}
	trans := evolution.PromoTransaction{
		ID:     transID,
		Amount: fromFloat(payout),
	}

	r, err := c.PromoPayout(currency, game, trans)
	for _, check := range checks {
		check(r, err)
	}
	return transID
}

func post(base, path, token string, body interface{}) (*evolution.StandardResponse, error) {
	a := fiber.Post(base + path).
		QueryString(fmt.Sprintf("authToken=%s", token)).
		Timeout(5 * time.Second).
		JSON(body)

	var resp evolution.StandardResponse
	status, b, err := a.Struct(&resp)
	if status != fiber.StatusOK {
		return nil, fmt.Errorf("evo/%s request failed with status [%v]: %s, Error: %s", path, status, string(b), err)
	} else if err != nil {
		return nil, testutils.Stack(err, fmt.Errorf("evo/%s request failed: %s", path, b))
	}

	return &resp, nil
}
