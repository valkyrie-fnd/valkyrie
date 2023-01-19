// Package caleta_test contains integration tests for verifying the Caleta provider implementation.
package caleta_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/valkyrie-fnd/valkyrie-stubs/datastore"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/provider/caleta"
	"github.com/valkyrie-fnd/valkyrie/provider/caleta/auth"
	"github.com/valkyrie-fnd/valkyrie/provider/internal/test"
)

const (
	currency            = "USD"
	gameCode            = "SomeSlotGame"
	initialCashBalance  = 1000
	initialPromoBalance = 10
	multiplier          = 100000 // Caleta represent floats as integers multiplied by 100000
)

type CaletaIntegrationTestSuite struct {
	test.IntegrationTestSuite
	client *RGIClient
	signer auth.Signer
}

// Runs all tests in suite below one by one
func TestSuite(t *testing.T) {
	caletaPrivatePEM, caletaPublicPEM, _ := testutils.GenerateRsaKey()
	signer, _ := caleta.NewSigner(caletaPrivatePEM)

	providerConfigFn := func(ds datastore.ExtendedDatastore) configs.ProviderConf {
		return configs.ProviderConf{
			Name: caleta.ProviderName,
			Auth: map[string]any{
				"verification_key": string(caletaPublicPEM),
				"operator_id":      "valkyrie",
			},
		}
	}

	testSuite := &CaletaIntegrationTestSuite{
		IntegrationTestSuite: test.IntegrationTestSuite{
			ProviderConfigFn: providerConfigFn,
		},
		signer: signer,
	}
	suite.Run(t, testSuite)
}

func (s *CaletaIntegrationTestSuite) SetupTest() {
	s.client = NewRGIClient(s.ValkyrieURL, s.BackdoorURL, s.signer)

	s.Require().NoError(s.client.SetupSession(currency))
}

/*
 * C's manual tests are:
 *
 * Test Open Game - Open all games and validate icons, game ids and different languages - not applicable
 * Test Open Demo Game - Open all games and confirm the DEMO version is working fine. - not applicable
 * Test Place Bet - Place bets on different currencies and games. Confirm balance and debit values are ok and check the back office. - check
 * Test Receive Winnings - Test different winnings from different games and situations: Bonus, Free Spins, Extra Balls - check
 * Test Rollback - Forcing a bet to fail by changing the Wallet URL - the system will force a rollback. Caleta RGS only rollback bets, winnings are never rolled back. - check
 * Test Retry - Forcing a Rollback or Win to fail by changing the Wallet URL - the system will try to retry the transaction until succeeding - not applicable
 * Test Bingos: Extra Balls - Play Bingos and test Extra Ball game - the operator must support additional bets inside the same round; rollbacks on extra balls don't invalidate the round, only the failed transaction. - check
 * Acceptance Tests on Client - Exploratory tests on some selected games to confirm game flow is ok. In this phase, we also perform Postman calls, simulating bad-intended users to confirm round will be invalidated. - not applicable
 *
 */

/*
Errors to test:
	RSERRORBETLIMITEXCEEDED        Status = "RS_ERROR_BET_LIMIT_EXCEEDED" TODO: not implemented in genericpam (yet)
	RSERRORDUPLICATETRANSACTION    Status = "RS_ERROR_DUPLICATE_TRANSACTION" - check
	RSERRORINVALIDGAME             Status = "RS_ERROR_INVALID_GAME" - check
	RSERRORINVALIDSIGNATURE        Status = "RS_ERROR_INVALID_SIGNATURE" - check
	RSERRORINVALIDTOKEN            Status = "RS_ERROR_INVALID_TOKEN" - check
	RSERRORNOTENOUGHMONEY          Status = "RS_ERROR_NOT_ENOUGH_MONEY" - check
	RSERRORTIMEOUT                 Status = "RS_ERROR_TIMEOUT" TODO: a bit tricky to test
	RSERRORTOKENEXPIRED            Status = "RS_ERROR_TOKEN_EXPIRED" - check
	RSERRORTRANSACTIONDOESNOTEXIST Status = "RS_ERROR_TRANSACTION_DOES_NOT_EXIST" - check
	RSERRORTRANSACTIONROLLEDBACK   Status = "RS_ERROR_TRANSACTION_ROLLED_BACK" - check
	RSERRORUNKNOWN                 Status = "RS_ERROR_UNKNOWN" - TODO: maybe applicable once we have implementation
	RSERRORUSERDISABLED            Status = "RS_ERROR_USER_DISABLED" - check
	RSERRORWRONGCURRENCY           Status = "RS_ERROR_WRONG_CURRENCY" - check
	RSERRORWRONGSYNTAX             Status = "RS_ERROR_WRONG_SYNTAX" - TODO: doesn't make sense to me, but maybe applicable once we have implementation
	RSERRORWRONGTYPES              Status = "RS_ERROR_WRONG_TYPES" - TODO: doesn't make sense to me, but maybe applicable once we have implementation
	RSOK                           Status = "RS_OK" - check
*/

