package genericpam

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

const (
	DriverName     = "generic"
	settlementType = pam.MIXED
)

func init() {
	pam.ClientFactory().
		Register(DriverName, func(args pam.ClientArgs) (pam.PamClient, error) {
			return Create(args.Config, args.Client)
		})
}

type genericPamConfig struct {
	Name   string `mapstructure:"name"`
	URL    string `mapstructure:"url"`
	APIKey string `mapstructure:"api_key"`
}

type GenericPam struct {
	rest    rest.HTTPClientJSONInterface
	baseURL string
	apiKey  string
}

func Create(cfg configs.PamConf, client rest.HTTPClientJSONInterface) (*GenericPam, error) {
	config, err := pam.GetConfig[genericPamConfig](cfg)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Creating %s pam client", config.Name)

	return &GenericPam{
		baseURL: config.URL,
		apiKey:  config.APIKey,
		rest:    client,
	}, nil
}

func getHeaders(apiKey, sessionToken, correlationID string) map[string]string {
	if correlationID == "" {
		correlationID = "-"
		log.Trace().Msg("no correlationID set, defaulting to '-'")
	}
	return map[string]string{
		"Authorization":    fmt.Sprintf("Bearer %s", apiKey),
		"X-Player-Token":   sessionToken,
		"X-Correlation-ID": correlationID,
	}
}

func (c *GenericPam) RefreshSession(rm pam.RefreshSessionRequestMapper) (*pam.Session, error) {
	ctx, r, err := rm()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/players/session", c.baseURL)
	var resp pam.SessionResponse
	headers := getHeaders(c.apiKey, r.Params.XPlayerToken, r.Params.XCorrelationID)
	req := &rest.HTTPRequest{
		URL:     url,
		Headers: headers,
		Query:   map[string]string{"provider": r.Params.Provider},
	}

	err = c.rest.PutJSON(ctx, req, &resp)
	if err = handleErrors(resp.Error, err, resp.Session); err != nil {
		return nil, err
	}

	return resp.Session, nil
}

func (c *GenericPam) GetBalance(rm pam.GetBalanceRequestMapper) (*pam.Balance, error) {
	ctx, r, err := rm()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/players/%s/balance", c.baseURL, r.PlayerID)
	resp := pam.BalanceResponse{}
	headers := getHeaders(c.apiKey, r.Params.XPlayerToken, r.Params.XCorrelationID)
	req := &rest.HTTPRequest{
		URL:     url,
		Headers: headers,
		Query:   map[string]string{"provider": r.Params.Provider},
	}

	err = c.rest.GetJSON(ctx, req, &resp)
	if err = handleErrors(resp.Error, err, resp.Balance); err != nil {
		return nil, err
	}

	return resp.Balance, nil
}

func (c *GenericPam) GetTransactions(rm pam.GetTransactionsRequestMapper) ([]pam.Transaction, error) {
	ctx, r, err := rm()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/players/%s/transactions", c.baseURL, r.PlayerID)
	var resp pam.GetTransactionsResponse
	headers := getHeaders(c.apiKey, r.Params.XPlayerToken, r.Params.XCorrelationID)
	query := map[string]string{"provider": r.Params.Provider}
	if r.Params.ProviderTransactionId != nil {
		query["providerTransactionId"] = *r.Params.ProviderTransactionId
	}
	if r.Params.ProviderBetRef != nil {
		query["providerBetRef"] = *r.Params.ProviderBetRef
	}
	req := &rest.HTTPRequest{
		URL:     url,
		Headers: headers,
		Query:   query,
	}

	err = c.rest.GetJSON(ctx, req, &resp)
	if err = handleErrors(resp.Error, err, resp.Transactions); err != nil {
		return nil, err
	}
	if len(*resp.Transactions) == 0 {
		return nil, pam.ValkyrieError{ValkErrorCode: pam.ValkErrOpTransNotFound, ErrMsg: "No transactions"}
	}

	return *resp.Transactions, nil
}

func (c *GenericPam) AddTransaction(rm pam.AddTransactionRequestMapper) (*pam.TransactionResult, error) {
	ctx, r, err := rm(pam.SixDecimalRounder)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/players/%s/transactions", c.baseURL, r.PlayerID)
	var resp pam.AddTransactionResponse
	headers := getHeaders(c.apiKey, r.Params.XPlayerToken, r.Params.XCorrelationID)
	req := &rest.HTTPRequest{
		URL:     url,
		Headers: headers,
		Query:   map[string]string{"provider": r.Params.Provider},
		Body:    &r.Body,
	}

	err = c.rest.PostJSON(ctx, req, &resp)
	if err = handleErrors(resp.Error, err, resp.TransactionResult); err != nil {
		if resp.TransactionResult != nil {
			// Special case, balance may still be included even if add transaction resulted in error.
			return resp.TransactionResult, err
		}
		return nil, err
	}

	return resp.TransactionResult, nil
}

func (c *GenericPam) GetGameRound(rm pam.GetGameRoundRequestMapper) (*pam.GameRound, error) {
	ctx, r, err := rm()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/players/%s/gamerounds/%s", c.baseURL, r.PlayerID, r.ProviderRoundID)
	var resp pam.GameRoundResponse
	headers := getHeaders(c.apiKey, r.Params.XPlayerToken, r.Params.XCorrelationID)
	req := &rest.HTTPRequest{
		URL:     url,
		Headers: headers,
		Query:   map[string]string{"provider": r.Params.Provider},
	}

	err = c.rest.GetJSON(ctx, req, &resp)
	if err = handleErrors(resp.Error, err, resp.Gameround); err != nil {
		return nil, err
	}

	return resp.Gameround, nil
}

func (c *GenericPam) GetSession(rm pam.GetSessionRequestMapper) (*pam.Session, error) {
	ctx, r, err := rm()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/players/session", c.baseURL)
	var resp pam.SessionResponse
	headers := getHeaders(c.apiKey, r.Params.XPlayerToken, r.Params.XCorrelationID)
	req := &rest.HTTPRequest{
		URL:     url,
		Headers: headers,
		Query:   map[string]string{"provider": r.Params.Provider},
	}

	err = c.rest.GetJSON(ctx, req, &resp)
	if err = handleErrors(resp.Error, err, resp.Session); err != nil {
		return nil, err
	}
	return resp.Session, nil
}

func (c *GenericPam) GetSettlementType() pam.SettlementType {
	return settlementType
}

func (c *GenericPam) GetTransactionHandling() pam.TransactionHandling {
	return pam.OPERATOR
}

// handleErrors does general error handling for a response and returns
// the most detailed error, or nil if no errors found.
func handleErrors[T any](pamError *pam.PamError, httpErr error, entity *T) error {
	if pamError != nil {
		// PamError has precedence since it contains more detailed error info from remote pam.
		return pam.ToValkyrieError(pamError)
	}
	if httpErr != nil {
		if errors.Is(httpErr, rest.TimeoutError) {
			return pam.ErrorWrapper("http client timeout", pam.ValkErrTimeout, httpErr)
		}
		return pam.ErrorWrapper("http client error", pam.ValkErrUndefined, httpErr)
	}
	if entity == nil {
		return pam.ValkyrieError{ValkErrorCode: pam.ValkErrUndefined, ErrMsg: "nil entity"}
	}
	return nil
}
