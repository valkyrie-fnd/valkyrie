package caleta

import (
	"time"

	"github.com/mitchellh/mapstructure"
)

type gameRoundRenderResponse struct {
	InlineResponse200
	Message string
	Code    int
}

type transactionRequestBody struct {
	RoundID    string `json:"round_id"`
	OperatorID string `json:"operator_id"`
}

type transactionResponse struct {
	RoundTransactions *[]roundTransaction `json:"transactions,omitempty"`
	Message           string              `json:"message"`
	RoundID           string              `json:"round_id"`
	Code              int                 `json:"code"`
}

type roundTransaction struct {
	CreatedTime  time.Time   `json:"created_time"`
	ClosedTime   time.Time   `json:"closed_time"`
	TxnUUID      string      `json:"txn_uuid"`
	Payload      payload     `json:"payload"`
	ID           int         `json:"id"`
	RoundID      int         `json:"round_id"`
	TxnType      int         `json:"txn_type"`
	Status       int         `json:"status"`
	CacheEntryID int         `json:"cache_entry_id"`
	Amount       MoneyAmount `json:"amount"`
}

type payload struct {
	ReferenceTransactionUUID *TransactionUuid `json:"reference_transaction_uuid,omitempty"`
	Bet                      string           `json:"bet"`
	Round                    Round            `json:"round"`
	Token                    Token            `json:"token"`
	Currency                 Currency         `json:"currency"`
	GameCode                 GameCode         `json:"game_code"`
	RequestUUID              RequestUuid      `json:"request_uuid"`
	SupplierUser             SupplierUser     `json:"supplier_user"`
	TransactionUUID          TransactionUuid  `json:"transaction_uuid"`
	GameID                   GameId           `json:"game_id"`
	JackpotContribution      MoneyAmount      `json:"jackpot_contribution"`
	Amount                   MoneyAmount      `json:"amount"`
	RoundClosed              RoundClosed      `json:"round_closed"`
	IsFree                   IsFree           `json:"is_free"`
}

type gameURLQuery struct {
	Country      Country      `url:"country"`
	Currency     Currency     `url:"currency"`
	DepositURL   string       `url:"deposit_url,omitempty"`
	GameCode     GameCode     `url:"game_code"`
	Lang         Language     `url:"lang"`
	LobbyURL     string       `url:"lobby_url,omitempty"`
	OperatorID   OperatorId   `url:"operator_id"`
	SubPartnerID SubPartnerId `url:"sub_partner_id,omitempty"`
	Token        Token        `url:"token,omitempty"`
	User         User         `url:"user,omitempty"`
}

type caletaLaunchConfig struct {
	DepositURL   *string `mapstructure:"deposit_url,omitempty"`
	LobbyURL     string  `mapstructure:"lobby_url,omitempty"`
	SubPartnerID string  `mapstructure:"sub_partner_id,omitempty"`
}

func getLaunchConfig(input map[string]interface{}) (*caletaLaunchConfig, error) {
	config := caletaLaunchConfig{}
	err := mapstructure.Decode(input, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
