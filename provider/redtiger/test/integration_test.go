// Package redtiger_test contains integration tests for verifying the Red Tiger provider implementation.
package redtiger_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/valkyrie-fnd/valkyrie-stubs/datastore"
	"github.com/valkyrie-fnd/valkyrie-stubs/utils"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/provider/internal/test"

	_ "github.com/joho/godotenv/autoload" // load .env file automatically
	"github.com/shopspring/decimal"

	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/provider/redtiger"
)

const (
	currency = "USD"
)

var gameID = "SomeSlotGame"

type RedTigerIntegrationTestSuite struct {
	test.IntegrationTestSuite
	client         *RGIClient
	initialBalance *redtiger.Balance
}

// Runs all tests in suite below one by one
func TestSuite(t *testing.T) {
	providerConfigFn := func(ds datastore.ExtendedDatastore) configs.ProviderConf {
		apiKey, _ := ds.GetProviderAPIKey(redtiger.ProviderName)
		sessionKey := ds.GetProviderTokens()[redtiger.ProviderName]
		return configs.ProviderConf{
			Name:     redtiger.ProviderName,
			BasePath: "/redtiger",
			Auth: map[string]any{
				"api_key":     testutils.EnvOrDefault("RT_API_KEY", apiKey.APIKey),
				"recon_token": testutils.EnvOrDefault("RT_RECON_TOKEN", sessionKey),
			},
		}
	}

	suite.Run(t, &RedTigerIntegrationTestSuite{
		IntegrationTestSuite: test.IntegrationTestSuite{
			ProviderConfigFn: providerConfigFn,
		},
	})
}