func (s *CaletaIntegrationTestSuite) Test_Check() {
	check, err := s.client.Check()
	s.Assert().NoError(err)
	if s.Assert().NotNil(check) {
		s.Assert().NotEmpty(check.Token)
	}
}

func (s *CaletaIntegrationTestSuite) Test_Check_Invalid_Session() {
	s.client.setSession("invalid-session")

	check, err := s.client.Check()
	s.Assert().Error(err)
	s.Assert().Nil(check)
}

func (s *CaletaIntegrationTestSuite) Test_Check_New_Session() {
	check, err := s.client.Check()
	s.Assert().NoError(err)
	session := check.Token

	check, err = s.client.Check()
	s.Assert().NoError(err)
	s.Assert().NotEmpty(check.Token)
	s.Assert().NotEqual(session, check.Token)
}

func (s *CaletaIntegrationTestSuite) Test_Balance() {
	balance, err := s.client.Balance(gameCode)
	s.Assert().NoError(err)
	s.assertBalance(balance, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_Balance_Invalid_Session() {

	s.client.setSession("invalid-session")

	balance, err := s.client.Balance(gameCode)
	s.Assert().NoError(err)
	s.assertBalanceError(balance, caleta.RSERRORINVALIDTOKEN)
}

func (s *CaletaIntegrationTestSuite) Test_Balance_Invalid_Signature() {
	// Generate a new private key used for signing
	privatePEM, _, _ := testutils.GenerateRsaKey()
	signer, _ := caleta.NewSigner(privatePEM)
	s.client = NewRGIClient(s.ValkyrieURL, s.BackdoorURL, signer)

	balance, err := s.client.Balance(gameCode)
	s.Assert().NoError(err)
	s.assertBalanceError(balance, caleta.RSERRORINVALIDSIGNATURE)
}

func (s *CaletaIntegrationTestSuite) Test_Bet() {
	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance-1)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_One_Cent() {
	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), 1000)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance-0.01)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_One_Milli() {
	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), 100)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance-0.001)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Ten_Micro() {
	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), 1)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance-0.00001)
}

