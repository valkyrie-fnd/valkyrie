package redtiger

import (
	"github.com/shopspring/decimal"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type Money pam.Amt
type JackpotMoney pam.Amt

func zeroMoney() Money {
	return Money(pam.ZeroAmount)
}

func (m Money) Equal(b Money) bool {
	return pam.Amt(m).Equal(pam.Amt(b))
}

func (m JackpotMoney) Equal(b JackpotMoney) bool {
	return pam.Amt(m).Equal(pam.Amt(b))
}

func (m JackpotMoney) MarshalJSON() ([]byte, error) {
	str := "\"" + decimal.Decimal(m).StringFixed(6) + "\""
	return []byte(str), nil
}

func (m *JackpotMoney) UnmarshalJSON(data []byte) error {
	d := decimal.Decimal(*m)
	err := d.UnmarshalJSON(data)
	*m = JackpotMoney(d)
	return err
}

func (m Money) toAmount() pam.Amount {
	return pam.Amount(m)
}

func (m Money) MarshalJSON() ([]byte, error) {
	str := "\"" + decimal.Decimal(m).StringFixed(2) + "\""
	return []byte(str), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	d := decimal.Decimal(*m)
	err := d.UnmarshalJSON(data)
	*m = Money(d)
	return err
}

type GameLaunchRequest struct {
	Token      string `url:"token"`
	Currency   string `url:"currency,omitempty"`
	UserID     string `url:"userId,omitempty"`
	LobbyURL   string `url:"lobbyUrl,omitempty"`
	DepositURL string `url:"depositUrl,omitempty"`
	Affiliate  string `url:"affiliate,omitempty"`
	Channel    string `url:"channel,omitempty"`
	Casino     string `url:"casino,omitempty"`
}

type rtGameLaunchConfig struct {
	PlayMode                   string `url:"playMode" validate:"required"`
	RealityCheckHistoryURL     string `url:"realityCheckHistoryUrl,omitempty"`
	RealityCheckLobbyURL       string `url:"realityCheckLobbyUrl,omitempty"`
	RealityCheckElapsedMinutes int    `url:"realityCheckElapsedMinutes,omitempty"`
	RealityCheckMinutes        int    `url:"realityCheckMinutes,omitempty"`
	HasAutoplayLimitLoss       bool   `url:"hasAutoplayLimitLoss"`
	HasHistory                 bool   `url:"hasHistory"`
	HasAutoplaySingleWinLimit  bool   `url:"hasAutoplaySingleWinLimit"`
	HasAutoplayStopOnJackpot   bool   `url:"hasAutoplayStopOnJackpot"`
	HasAutoplayStopOnBonus     bool   `url:"hasAutoplayStopOnBonus"`
	HasAutoplayTotalSpins      bool   `url:"hasAutoplayTotalSpins"`
	HasFreeBets                bool   `url:"hasFreeBets"`
	FullScreen                 bool   `url:"fullScreen"`
	HasRoundID                 bool   `url:"hasRoundId"`
	HasRealPlayButton          bool   `url:"hasRealPlayButton"`
}

// RT base request and response
type BaseRequest struct {
	Token    string `json:"token" validate:"required,min=32,max=128"`
	UserID   string `json:"userId" validate:"max=36"`
	Casino   string `json:"casino" validate:"max=50"`
	Currency string `json:"currency" validate:"max=8"`
	IP       string `json:"ip"`
}

type BaseResponse struct {
	Token    string `json:"token" validate:"min=32,max=128"`
	Currency string `json:"currency" validate:"max=8"`
}

// RT refund stuff
type RefundRequest struct {
	BaseRequest
	Transaction TransactionStake `json:"transaction"`
	Game        Game             `json:"game"`
	Round       Round            `json:"round"`
	Promo       Promo            `json:"promo"`
}

type RefundResult struct {
	Token string  `json:"token"`
	ID    string  `json:"id"`
	Stake Balance `json:"stake"`
}

type RefundResponseWrapper struct {
	Error    *Error       `json:"error,omitempty"`
	Result   RefundResult `json:"result,omitempty"`
	Balance  Balance      `json:"balance"`
	Currency string       `json:"currency"`
	Response
}

// RT payout stuff
type PayoutRequest struct {
	BaseRequest
	Transaction TransactionPayout `json:"transaction"`
	Game        Game              `json:"game"`
	Jackpot     Jackpot           `json:"jackpot"`
	Round       Round             `json:"round"`
	Promo       Promo             `json:"promo"`
	Retry       bool              `json:"retry"`
}

type TransactionPayout struct {
	Sources     Sources       `json:"sources"`
	Details     PayoutDetails `json:"details"`
	ID          string        `json:"id" validate:"max=32"`
	Payout      Money         `json:"payout" validate:"min=0"`
	PayoutPromo Money         `json:"payoutPromo" validate:"min=0"`
}

type PayoutDetails struct {
	Game    Money        `json:"game"`
	Jackpot JackpotMoney `json:"jackpot"`
}

type Sources struct {
	Jackpot  map[string]JackpotMoney `json:"jackpot"`
	Lines    Money                   `json:"lines"`
	Features Money                   `json:"features"`
}

type PayoutResponse struct {
	BaseResponse
	ID      string  `json:"id"`
	Payout  Balance `json:"payout"`
	Balance Balance `json:"balance"`
}

type PayoutResponseWrapper struct {
	Error  *Error         `json:"error,omitempty"`
	Result PayoutResponse `json:"result,omitempty"`
	Response
}

// RT stake stuff

type StakeRequest struct {
	BaseRequest
	Transaction TransactionStake `json:"transaction"`
	Game        Game             `json:"game"`
	Round       Round            `json:"round"`
	Promo       Promo            `json:"promo,omitempty"`
}

type StakeResponse struct {
	BaseResponse
	ID      string  `json:"id"`
	Stake   Balance `json:"stake"`
	Balance Balance `json:"balance"`
}

type StakeResponseWrapper struct {
	Error  *Error        `json:"error,omitempty"`
	Result StakeResponse `json:"result,omitempty"`
	Response
}

type Response struct {
	Success bool `json:"success"`
}

type StakeDetails struct {
	Game    Money `json:"game"`
	Jackpot Money `json:"jackpot"`
}

type TransactionStake struct {
	ID         string       `json:"id" validate:"max=32"`
	Stake      Money        `json:"stake"`
	StakePromo Money        `json:"stakePromo"`
	Details    StakeDetails `json:"details"`
}

type Round struct {
	ID     string `json:"id" validate:"max=32"`
	Starts bool   `json:"starts"`
	Ends   bool   `json:"ends"`
}

type Game struct {
	Type    string `json:"type"`
	Key     string `json:"key" validate:"max=128"`
	Version string `json:"version" validate:"max=128"`
}

type Jackpot struct {
	Group        string   `json:"group" validate:"max=100"`
	Contribution string   `json:"contribution"`
	Pots         []string `json:"pots"`
}

type Promo struct {
	Type         string `json:"type"`
	InstanceCode string `json:"instanceCode" validate:"max=64"`
	CampaignCode string `json:"campaignCode" validate:"max=64"`
	InstanceID   int    `json:"instanceId"`
	CampaignID   int    `json:"campaignId"`
}

// RT auth request
type AuthRequest struct {
	BaseRequest
	Channel   string `json:"channel,omitempty" validate:"max=8"`
	Affiliate string `json:"affiliate,omitempty" validate:"max=255"`
	Extras    string `json:"extras,omitempty"`
}

// RT balance object
type Balance struct {
	Cash  Money `json:"cash"`
	Bonus Money `json:"bonus"`
}

// RT auth result object
type AuthResponse struct {
	BaseResponse
	UserID   string  `json:"userId" validate:"max=32"`
	Casino   string  `json:"casino" validate:"max=50"`
	Country  string  `json:"country" validate:"len=2"`
	Language string  `json:"language" validate:"max=6" `
	Balance  Balance `json:"balance"`
}

// RT auth response
type AuthResponseWrapper struct {
	Error   *Error       `json:"error,omitempty"`
	Result  AuthResponse `json:"result,omitempty"`
	Success bool         `json:"success"`
}

type Error struct {
	Message string      `json:"message"`
	Code    RTErrorCode `json:"code"`
}

// RT error response
type ErrorResponse struct {
	Error   Error `json:"error"`
	Success bool  `json:"success"`
}