func (s *RedTigerIntegrationTestSuite) SetupTest() {
	s.client, s.initialBalance = s.prepareCase()
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_FOR_USER_TOKEN() {
	s.client = NewRGIClient(s.ValkyrieURL, s.ProviderConfig.Auth["api_key"].(string), s.BackdoorURL)
	s.Require().NoError(s.client.SetupSession(currency))

	s.Assert().NotEmpty(s.client.baseRequest.Token, "a token should be set for the s.client")
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_AUTH_WITH_VALID_TOKEN() {
	s.client = NewRGIClient(s.ValkyrieURL, s.ProviderConfig.Auth["api_key"].(string), s.BackdoorURL)
	s.Require().NoError(s.client.SetupSession(currency))
	resp, err := s.client.Auth()
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_AUTH_WITH_INVALID_TOKEN() {
	s.client = NewRGIClient(s.ValkyrieURL, s.ProviderConfig.Auth["api_key"].(string), s.BackdoorURL)

	s.Require().NoError(s.client.SetupSession(currency), "Request should not produce hard error")
	s.Require().NoError(s.client.SetSessionToken("thistokenshouldnotwork-But-still-32-chars"))

	resp, err := s.client.Auth()
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().Nil(resp)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST() {
	resp, err := s.client.Stake(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST_WITH_EXISTING_TRANSACTION_ID() {
	roundID := rnd()
	transactionID := rnd()
	resp, err := s.client.Stake(gameID, transactionID, roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")
	balance := resp.Result.Balance.Cash

	// Firing twice should result in 200 OK but the balance should stay the same
	resp, err = s.client.Stake(gameID, transactionID, roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
	s.Assert().Equal(balance, resp.Result.Balance.Cash)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_WITH_AN_INVALID_USER_ID() {
	s.client.baseRequest.UserID = "invalid-user"
	resp, err := s.client.Stake(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), nil)
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.NotAuthorized, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST_WITH_AN_INVALID_CURRENCY() {
	s.client.baseRequest.Currency = "FIM"
	resp, err := s.client.Stake(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), nil)

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.InvalidUserCurrency, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST_WITH_A_NEGATIVE_STAKE() {
	resp, err := s.client.Stake(gameID, rnd(), rnd(), toMoney(-10.0), toMoney(0.0), nil)

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.InvalidInput, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST_WITH_A_GREATER_STAKE_THAN_BALANCE() {
	resp, err := s.client.Stake(gameID, rnd(), rnd(), addToMoney(s.initialBalance.Cash, 1), toMoney(0.0), nil)

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.InsufficientFunds, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST_WITH_STAKE_0() {
	resp, err := s.client.Stake(gameID, rnd(), rnd(), toMoney(0.0), toMoney(0.0), nil)

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
	s.Assert().True(decimal.Decimal(pam.Amt(resp.Result.Balance.Cash)).Equal(decimal.Decimal(toMoney(1000.0))))
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_SEVERAL_BET_REQUESTS_WITH_SAME_ROUNDID() {
	roundID := rnd()
	_, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")

	// Firing twice with the same round ID should be allowed as long as the round is not ended
	resp, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_SEVERAL_BET_REQUESTS_WITH_SAME_ROUNDID_AFTER_PAYOUT() {
	roundID := rnd()
	_, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")

	_, err = s.client.Payout(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().NoError(err, "Request should not produce hard error")

	// Firing twice with the same round ID should not be allowed if the round is is ended
	resp, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST_WITH_RECON_TOKEN() {
	s.client.baseRequest.Token = s.ProviderConfig.Auth["recon_token"].(string)
	resp, err := s.client.Stake(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), nil)
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.NotAuthorized, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BET_REQUEST_WITH_PROMO_PARAMS() {
	resp, err := s.client.Stake(gameID, rnd(), rnd(), toMoney(10.0), toMoney(5.0), &redtiger.Promo{})

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST() {
	roundID := rnd()
	// Creates a gameround
	_, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")

	resp, err := s.client.Payout(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_WITH_EXISTING_ID() {
	roundID := rnd()
	// Creates a gameround
	_, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")

	transID := rnd()

	resp, err := s.client.Payout(gameID, transID, roundID, toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().NoError(err, "Request should not produce hard error")
	balance := resp.Result.Balance.Cash
	// Firing twice should result in 200 OK but the balance should stay the same
	resp, err = s.client.Payout(gameID, transID, roundID, toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().Equal(balance, resp.Result.Balance.Cash)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_WITH_RECON_TOKEN() {
	roundID := rnd()
	_, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")
	s.client.baseRequest.Token = s.ProviderConfig.Auth["recon_token"].(string)
	resp, err := s.client.Payout(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_WITH_INVALID_USER_ID() {
	s.client.baseRequest.UserID = "invalid-user"

	resp, err := s.client.Payout(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.NotAuthorized, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_WITH_RECON_TOKEN_AND_INVALID_USER_ID() {
	s.client.baseRequest.UserID = "invalid-user"
	s.client.baseRequest.Token = s.ProviderConfig.Auth["recon_token"].(string)
	resp, err := s.client.Payout(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.NotAuthorized, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_WITH_INVALID_CURRENCY() {
	s.client.baseRequest.Currency = "FIM"
	resp, err := s.client.Payout(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.InvalidUserCurrency, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_A_NEGATIVE_PAYOUT() {
	resp, err := s.client.Payout(gameID, rnd(), rnd(), toMoney(-10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.InvalidInput, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_WITH_INVALID_TOKEN() {
	s.client.baseRequest.Token = "SomeInvalidToken-Still-needs-to-be-long"
	resp, err := s.client.Payout(gameID, rnd(), rnd(), toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.NotAuthorized, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_AN_AWARD_REQUEST_WITH_PROMO_PARAMS() {
	roundID := rnd()
	_, err := s.client.Stake(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")
	resp, err := s.client.Payout(gameID, rnd(), roundID, toMoney(10.0), toMoney(10.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_REFUND() {
	transID := rnd()
	roundID := rnd()
	stakeRes, err := s.client.Stake(gameID, transID, roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err)
	s.Assert().True(stakeRes.Success)
	// Assert that stake has debit:ed the cash balance
	s.Assert().Equal(subFromMoney(s.initialBalance.Cash, 10.0), stakeRes.Result.Balance.Cash)
	s.Assert().Equal(s.initialBalance.Bonus, stakeRes.Result.Balance.Bonus)

	refundRes, err := s.client.Refund(transID, gameID, roundID, toMoney(10.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(refundRes.Success)
	// Assert that refund has credit:ed the balance corresponding to stake
	s.Assert().Equal(s.initialBalance.Cash, refundRes.Balance.Cash)
	s.Assert().Equal(s.initialBalance.Bonus, refundRes.Balance.Bonus)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_REFUND_INVALID_AMOUNT_IGNORED() {
	transID := rnd()
	roundID := rnd()
	stakeRes, err := s.client.Stake(gameID, transID, roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err)
	s.Assert().True(stakeRes.Success)
	// Assert that stake has debit:ed the cash balance
	s.Assert().Equal(subFromMoney(s.initialBalance.Cash, 10.0), stakeRes.Result.Balance.Cash)
	s.Assert().Equal(s.initialBalance.Bonus, stakeRes.Result.Balance.Bonus)

	refundRes, err := s.client.Refund(transID, gameID, roundID, toMoney(47.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(refundRes.Success)
	// Assert that the amount from the stake with 'transID' will be used when cancelled, not the incorrectly
	// specified amount '47.0'
	s.Assert().Equal(s.initialBalance.Cash, refundRes.Balance.Cash)
	s.Assert().Equal(s.initialBalance.Bonus, refundRes.Balance.Bonus)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_REFUND_NON_EXISTING_TRANSACTION() {
	res, err := s.client.Refund(rnd(), gameID, rnd(), toMoney(10.0))
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(res.Success)
	s.Assert().Equal(redtiger.TransactionNotFound, res.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_REFUND_BUYIN_TRANSACTION() {
	transID := rnd()
	roundID := rnd()
	resp, _ := s.client.PromoBuyin(gameID, transID, roundID, toMoney(10.0), toMoney(0.0))
	s.Assert().Nil(resp.Error)
	res, err := s.client.PromoRefund(transID, gameID, roundID, toMoney(10.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(res.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_REFUND_A_REFUNDED_TRANSACTION() {
	transID := rnd()
	roundID := rnd()
	respStake, err := s.client.Stake(gameID, transID, roundID, toMoney(10.0), toMoney(0.0), nil)
	s.Require().NoError(err, "Request should not produce hard error")

	balanceAfterStake := respStake.Result.Balance.Cash
	resp, err := s.client.Refund(transID, gameID, roundID, toMoney(10.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
	balance := resp.Balance.Cash

	s.Assert().True(decimal.Decimal(pam.Amt(balance)).GreaterThan(decimal.Decimal(balanceAfterStake)))
	resp, err = s.client.Refund(transID, gameID, roundID, toMoney(10.0))
	balanceSecondCall := resp.Balance.Cash
	// Refunding the same twice does nothing to balance
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
	s.Assert().Equal(balance, balanceSecondCall)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_REFUND_WITH_AN_INVALID_TOKEN() {
	transID := rnd()
	roundID := rnd()
	_, _ = s.client.Stake(gameID, transID, roundID, toMoney(10.0), toMoney(0.0), nil)

	s.client.baseRequest.Token = "InvalidToken-but-long-enough-for-min-requirement"
	res, err := s.client.Refund(transID, gameID, roundID, toMoney(10.0))
	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(res.Success)
	s.Assert().Equal(redtiger.NotAuthorized, res.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_BUYIN_REQUEST() {
	resp, err := s.client.PromoBuyin(gameID, rnd(), rnd(), toMoney(10.0), toMoney(10.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_MAKE_BUYIN_WITH_A_NEGATIVE_STAKE() {
	resp, err := s.client.PromoBuyin(gameID, rnd(), rnd(), toMoney(-10.0), toMoney(-10.0))

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.InvalidInput, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_MAKE_BUYIN_WITH_AN_INVALID_TOKEN() {
	s.client.baseRequest.Token = "Invalid-token-with-min-requirement-32-chars"
	resp, err := s.client.PromoBuyin(gameID, rnd(), rnd(), toMoney(10.0), toMoney(10.0))

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.NotAuthorized, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_MAKE_BUYIN_WITH_USED_TRANSACTION_ID() {
	transID := rnd()
	roundID := rnd()
	resp, err := s.client.PromoBuyin(gameID, transID, roundID, toMoney(10.0), toMoney(10.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
	// Firing twice should result in 200 OK but the balance should stay the same
	balance := resp.Result.Balance.Cash

	resp, err = s.client.PromoBuyin(gameID, transID, roundID, toMoney(10.0), toMoney(10.0))
	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
	s.Assert().Equal(balance, resp.Result.Balance.Cash)
}

func (s *RedTigerIntegrationTestSuite) Test_PROMO_REFUND() {
	transID := rnd()
	roundID := rnd()
	resp, err := s.client.PromoBuyin(gameID, transID, roundID, toMoney(10.0), toMoney(10.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)

	respRef, errRef := s.client.PromoRefund(transID, gameID, roundID, toMoney(10.0))
	s.Require().NoError(errRef, "Request should not produce hard error")
	s.Assert().True(respRef.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_REFUND_BUYIN_WITH_AN_INVALID_TOKEN() {
	transID := rnd()
	roundID := rnd()
	resp, err := s.client.PromoBuyin(gameID, transID, roundID, toMoney(10.0), toMoney(10.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)
	s.client.baseRequest.Token = "InvalidTokenWhichIsLongEnoughToPassTheRestriction"
	respRef, errRef := s.client.PromoRefund(transID, gameID, roundID, toMoney(10.0))
	s.Require().Error(errRef, "Request should produce hard error")
	s.Assert().False(respRef.Success)
	s.Assert().Equal(redtiger.NotAuthorized, respRef.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_TRY_TO_REFUND_NON_EXISTING_BUYIN_TRANSACTION() {
	transID := rnd()
	roundID := rnd()

	resp, err := s.client.PromoRefund(transID, gameID, roundID, toMoney(10.0))

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.TransactionNotFound, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_MAKES_A_BUYIN_REQUEST_WITH_AN_INVALID_CURRENCY() {
	s.client.baseRequest.Currency = "EUR"
	transID := rnd()
	roundID := rnd()
	resp, err := s.client.PromoBuyin(gameID, transID, roundID, toMoney(10.0), toMoney(10.0))

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(resp.Success)
	s.Assert().Equal(redtiger.InvalidUserCurrency, resp.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_PROMO_SETTLE() {
	transID := rnd()
	roundID := rnd()

	resp, err := s.client.PromoBuyin(gameID, transID, roundID, toMoney(10.0), toMoney(10.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(resp.Success)

	res, err := s.client.PromoSettle(gameID, rnd(), roundID, toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(res.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_PROMO_SETTLE_RECON() {
	s.client.baseRequest.Token = s.ProviderConfig.Auth["recon_token"].(string)
	res, err := s.client.PromoSettle(gameID, rnd(), "", toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(res.Success)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_PROMO_SETTLE_WITH_USED_TRANSACTION_ID() {
	transID := rnd()
	res, err := s.client.PromoSettle(gameID, transID, "", toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))

	s.Require().NoError(err, "Request should not produce hard error")
	s.Assert().True(res.Success)
	// Firing twice should result in 200 OK but the balance should stay the same
	balance := res.Result.Balance.Cash

	res, err = s.client.PromoSettle(gameID, transID, "", toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))
	s.Require().NoError(err, "Request should produce not hard error")
	s.Assert().True(res.Success)
	s.Assert().Equal(balance, res.Result.Balance.Cash)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_PROMO_SETTLE_WITH_AN_INVALID_TOKEN() {
	s.client.baseRequest.Token = "InvalidToken-long-enough-for-min-requirement"
	transID := rnd()
	res, err := s.client.PromoSettle(gameID, transID, "", toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(res.Success)
	s.Assert().Equal(redtiger.NotAuthorized, res.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) Test_TEST_PROMO_SETTLE_WITH_AN_INVALID_CURRENCY() {
	transID := rnd()
	s.client.baseRequest.Currency = "EUR"
	res, err := s.client.PromoSettle(gameID, transID, "", toMoney(10.0), toMoney(0.0), toMoney(0.0), toJackpotMoney(0.0))

	s.Require().Error(err, "Request should produce hard error")
	s.Assert().False(res.Success)
	s.Assert().Equal(redtiger.InvalidUserCurrency, res.Error.Code)
}

func (s *RedTigerIntegrationTestSuite) prepareCase() (*RGIClient, *redtiger.Balance) {
	s.client = NewRGIClient(s.ValkyrieURL, s.ProviderConfig.Auth["api_key"].(string), s.BackdoorURL)
	s.Require().NoError(s.client.SetupSession(currency), "Request should not produce hard error")
	resp, err := s.client.Auth()
	s.Require().NoError(err, "Request should not produce hard error")

	return s.client, &resp.Result.Balance
}

func rnd() string {
	return utils.RandomString(10)
}

func addToMoney(money redtiger.Money, val float64) redtiger.Money {
	res := decimal.Decimal(money).Add(decimal.NewFromFloat(val))
	return redtiger.Money(res)
}

func subFromMoney(money redtiger.Money, val float64) redtiger.Money {
	res := decimal.Decimal(money).Sub(decimal.NewFromFloat(val))
	return redtiger.Money(res)
}

func toMoney(val float64) redtiger.Money {
	return redtiger.Money(testutils.NewFloatAmount(val))
}

func toJackpotMoney(val float64) redtiger.JackpotMoney {
	return redtiger.JackpotMoney(testutils.NewFloatAmount(val))
}