// Test Bingos: Extra Balls - Play Bingos and test Extra Ball game - the operator must support additional bets inside the same round;
// rollbacks on extra balls don't invalidate the round, only the failed transaction.
func (s *CaletaIntegrationTestSuite) Test_Bet_Multiple() {
	round := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, uuid(), multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	bet, err = s.client.Bet(gameCode, currency, round, uuid(), multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	bet, err = s.client.Bet(gameCode, currency, round, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance-3)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Invalid_Session() {
	s.client.setSession("invalid-session")

	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORINVALIDTOKEN)
}

// RS_ERROR_TOKEN_EXPIRED - when a new token exists (this rule applies only for /wallet/bet
func (s *CaletaIntegrationTestSuite) Test_Bet_Expired_Session() {
	check, err := s.client.Check()
	s.Assert().NoError(err)
	session := *check.Token // save old session

	_, err = s.client.Check()
	s.Assert().NoError(err)

	s.client.setSession(session) // use old session

	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORTOKENEXPIRED)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Negative() {
	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), -multiplier)
	s.Assert().NoError(err)
	if s.Assert().NotNil(bet) {
		s.Assert().NotEqual(bet.Status, caleta.RSOK) // TODO: not sure about errorcode

		s.Assert().Nil(bet.User)
		s.Assert().Nil(bet.Balance)
		s.Assert().Nil(bet.Currency)
	}
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Exceeding_Balance() {
	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), initialCashBalance*multiplier+1)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORNOTENOUGHMONEY)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Blatantly_Exceeding_Balance() {
	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), math.MaxInt32) // json format int32 max value
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORNOTENOUGHMONEY)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Invalid_Game() {
	bet, err := s.client.Bet("invalid game", currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORINVALIDGAME)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Wrong_Currency() {
	bet, err := s.client.Bet(gameCode, "EUR", uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORWRONGCURRENCY)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Invalid_Currency() {
	bet, err := s.client.Bet(gameCode, "schmeckles", uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORWRONGCURRENCY)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Invalid_Signature() {
	// Generate a new private key used for signing
	privatePEM, _, _ := testutils.GenerateRsaKey()
	signer, _ := caleta.NewSigner(privatePEM)
	s.client = NewRGIClient(s.ValkyrieURL, s.BackdoorURL, signer)

	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORINVALIDSIGNATURE)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Duplicate_TransactionId() {
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, uuid(), transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	// Same user and session, duplicate transactionID
	bet, err = s.client.Bet(gameCode, currency, uuid(), transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORDUPLICATETRANSACTION)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Duplicate_TransactionId_For_Different_Users() {
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, uuid(), transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	// Set up a new user and session
	bet, err = s.client.Bet(gameCode, currency, uuid(), transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORDUPLICATETRANSACTION)
}

func (s *CaletaIntegrationTestSuite) Test_PromoBet() {
	bet, err := s.client.PromoBet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_PromoBet_Wrong_Currency() {
	bet, err := s.client.PromoBet(gameCode, "EUR", uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORWRONGCURRENCY)
}

func (s *CaletaIntegrationTestSuite) Test_PromoBet_Invalid_Currency() {
	bet, err := s.client.PromoBet(gameCode, "schmeckles", uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORWRONGCURRENCY)
}

// RS_ERROR_TOKEN_EXPIRED - when a new token exists (this rule applies only for /wallet/bet
func (s *CaletaIntegrationTestSuite) Test_PromoBet_Expired_Session() {
	check, err := s.client.Check()
	s.Assert().NoError(err)
	session := *check.Token // save old session

	_, err = s.client.Check()
	s.Assert().NoError(err)

	s.client.setSession(session) // use old session

	bet, err := s.client.PromoBet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORTOKENEXPIRED)
}

func (s *CaletaIntegrationTestSuite) Test_PromoBet_Exceeding_Balance() {
	bet, err := s.client.PromoBet(gameCode, currency, uuid(), uuid(), initialPromoBalance*multiplier+1)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORNOTENOUGHMONEY)
}

func (s *CaletaIntegrationTestSuite) Test_Win() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), 2*multiplier)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance+1)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Missing_Bet() {
	win, err := s.client.Win(gameCode, currency, uuid(), uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORTRANSACTIONDOESNOTEXIST)
}

func (s *CaletaIntegrationTestSuite) Test_Win_One_Cent() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), 1000)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance-0.99)
}

func (s *CaletaIntegrationTestSuite) Test_Win_One_Milli() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), 100)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance-0.999)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Ten_Micro() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), 1)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance-0.99999)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Negative() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), -multiplier)
	s.Assert().NoError(err)
	if s.Assert().NotNil(win) {
		s.Assert().NotEqual(win.Status, caleta.RSOK) // TODO: not sure about errorcode

		s.Assert().Nil(win.User)
		s.Assert().Nil(win.Balance)
		s.Assert().Nil(win.Currency)
	}
}

