package caleta_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	uid "github.com/google/uuid"
	"github.com/valkyrie-fnd/valkyrie-stubs/backdoors"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/provider/caleta"
	"github.com/valkyrie-fnd/valkyrie/provider/caleta/auth"
)

var emptyMap = make(map[string]string)

var nowTimestamp = func() map[string]string {
	return map[string]string{"X-Msg-Timestamp": time.Now().UTC().Format(caleta.TimestampFormat)}
}

const (
	RoundOpen   caleta.RoundClosed = false
	RoundClosed caleta.RoundClosed = true
)

// RGI client
type RGIClient struct {
	providerURL string
	backdoorURL string
	timeout     time.Duration
	signer      auth.Signer
	session     string
	userID      string
}

func NewRGIClient(providerURL, backdoorURL string, signer auth.Signer) *RGIClient {
	return &RGIClient{
		providerURL: providerURL,
		backdoorURL: backdoorURL,
		timeout:     2 * time.Second,
		signer:      signer,
	}
}

func (api *RGIClient) SetupSession(currency string) error {
	req := backdoors.SessionRequest{
		Currency:    &currency,
		CashAmount:  testutils.Ptr[float64](initialCashBalance),
		PromoAmount: testutils.Ptr[float64](initialPromoBalance),
		Provider:    caleta.ProviderName,
	}

	a := fiber.Post(fmt.Sprintf("%s/session", api.backdoorURL)).
		Timeout(api.timeout).
		JSON(&req)

	var resp backdoors.SessionResponse
	if c, b, errs := a.Struct(&resp); c != fiber.StatusOK {
		return errors.Join(append(errs, fmt.Errorf("session request failed: %s", b))...)
	} else if !resp.Success {
		return fmt.Errorf("session request failed")
	}

	api.session = resp.Result.Token
	api.userID = resp.Result.UserID

	return nil
}

func (api *RGIClient) BlockAccount(currency string) error {
	req := backdoors.SessionRequest{
		UserID:    testutils.Ptr(api.userID),
		Currency:  &currency,
		Provider:  caleta.ProviderName,
		IsBlocked: testutils.Ptr(true),
	}

	a := fiber.Post(fmt.Sprintf("%s/session", api.backdoorURL)).
		Timeout(api.timeout).
		JSON(&req)

	var resp backdoors.SessionResponse
	if c, b, errs := a.Struct(&resp); c != fiber.StatusOK {
		return errors.Join(append(errs, fmt.Errorf("session request failed: %s", b))...)
	} else if !resp.Success {
		return fmt.Errorf("session request failed")
	}

	api.session = resp.Result.Token

	return nil
}

func (api *RGIClient) setSession(session string) {
	api.session = session
}

