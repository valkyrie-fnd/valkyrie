package example

import "github.com/valkyrie-fnd/valkyrie/pam"

type errorResponse struct {
	Err string `json:"error" validate:"required"`
}

func (e errorResponse) Error() string {
	return e.Err
}

type baseRequest struct {
	Token string `json:"token" validate:"required"`
}

// balanceRequest follow game provider request structure.
type balanceRequest struct {
	baseRequest
	Currency string `json:"currency" validate:"required"`
}

type balanceResponse struct {
	Balance pam.Amount `json:"balance" validate:"required"`
}

type authRequest struct {
	baseRequest
}

type authResponse struct {
	Token string `json:"token" validate:"required"`
}

type betRequest struct {
	baseRequest
	Amount        pam.Amount `json:"amount" validate:"required"`
	Currency      string     `json:"currency" validate:"required"`
	GameID        string     `json:"gameId" validate:"required"`
	RoundID       string     `json:"roundId" validate:"required"`
	TransactionID string     `json:"transactionId" validate:"required"`
}

type betResponse struct {
	Balance pam.Amount `json:"balance" validate:"required"`
}

type winRequest struct {
	baseRequest
	Amount        pam.Amount `json:"amount" validate:"required"`
	Currency      string     `json:"currency" validate:"required"`
	GameID        string     `json:"gameId" validate:"required"`
	RoundID       string     `json:"roundId" validate:"required"`
	TransactionID string     `json:"transactionId" validate:"required"`
}

type winResponse struct {
	Balance pam.Amount `json:"balance" validate:"required"`
}