func (s *CaletaIntegrationTestSuite) Test_Win_Invalid_Game() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win("invalid game", currency, round, transactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORINVALIDGAME)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Wrong_Currency() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win(gameCode, "EUR", round, transactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORWRONGCURRENCY)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Invalid_Currency() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.Win(gameCode, "schmeckles", round, transactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORWRONGCURRENCY)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Invalid_Signature() {
	// Generate a new private key used for signing
	privatePEM, _, _ := testutils.GenerateRsaKey()
	signer, _ := caleta.NewSigner(privatePEM)
	s.client = NewRGIClient(s.ValkyrieURL, s.BackdoorURL, signer)

	win, err := s.client.Win(gameCode, currency, uuid(), uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORINVALIDSIGNATURE)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Duplicate_TransactionId() {
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, uuid(), transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	// Same user and session, duplicate transactionID
	win, err := s.client.Win(gameCode, currency, uuid(), transactionID, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORDUPLICATETRANSACTION)
}

func (s *CaletaIntegrationTestSuite) Test_Win_Duplicate_TransactionId_For_Different_Users() {
	transactionID := uuid()
	round := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	// Set up a new user and session
	s.Require().NoError(s.client.SetupSession(currency))

	win, err := s.client.Win(gameCode, currency, round, transactionID, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORDUPLICATETRANSACTION)
}

func (s *CaletaIntegrationTestSuite) Test_PromoWin() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.PromoBet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.PromoWin(gameCode, currency, round, transactionID, uuid(), 2*multiplier)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_PromoWin_Wrong_Currency() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.PromoWin(gameCode, "EUR", round, transactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORWRONGCURRENCY)
}

func (s *CaletaIntegrationTestSuite) Test_PromoWin_Invalid_Currency() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	win, err := s.client.PromoWin(gameCode, "schmeckles", round, transactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORWRONGCURRENCY)
}

func (s *CaletaIntegrationTestSuite) Test_PromoWin_Missing_Bet() {
	win, err := s.client.PromoWin(gameCode, currency, uuid(), uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORTRANSACTIONDOESNOTEXIST)
}

func (s *CaletaIntegrationTestSuite) Test_Rollback() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)
	s.Assert().Equal(*bet.Balance/multiplier, initialCashBalance-1)

	rollback, err := s.client.Rollback(gameCode, round, transactionID, uuid(), RoundClosed)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance)
}

// Test Bingos: Extra Balls - Play Bingos and test Extra Ball game - the operator must support additional bets inside the same round;
// rollbacks on extra balls don't invalidate the round, only the failed transaction.
func (s *CaletaIntegrationTestSuite) Test_Rollback_Multiple_Bet() {
	round := uuid()
	firstTransactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, firstTransactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	secondTransactionID := uuid()
	bet, err = s.client.Bet(gameCode, currency, round, secondTransactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	rollback, err := s.client.Rollback(gameCode, round, firstTransactionID, uuid(), RoundOpen)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance-1)

	win, err := s.client.Win(gameCode, currency, round, secondTransactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_Rollback_Missing_Referenced_Transaction() {
	// Trying to roll back a non-existing bet transaction.
	// Caleta prefers that Valkyrie just returns OK and balance (idempotent) in these cases.
	rollback, err := s.client.Rollback(gameCode, uuid(), uuid(), uuid(), RoundClosed)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_Rollback_Then_Win() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)
	s.Assert().Equal(*bet.Balance/multiplier, initialCashBalance-1)

	rollback, err := s.client.Rollback(gameCode, round, transactionID, uuid(), RoundOpen)
	s.Assert().NoError(err)
	s.Assert().Equal(rollback.Status, caleta.RSOK)
	s.Assert().Equal(*rollback.Balance/multiplier, initialCashBalance)

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORTRANSACTIONROLLEDBACK)
}

func (s *CaletaIntegrationTestSuite) Test_PromoRollback() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.PromoBet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)
	s.Assert().Equal(*bet.Balance/multiplier, initialCashBalance)

	rollback, err := s.client.PromoRollback(gameCode, round, transactionID, uuid(), RoundClosed)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_PromoRollback_Then_Win() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.PromoBet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance)

	rollback, err := s.client.PromoRollback(gameCode, round, transactionID, uuid(), RoundOpen)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance)

	win, err := s.client.PromoWin(gameCode, currency, round, transactionID, uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(win, caleta.RSERRORTRANSACTIONROLLEDBACK)
}

// A token should be valid for all bet transactions until a new one is generated, expired tokens should continue
// to be valid for Win/Rollback transactions if is related to a previous existing bet.
func (s *CaletaIntegrationTestSuite) Test_Win_Using_Expired_Session() {
	check, err := s.client.Check()
	s.Assert().NoError(err)
	session := *check.Token // save old session

	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance-1)

	_, err = s.client.Check()
	s.Assert().NoError(err)

	s.client.setSession(session) // use old session

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), 2*multiplier)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance+1)
}