// Change the initial token received on /game/url for a new one that will be used on wallet transactions.
func (api *RGIClient) Check() (*caleta.InlineResponse2001, error) {
	body := caleta.WalletCheckBody{
		Token: api.session,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/check")
	resp, err := postJSON[caleta.WalletCheckBody, caleta.InlineResponse2001](url, body, api.timeout, api.signer, emptyMap)
	if err != nil {
		return nil, err
	}

	api.session = *resp.Token

	return resp, nil
}

// Called when player's balance is needed. Operator is expected to return player's current balance.
// Game id is provided to help Operator with player's activity statistics.
func (api *RGIClient) Balance(gameCode string) (*caleta.BalanceResponse, error) {
	body := caleta.WalletBalanceBody{
		GameCode: gameCode,
		// GameId: deprecated field
		RequestUuid:  uuid(),
		SupplierUser: api.userID,
		Token:        api.session,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/balance")
	resp, err := postJSON[caleta.WalletBalanceBody, caleta.BalanceResponse](url, body, api.timeout, api.signer, emptyMap)
	if err != nil {
		return nil, err
	}
	if body.RequestUuid != resp.RequestUuid {
		return nil, fmt.Errorf("request UUID changed")
	}
	return resp, nil
}

// Called when the User places a bet (debit). Operator is expected to decrease player's balance by amount and return new balance.
// Each bet has transaction_uuid which is unique identifier of this transaction. Before altering of User's balance,
// Operator has to check that bet wasn't processed before.
//
// There might be Retry Policy: In case of network fail (HTTP 502, timeout, nxdomain, etc.), we will retry 3 times with 1 sec of timeout.
// If we do not receive 200 HTTP status, this transaction will be counted as failed and there is no rollback for this operation.
func (api *RGIClient) Bet(gameCode, currency, round, transactionID string, amount int) (*caleta.BalanceResponse, error) {

	body := caleta.WalletBetBody{
		// Bet:             nil,
		// CampaignUuid:    nil,
		// GameId:          nil,
		// RewardUuid:      nil,
		Amount:          amount,
		Currency:        caleta.Currency(currency),
		GameCode:        gameCode,
		IsFree:          false,
		RequestUuid:     uuid(),
		Round:           round,
		RoundClosed:     false,
		SupplierUser:    api.userID,
		Token:           api.session,
		TransactionUuid: transactionID,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/bet")
	resp, err := postJSON[caleta.WalletBetBody, caleta.BalanceResponse](url, body, api.timeout, api.signer, nowTimestamp())
	if err != nil {
		return nil, err
	}
	if body.RequestUuid != resp.RequestUuid {
		return nil, fmt.Errorf("request UUID changed")
	}
	return resp, nil
}

func (api *RGIClient) PromoBet(gameCode, currency, round, transactionID string, amount int) (*caleta.BalanceResponse, error) {

	// Promo differs slightly from regular non-promo body, setting the CampaignUuid, IsFree and RewardUuid
	body := caleta.WalletBetBody{
		// Bet:             nil,
		Amount:          amount,
		CampaignUuid:    testutils.Ptr(uuid()),
		Currency:        caleta.Currency(currency),
		GameCode:        gameCode,
		IsFree:          true,
		RequestUuid:     uuid(),
		RewardUuid:      testutils.Ptr(uuid()),
		Round:           round,
		RoundClosed:     false,
		SupplierUser:    api.userID,
		Token:           api.session,
		TransactionUuid: transactionID,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/bet")
	resp, err := postJSON[caleta.WalletBetBody, caleta.BalanceResponse](url, body, api.timeout, api.signer, nowTimestamp())
	if err != nil {
		return nil, err
	}
	if body.RequestUuid != resp.RequestUuid {
		return nil, fmt.Errorf("request UUID changed")
	}
	return resp, nil
}

// Called when the User wins (credit). Operator is expected to increase player's balance by amount and return new balance.
// reference_transaction_uuid show to which bet this win is related. Each win has transaction_uuid which is unique identifier of this transaction.
// Before any altering of User's balance, Operator has to check that win wasn't processed before.
//
// Retry Policy: In case of network fail (HTTP 502, timeout, nxdomain, etc.) we will retry 3 times with 1 sec of timeout.
// The rest of retry logic is left to provider's RGS: the retries may continue indefinitely or the bet may be rolled back, and the money returned back to user.
func (api *RGIClient) Win(gameCode, currency, round, refTransactionID, transactionID string, amount int) (*caleta.BalanceResponse, error) {

	body := caleta.WalletWinBody{
		// CampaignUuid:             nil,
		// Bet:                      nil,
		// GameId:                   nil,
		// RewardUuid:               nil,
		Amount:                   amount,
		Currency:                 caleta.Currency(currency),
		GameCode:                 gameCode,
		IsFree:                   false,
		ReferenceTransactionUuid: refTransactionID,
		RequestUuid:              uuid(),
		Round:                    round,
		RoundClosed:              true,
		SupplierUser:             api.userID,
		Token:                    api.session,
		TransactionUuid:          transactionID,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/win")
	resp, err := postJSON[caleta.WalletWinBody, caleta.BalanceResponse](url, body, api.timeout, api.signer, nowTimestamp())
	if err != nil {
		return nil, err
	}
	if body.RequestUuid != resp.RequestUuid {
		return nil, fmt.Errorf("request UUID changed")
	}
	return resp, nil
}

func (api *RGIClient) PromoWin(gameCode, currency, round, refTransactionID, transactionID string, amount int) (*caleta.BalanceResponse, error) {

	// Promo differs slightly from regular non-promo body, setting the CampaignUuid, IsFree and RewardUuid
	body := caleta.WalletWinBody{
		// Bet:                      nil,
		// GameId:                   nil,
		Amount:                   amount,
		Currency:                 caleta.Currency(currency),
		CampaignUuid:             testutils.Ptr(uuid()),
		GameCode:                 gameCode,
		IsFree:                   true,
		ReferenceTransactionUuid: refTransactionID,
		RequestUuid:              uuid(),
		RewardUuid:               testutils.Ptr(uuid()),
		Round:                    round,
		RoundClosed:              true,
		SupplierUser:             api.userID,
		Token:                    api.session,
		TransactionUuid:          transactionID,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/win")
	resp, err := postJSON[caleta.WalletWinBody, caleta.BalanceResponse](url, body, api.timeout, api.signer, nowTimestamp())
	if err != nil {
		return nil, err
	}
	if body.RequestUuid != resp.RequestUuid {
		return nil, fmt.Errorf("request UUID changed")
	}
	return resp, nil
}

// Called when there is need to roll back the effect of the referenced transaction.
// Operator is expected to find referenced transaction, roll back its effects and return the player's new balance.
func (api *RGIClient) Rollback(gameCode, round, refTransactionID, transactionID string, roundStatus caleta.RoundClosed) (*caleta.BalanceResponse, error) {

	body := caleta.WalletRollbackBody{
		//GameId:                   nil,
		//IsFree:                   nil,
		GameCode:                 gameCode,
		ReferenceTransactionUuid: refTransactionID,
		RequestUuid:              uuid(),
		Round:                    round,
		RoundClosed:              roundStatus,
		User:                     &api.userID,
		Token:                    api.session,
		TransactionUuid:          transactionID,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/rollback")
	resp, err := postJSON[caleta.WalletRollbackBody, caleta.BalanceResponse](url, body, api.timeout, api.signer, nowTimestamp())
	if err != nil {
		return nil, err
	}
	if body.RequestUuid != resp.RequestUuid {
		return nil, fmt.Errorf("request UUID changed")
	}
	return resp, nil
}

func (api *RGIClient) PromoRollback(gameCode, round, refTransactionID, transactionID string, roundStatus caleta.RoundClosed) (*caleta.BalanceResponse, error) {

	// Promo differs slightly from regular non-promo body, setting IsFree
	body := caleta.WalletRollbackBody{
		//GameId:                   nil,
		GameCode:                 gameCode,
		IsFree:                   testutils.Ptr(true),
		ReferenceTransactionUuid: refTransactionID,
		RequestUuid:              uuid(),
		Round:                    round,
		RoundClosed:              roundStatus,
		User:                     &api.userID,
		Token:                    api.session,
		TransactionUuid:          transactionID,
	}

	url := fmt.Sprintf("%s%s", api.providerURL, "/wallet/rollback")
	resp, err := postJSON[caleta.WalletRollbackBody, caleta.BalanceResponse](url, body, api.timeout, api.signer, nowTimestamp())
	if err != nil {
		return nil, err
	}
	if body.RequestUuid != resp.RequestUuid {
		return nil, fmt.Errorf("request UUID changed")
	}
	return resp, nil
}

func postJSON[Body any, Response any](url string, body Body, timeout time.Duration, signer auth.Signer,
	additionalHeaders map[string]string) (*Response, error) {

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	sign, err := signer.Sign(bodyBytes)
	if err != nil {
		return nil, err
	}

	a := fiber.Post(url).
		ContentType(fiber.MIMEApplicationJSON).
		Body(bodyBytes).
		SetBytesV("X-Auth-Signature", sign).
		Timeout(timeout)

	for k, v := range additionalHeaders {
		a = a.Set(k, v)
	}

	var resp Response

	if status, b, errs := a.Struct(&resp); status != fiber.StatusOK || len(errs) > 0 {
		return nil, errors.Join(append(errs, fmt.Errorf("%s request failed with status [%v]: %s.", url, status, b))...)
	}

	return &resp, nil
}

func uuid() string {
	return uid.NewString()
}