func (s *CaletaIntegrationTestSuite) Test_PromoWin_Using_Expired_Session() {
	check, err := s.client.Check()
	s.Assert().NoError(err)
	session := *check.Token // save old session

	round := uuid()
	transactionID := uuid()
	bet, err := s.client.PromoBet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance)

	_, err = s.client.Check()
	s.Assert().NoError(err)

	s.client.setSession(session) // use old session

	win, err := s.client.PromoWin(gameCode, currency, round, transactionID, uuid(), 2*multiplier)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_PromoRollback_Using_Expired_Session() {
	check, err := s.client.Check()
	s.Assert().NoError(err)
	session := *check.Token // save old session

	round := uuid()
	transactionID := uuid()
	bet, err := s.client.PromoBet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance)

	_, err = s.client.Check()
	s.Assert().NoError(err)

	s.client.setSession(session) // use old session

	rollback, err := s.client.PromoRollback(gameCode, round, transactionID, uuid(), RoundClosed)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_Bet_Blocked() {
	s.Require().NoError(s.client.BlockAccount(currency))

	bet, err := s.client.Bet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORUSERDISABLED)
}

func (s *CaletaIntegrationTestSuite) Test_PromoBet_Blocked() {
	s.Require().NoError(s.client.BlockAccount(currency))

	bet, err := s.client.PromoBet(gameCode, currency, uuid(), uuid(), multiplier)
	s.Assert().NoError(err)
	s.assertBalanceError(bet, caleta.RSERRORUSERDISABLED)
}

// Spec is not clear, but guessing we should still allow wins and rollbacks for disabled users
func (s *CaletaIntegrationTestSuite) Test_Win_Not_Blocked() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	s.Require().NoError(s.client.BlockAccount(currency))

	win, err := s.client.Win(gameCode, currency, round, transactionID, uuid(), 2*multiplier)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance+1)
}

func (s *CaletaIntegrationTestSuite) Test_PromoWin_Not_Blocked() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.PromoBet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.Assert().Equal(bet.Status, caleta.RSOK)

	s.Require().NoError(s.client.BlockAccount(currency))

	win, err := s.client.PromoWin(gameCode, currency, round, transactionID, uuid(), 2*multiplier)
	s.Assert().NoError(err)
	s.assertBalance(win, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_Rollback_Not_Blocked() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.Bet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance-1)

	s.Require().NoError(s.client.BlockAccount(currency))

	rollback, err := s.client.Rollback(gameCode, round, transactionID, uuid(), RoundClosed)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) Test_PromoRollback_Not_Blocked() {
	round := uuid()
	transactionID := uuid()
	bet, err := s.client.PromoBet(gameCode, currency, round, transactionID, multiplier)
	s.Assert().NoError(err)
	s.assertBalance(bet, initialCashBalance)

	s.Require().NoError(s.client.BlockAccount(currency))

	rollback, err := s.client.PromoRollback(gameCode, round, transactionID, uuid(), RoundClosed)
	s.Assert().NoError(err)
	s.assertBalance(rollback, initialCashBalance)
}

func (s *CaletaIntegrationTestSuite) assertBalance(balance *caleta.BalanceResponse, expectedBalance float64) {
	if s.Assert().NotNil(balance) {
		s.Assert().Equal(caleta.RSOK, balance.Status)

		s.Assert().NotNil(balance.User)
		s.Assert().Equal(s.client.userID, *balance.User)

		s.Assert().NotNil(balance.Balance)
		s.Assert().Equal(int(expectedBalance*multiplier), *balance.Balance)

		s.Assert().NotNil(balance.Currency)
		s.Assert().Equal(currency, string(*balance.Currency))
	}
}

func (s *CaletaIntegrationTestSuite) assertBalanceError(balance *caleta.BalanceResponse, status caleta.Status) {
	if s.Assert().NotNil(balance) {
		s.Assert().Equal(status, balance.Status)

		s.Assert().Nil(balance.User)
		s.Assert().Nil(balance.Balance)
		s.Assert().Nil(balance.Currency)
	}
}
