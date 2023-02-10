// A copy of Evo's licensee integration tests.
package evolution_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valkyrie-fnd/valkyrie-stubs/datastore"

	"github.com/valkyrie-fnd/valkyrie/provider/internal/test"

	"github.com/shopspring/decimal"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/provider/evolution"
)

const (
	currency         = "EUR"
	userOne          = "5000001"
	userTwo          = "5000002"
	tableIDTestValue = "someLiveGame"
)

type EvolutionIntegrationTestSuite struct {
	test.IntegrationTestSuite
}

// Runs all tests in suite below one by one
func TestSuite(t *testing.T) {
	providerConfigFn := func(ds datastore.ExtendedDatastore) configs.ProviderConf {
		apiKey, _ := ds.GetProviderAPIKey(evolution.ProviderName)
		return configs.ProviderConf{
			Name:     evolution.ProviderName,
			BasePath: "/evolution",
			Auth: map[string]any{
				"api_key": testutils.EnvOrDefault("EVO_API_KEY", apiKey.APIKey),
			},
		}
	}
	suite.Run(t, &EvolutionIntegrationTestSuite{
		IntegrationTestSuite: test.IntegrationTestSuite{
			ProviderConfigFn: providerConfigFn,
		},
	})
}

// Verification of incorrect SID supply case: if a new SID returned after initialization,
// further requests storing old SID must be rejected.
//
//	Expected behavior: Requests with old SID should be rejected.
func (s *EvolutionIntegrationTestSuite) Test_checkUsingNewSID() {
	client := NewEvo(s.ValkyrieURL, s.BackdoorURL, s.ProviderConfig.Auth["api_key"].(string))
	// 1. Call service method SID(). Get new SID.
	sid, err := client.SID(userOne, 'P')
	s.Require().NoError(err)
	// 2. Initialize (method check() call).
	//		Verify response - response must contain status OK.
	client.checkWithAssert(sid.SID, userOne, func(cur *evolution.CheckResponse, err error) {
		s.Assert().Equal("OK", cur.Status, "Check should return OK status")
	})

	// 3. If after initialize response contains new SID, must verify whether old SID can be used.
	s.Assert().NotEqual(sid.SID, client.reqBase.SID, "Sid should be updated after check()")

	// 4. Check balance with old SID. Verify response - response? must contain failures (status INVALID_SID).
	client.reqBase.SID = sid.SID
	client.balanceWithAnyAssert(s.statusCodeAssert("INVALID_SID"))

	// 5. Place bet with old SID (method debit() with default amount).
	//		Verify response - response must contain failures (status INVALID_SID).
	gameID, transID, refID := rnd(), rnd(), rnd()
	client.placeBet(10, tableIDTestValue, gameID, transID, refID, s.statusCodeAssert("INVALID_SID"))

	// 6. End game with old SID (method credit() with default amount).
	//		Verify response - response must contain failures (status INVALID_SID or BET_DOES_NOT_EXIST).
	client.settleBet(10, tableIDTestValue, gameID, transID, s.statusCodeAssert("INVALID_SID"))
}

// Verification of main wallet methods before running specific test cases/scenarios. Main wallet methods are: get balance, place bet, settle bet.
//
//	Expected behavior: get balance, place bet and credit the bet, check balance changes.
func (s *EvolutionIntegrationTestSuite) Test_allMethodsCheckBeforeRunningTests() {
	client, init := s.prepareCase(userOne)

	// 4. Place bet (method debit() with default amount). Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	client.placeBet(1, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), init.Balance, -1))

	// 5. End game (method credit() with default amount). Verify response - response must contain status OK. Verify balance.
	client.settleBet(1, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), init.Balance, 0))

	// 6. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), init.Balance, 0)
}

// Verifies that bet can be handled out only once.
//
//	Expected behavior: Second request to place same bet should be rejected.
func (s *EvolutionIntegrationTestSuite) Test_debitBetAlreadyExist() {
	client, initialBalance := s.prepareCase(userOne)

	//  4. Place bet. Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	transID := client.placeBet(1, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), initialBalance.Balance, -1))

	//  5. Place bet with the same transID and refId (see previous step). Verify response - response
	//  must contain failures (status BET_ALREADY_EXIST or OK). Verify balance.)
	client.placeBet(1, tableIDTestValue, transID, gameRound, refID, balanceStatusAssert(s.T(), initialBalance.Balance, -1))

	//  6. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialBalance.Balance, -1, "Only one debit should affect balance")

	//  7. End game (method credit() with default amount). Verify response - response must contain status OK. Verify balance.
	client.settleBet(1, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), initialBalance.Balance, 0))

	//  8. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialBalance.Balance, 0)
}

// Verifies that bet can be canceled and balance changes appropriately.
//
//	Expected behavior: Cancel request should be proceeded correctly.
func (s *EvolutionIntegrationTestSuite) Test_debitCancel() {
	client, initialBalance := s.prepareCase(userOne)

	// 4. Place bet. Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	transID := client.placeBet(1, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), initialBalance.Balance, -1))

	// 5. Cancel bet (with the same transactionId(see previous step)). Verify response -
	//       response must contain status OK. Verify balance.
	client.cancelBet(1, tableIDTestValue, gameRound, transID, refID, balanceStatusAssert(s.T(), initialBalance.Balance, 0))

	// 6. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialBalance.Balance, 0)
}

// Verifies that bet can be placed, paid out and balance changes appropriately.
//
//	Expected behavior: Credit request should be proceeded correctly.
func (s *EvolutionIntegrationTestSuite) Test_debitCredit() {
	client, init := s.prepareCase(userOne)

	// 4. Place bet. Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	client.placeBet(2, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), init.Balance, -2))

	// 5. End game (method credit() with default amount).
	//	  Verify response - response must contain status OK. Verify balance.
	client.settleBet(2, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), init.Balance, 0))

	// 6. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), init.Balance, 0)
}

// Verifies that bet can be paid out only once. Attempts to payout for same bet
// should be rejected and contain error BET_ALREADY_SETTLED.
//
//	Expected behavior: Second request to payout for same bet should be rejected.
//
// Note - for game-wise settlement probably the game round is the most relevant property to check. I.e. one game round should only be settled once
// Note - however we use refId and check if there is already a deposit for this ID
func (s *EvolutionIntegrationTestSuite) Test_debitCreditBetAlreadySettled() {
	client, init := s.prepareCase(userOne)

	// 4. Place bet. Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	client.placeBet(2, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), init.Balance, -2))

	// 5. End game (method credit() with default amount). Verify response - response must contain
	//    status OK. Verify balance.

	client.settleBet(2, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), init.Balance, 0))

	// 6. End game (method credit() with default amount) with the same refId(see previous step).
	//    Verify response - response must contain failures (status BET_ALREADY_SETTLED or OK). Verify balance.
	//    Note - there is no point in validating the balance in the return since an error is returned.
	//    Balance is asserted in next step
	client.settleBet(2, tableIDTestValue, gameRound, refID)

	// 7. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), init.Balance, 0)
}

// Verifies if credit does not succeed for a bet with non-existing reference ID
// and BET_DOES_NOT_EXIST response is the case.
// Random - and must not be accepted/processed UNEXPECTED scenario in terms of Evolution.
//
//	Expected behavior: Request with fake reference ID should be rejected.
func (s *EvolutionIntegrationTestSuite) Test_debitCreditBetDoesNotExist() {
	client, init := s.prepareCase(userOne)

	// 4. Place bet. Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	client.placeBet(2, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), init.Balance, -2))

	// 5. End game with fake refId. Verify response - response must contain failures (status BET_DOES_NOT_EXIST). Verify balance.
	client.settleBet(2, tableIDTestValue, gameRound, rnd(), s.statusCodeAssert("BET_DOES_NOT_EXIST"),
		s.balanceAssert(init.Balance, -2, "balance after bet"))

	// 6. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), init.Balance, -2, "Balance unaffected after credit with fake refid")

	// 7. End game (method credit() call). Verify response - response must contain status OK. Verify balance.
	client.settleBet(2, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), init.Balance, 0))

	// 8. Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), init.Balance, 0)
}

// For casinos that don't use depositAfterCancel
// Expected behavior: Cancel bet request should be rejected.
func (s *EvolutionIntegrationTestSuite) Test_debitCreditCancel() {
	client, init := s.prepareCase(userOne)

	// Place Bet. Verify response - response must contain status OK. Verify Balance.
	gameRound, refID := rnd(), rnd()
	transID := client.placeBet(2, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), init.Balance, -2))

	// Settle Bet. Verify response - response must contain status OK. Verify Balance.
	client.settleBet(2, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), init.Balance, 0))

	// Cancel Bet. Verify response - response must contain status BET_ALREADY_SETTLED. Verify balance.
	client.cancelBet(2, tableIDTestValue, gameRound, transID, refID, s.statusCodeAssert("BET_ALREADY_SETTLED"),
		s.balanceAssert(init.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify Balance.
	client.balanceWithAssert(s.T(), init.Balance, 0)
}

// Verifies if separate bets placed within different sids and channels are processed successfully.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_sameUserPlayingWithDifferentChannelType() {
	// 1. Call service method sid() with channelType = 'P'. Getting new SID('sid_A').
	client := NewEvo(s.ValkyrieURL, s.BackdoorURL, s.ProviderConfig.Auth["api_key"].(string))
	sidA, err := client.SID(userOne, 'P')
	s.Assert().NoError(err)

	initialBalance, _ := client.Balance(currency)

	// 2. Initialize (method check() call) with channelType = 'P'. Verify response - response must contain status OK.
	sidA.SID = client.checkWithAssert(sidA.SID, userOne, func(cur *evolution.CheckResponse, err error) {
		s.Assert().NoError(err)
		s.Assert().Equal("OK", cur.Status)
	})

	// 3. Call service method sid() with channelType = 'M'. Getting new SID('sidB').
	sidB, err := client.SID(userOne, 'M')
	s.Assert().NoError(err)

	// 4. Initialize with channelType = 'M'. Verify response - response must contain status OK.
	sidB.SID = client.checkWithAssert(sidB.SID, userOne, func(cur *evolution.CheckResponse, err error) {
		s.Assert().NoError(err)
		s.Assert().Equal("OK", cur.Status)
	})

	// 5. Check balance for sid = 'sid_A'. Verify response - response must contain status OK.
	client.reqBase.SID = sidA.SID
	client.balanceWithAnyAssert(s.statusCodeAssert("OK", "sidA balance should work"))

	// 6. Check balance for sid = 'sid_B'. Verify response - response must contain status OK.
	client.reqBase.SID = sidB.SID
	client.balanceWithAnyAssert(s.statusCodeAssert("OK", "Sid_B balance should work"))

	// 7. Place bet for sid = 'sid_A'. Verify response - response must contain status OK. Verify balance.
	client.reqBase.SID = sidA.SID
	gameRound, refID := rnd(), rnd()
	client.placeBet(2, tableIDTestValue, rnd(), gameRound, refID, balanceStatusAssert(s.T(), initialBalance.Balance, -2))

	// 8. Place bet for sid = 'sid_B'. Verify response - response must contain status OK. Verify balance.
	client.reqBase.SID = sidB.SID
	client.placeBet(3, tableIDTestValue, rnd(), gameRound, rnd(), balanceStatusAssert(s.T(), initialBalance.Balance, -2-3))

	// 9. End game (method credit() call) for both bets with different sids.
	//		Verify response - response must contain status OK. Verify balance.
	client.settleBet(2+3, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), initialBalance.Balance, 0))

	// 10. Check balance for sid = 'sidA'. Verify response - response must contain status OK. Verify balance.
	client.reqBase.SID = sidA.SID
	client.balanceWithAssert(s.T(), initialBalance.Balance, 0, "sidA balance should match initial")

	// 11. Check balance for sid = 'sid_B'. Verify response - response must contain status OK. Verify balance.
	client.reqBase.SID = sidB.SID
	client.balanceWithAssert(s.T(), initialBalance.Balance, 0, "Sid_B balance should match initial")
}

// Verifies if multiple (2) bets placed by different users at a time are processed successfully.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_twoUsersBjGameEachPlaying_2_hands() {
	// Runs for games with Multi-transaction support when at least two users are defined.
	// Call service method sid() for userId = 'a'. Getting new SID('sid_A').
	// Initialize (method check() call) for userId = 'a'. Verify response - response must contain status OK.
	// Check balance for userId = 'a'. Verify response - response must contain status OK.
	clientA, initA := s.prepareCase(userOne)

	// Call service method sid() for userId = 'b'. Getting new SID('sid_B').
	// Initialize (method check() call) for userId = 'b'. Verify response - response must contain status OK.
	// Check balance for userId = 'b'. Verify response - response must contain status OK.
	clientB, initB := s.prepareCase(userTwo)

	// Place bet for userId = 'a'. Verify response - response must contain status OK. Verify balance.
	gameID, refIDA1 := rnd(), rnd()
	clientA.placeBet(2, tableIDTestValue, "", gameID, refIDA1, balanceStatusAssert(s.T(), initA.Balance, -2))

	// Place bet for userId = 'b'. Verify response - response must contain status OK. Verify balance.
	gameID, refIDB1 := rnd(), rnd()
	clientB.placeBet(3, tableIDTestValue, "", gameID, refIDB1, balanceStatusAssert(s.T(), initB.Balance, -3))

	// Place bet for userId = 'a'. Verify response - response must contain status OK. Verify balance.
	refIDA2 := rnd()
	clientA.placeBet(2, tableIDTestValue, "", gameID, refIDA2, balanceStatusAssert(s.T(), initA.Balance, -2-2))

	// Place bet for userId = 'b'. Verify response - response must contain status OK. Verify balance.
	refIDB2 := rnd()
	clientB.placeBet(3, tableIDTestValue, "", gameID, refIDB2, balanceStatusAssert(s.T(), initB.Balance, -3-3))

	// End game (method credit() call) for both bets (userId = 'a').
	//		Verify response - response must contain status OK. Verify balance.
	clientA.settleBet(2, tableIDTestValue, gameID, refIDA1, balanceStatusAssert(s.T(), initA.Balance, -2))
	clientA.settleBet(2, tableIDTestValue, gameID, refIDA2, balanceStatusAssert(s.T(), initA.Balance, 0))

	// End game (method credit() call) for both bets (userId = 'b').
	//		Verify response - response must contain status OK. Verify balance.
	clientB.settleBet(3, tableIDTestValue, gameID, refIDB1, balanceStatusAssert(s.T(), initB.Balance, -3))
	clientB.settleBet(3, tableIDTestValue, gameID, refIDB2, balanceStatusAssert(s.T(), initB.Balance, 0))

	// Check balance for userId = 'a'. Verify response - response must contain status OK. Verify balance.
	clientA.balanceWithAssert(s.T(), initA.Balance, 0)

	// Check balance for userId = 'b'. Verify response - response must contain status OK. Verify balance.
	clientB.balanceWithAssert(s.T(), initB.Balance, 0)
}

// Verifies if user A can place a bet before user's A previous bet had been settled.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_verifyCanPlaceBetBeforeSettleForAnotherGame() {
	// Call service method sid(). Get new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, init := s.prepareCase(userOne)

	// Place bet for game id = 'a'. Verify response - response must contain status OK. Verify balance.
	roundA, refIDA := rnd(), rnd()
	client.placeBet(2, tableIDTestValue, "", roundA, refIDA, balanceStatusAssert(s.T(), init.Balance, -2))

	// Place bet for another game id = 'b'. Verify response - response must contain status OK. Verify balance.
	roundB, refIDB := rnd(), rnd()
	client.placeBet(2, tableIDTestValue, "", roundB, refIDB, balanceStatusAssert(s.T(), init.Balance, -2-2))

	// End game (method credit() call) for game id = 'b'.
	//		Verify response - response must contain status OK. Verify balance.
	client.settleBet(2, tableIDTestValue, roundB, refIDB, balanceStatusAssert(s.T(), init.Balance, -2))

	// End game (method credit() call) for game id = 'a'.
	//		Verify response - response must contain status OK. Verify balance.
	client.settleBet(2, tableIDTestValue, roundA, refIDA, balanceStatusAssert(s.T(), init.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), init.Balance, 0)
}

// Verifies that non-existing userID can't be used to perform requests.
//
//	Expected behavior: All requests with non-existing user ID should be rejected.
func (s *EvolutionIntegrationTestSuite) Test_checkUsingInvalidUserId() {

	validUser, fakeUser := userOne, "fakeUserId_"+testutils.RandomString(2)
	// Call service method sid(). Getting new SID.
	client, initialBalance := s.prepareCase(validUser)

	// Put in requests userId = 'fakeUserId'. Initialize (method check() call).
	//		Verify response - response must contain status INVALID_SID.
	_ = client.checkWithAssert(client.reqBase.SID, fakeUser, func(cur *evolution.CheckResponse, err error) {
		s.Require().NoError(err)
		s.Assert().Equal("INVALID_SID", cur.Status, "Check on invalid user id should fail")
	})

	// Put in requests correct userId value.  Initialize (method check() call).
	//		Verify response - response must contain status OK.
	_ = client.checkWithAssert(client.reqBase.SID, validUser, func(cur *evolution.CheckResponse, err error) {
		s.Require().NoError(err, "Check request should work")
		s.Assert().Equal("OK", cur.Status)
	})

	// Put in requests correct userId value. Check balance.
	//		Verify response - response must contain status OK.
	client.reqBase.UserID = validUser
	client.balanceWithAnyAssert(s.statusCodeAssert("OK", "Valid user gets balance"))

	// Put in requests userId = 'fakeUserId'. Place bet.
	//		Verify response - response must contain status INVALID_SID. Verify balance.
	client.reqBase.UserID = fakeUser
	gameRoundID, refID := rnd(), rnd()
	client.placeBet(10, tableIDTestValue, "", gameRoundID, refID, s.statusCodeAssert("INVALID_SID"))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.reqBase.UserID = validUser
	client.balanceWithAssert(s.T(), initialBalance.Balance, 0, "Initial balance should still be unaffected")

	// Put in requests correct userId value. Place bet.
	//		 Verify response - response must contain status OK. Verify balance.
	client.reqBase.UserID = validUser
	transID1 := client.placeBet(10, tableIDTestValue, "", gameRoundID, refID, balanceStatusAssert(s.T(), initialBalance.Balance, -10))

	// Put in requests userId = 'fakeUserId'. Cancel Bet.
	//		Verify response - response must contain status INVALID_SID. Verify balance.
	client.reqBase.UserID = fakeUser
	client.cancelBet(10, tableIDTestValue, gameRoundID, transID1, refID, s.statusCodeAssert("INVALID_SID"))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.reqBase.UserID = validUser
	client.balanceWithAssert(s.T(), initialBalance.Balance, -10, "Fake user transactions should not affect balance")

	// Put in requests userId = 'fakeUserId'. Credit Bet.
	//		Verify response - response must contain status INVALID_SID. Verify balance.
	client.reqBase.UserID = fakeUser
	client.settleBet(2, tableIDTestValue, gameRoundID, refID, s.statusCodeAssert("INVALID_SID"))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.reqBase.UserID = validUser
	client.balanceWithAssert(s.T(), initialBalance.Balance, -10, "Fake user transactions should not affect balance")

	// Put in requests correct userId value. Credit Bet.
	//		Verify response - response must contain status OK. Verify balance.
	client.reqBase.UserID = validUser
	client.settleBet(10, tableIDTestValue, gameRoundID, refID, balanceStatusAssert(s.T(), initialBalance.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialBalance.Balance, 0, "Back to original balance")
}

// Verifies if cancel does not succeed for a bet with non-existing reference ID and BET_DOES_NOT_EXIST
//
//	response is the case
//	UNEXPECTED scenario in terms of Evolution
//		Expected behavior: Request with non-existing reference ID should be rejected
func (s *EvolutionIntegrationTestSuite) Test_debitCancelBetDoesNotExist() {
	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place bet. Verify response - response must contain status OK.
	//		Verify balance.
	gameRound, refID := rnd(), rnd()
	transID := client.placeBet(2, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), initialResp.Balance, -2))

	// Cancel bet with fake transactionId and refId.
	//		Verify response - response must contain failures (status BET_DOES_NOT_EXIST). Verify balance.
	client.cancelBet(2, tableIDTestValue, gameRound, rnd(), rnd(), s.statusCodeAssert("BET_DOES_NOT_EXIST"),
		s.balanceAssert(initialResp.Balance, -2))

	// Check balance. Verify response - response must contain status OK.
	//		Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, -2, "Reject cancel should not affect balance")

	// Cancel bet. Verify response - response must contain status OK. Verify balance. Check balance.
	//		Verify response - response must contain status OK.
	client.cancelBet(2, tableIDTestValue, gameRound, transID, refID, balanceStatusAssert(s.T(), initialResp.Balance, 0))

	// Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "Back to original balance")
}

// Verifies that bet can be canceled only once.
//
//	Expected behavior: Second request for bet cancel should be rejected.
func (s *EvolutionIntegrationTestSuite) Test_debitCancelBetAlreadySettled() {
	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place bet. Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	transID := client.placeBet(3, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), initialResp.Balance, -3))

	// Cancel bet.
	//		Verify response - response must contain status OK.
	//		Verify balance.
	client.cancelBet(3, tableIDTestValue, gameRound, transID, refID, balanceStatusAssert(s.T(), initialResp.Balance, 0))

	// Cancel bet with the same transactionId and refId(see previous step).
	//		Verify response - response must contain failures (status BET_ALREADY_SETTLED or OK). Verify balance.
	client.cancelBet(3, tableIDTestValue, gameRound, transID, refID,
		s.statusCodeAssert("BET_ALREADY_SETTLED"), s.balanceAssert(initialResp.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "Double cancel should only have one effect")
}

// Verification of 0 amount payout processing.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_creditWithZeroAmount() {
	// 	Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place bet. Verify response - response must contain status OK. Verify balance.
	gameRound, refID := rnd(), rnd()
	client.placeBet(3, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), initialResp.Balance, -3))

	// End game (method credit() call with amount 0).
	//		Verify response - response must contain status OK. Verify balance.
	client.settleBet(0, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), initialResp.Balance, -3))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, -3, "Zero credit(end game round) results in loss")
}

// Verifies if separate bets placed within different SIDs are processed successfully.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_debitCreditWithDifferentSids() {
	// Runs for games with Multi-transaction support and Mixed settlement type.
	// Call service method sid(). Get first SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place bet for first SID. Verify response - response must contain status OK. Verify balance.
	gameID, ref1 := rnd(), rnd()
	client.placeBet(3, tableIDTestValue, "", gameID, ref1, balanceStatusAssert(s.T(), initialResp.Balance, -3))

	// Call service method sid(). Get second SID. Verify response - first and second SID must be different.
	// Initialize. Verify response - response must contain status OK.
	_, _ = client.SID(userOne, 'P')

	// Place bet for another second SID. Verify response - response must contain status OK. Verify balance.
	ref2 := rnd()
	client.placeBet(4, tableIDTestValue, "", gameID, ref2, balanceStatusAssert(s.T(), initialResp.Balance, -3-4))

	// End game (method credit() call) for both bets with different SIDs.
	//		Verify response - response must contain status OK. Verify balance.
	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.settleBet(3, tableIDTestValue, gameID, ref1, balanceStatusAssert(s.T(), initialResp.Balance, -4))
	client.settleBet(4, tableIDTestValue, gameID, ref2, balanceStatusAssert(s.T(), initialResp.Balance, 0))
}

// For casinos that don't use depositAfterCancel.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_multiStepGameTwoDebitsCreditCancel() {
	// Runs for games with Multi-transaction support.
	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place initial bet. Verify response - response must contain status OK. Verify balance.
	gameID, ref1 := rnd(), rnd()
	client.placeBet(7, tableIDTestValue, "", gameID, ref1, balanceStatusAssert(s.T(), initialResp.Balance, -7))

	// Place next bet. Verify response - response must contain status OK. Verify balance.
	ref2 := rnd()
	trans2 := client.placeBet(5, tableIDTestValue, "", gameID, ref2, balanceStatusAssert(s.T(), initialResp.Balance, -7-5))

	// Settle initial bet. Verify response - response must contain status OK. Verify balance.
	client.settleBet(7, tableIDTestValue, gameID, ref1, balanceStatusAssert(s.T(), initialResp.Balance, -5))

	// Cancel next bet. Verify response - response must contain status OK. Verify balance.
	client.cancelBet(5, tableIDTestValue, gameID, trans2, ref2, balanceStatusAssert(s.T(), initialResp.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0)
}

// Verifies that different bets can be canceled and settled sequentially
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_multiStepGameTwoDebitsCancelCredit() {
	// Runs for games with Multi-transaction support.
	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place initial bet. Verify response - response must contain status OK. Verify balance.
	gameID, refID1 := rnd(), rnd()
	client.placeBet(3, tableIDTestValue, "", gameID, refID1, balanceStatusAssert(s.T(), initialResp.Balance, -3))

	// Place next bet. Verify response - response must contain status OK. Verify balance.
	refID2 := rnd()
	trans2 := client.placeBet(2, tableIDTestValue, "", gameID, refID2, balanceStatusAssert(s.T(), initialResp.Balance, -3-2))

	// Cancel next bet. Verify response - response must contain status OK. Verify balance.
	client.cancelBet(2, tableIDTestValue, gameID, trans2, refID2, balanceStatusAssert(s.T(), initialResp.Balance, -3))

	// Settle initial bet. Verify response - response must contain status OK. Verify balance.
	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.settleBet(3, tableIDTestValue, gameID, refID1, balanceStatusAssert(s.T(), initialResp.Balance, 0))
}

// Bet amount is bigger than balance. If balance returned in response with error code insufficient_funds,
// it must be same as was for 1st balance call.
//
//	Expected behavior: Balance value should be the same in both check balance requests.
func (s *EvolutionIntegrationTestSuite) Test_insufficientFunds() {
	// Initialize (method check() call). Verify response - response must contain status OK
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place initial bet. Verify response - response must contain status INSUFFICIENT_FUNDS. Verify balance.
	client.placeBet(multiply(initialResp.Balance, 3), tableIDTestValue, "", rnd(), rnd(),
		s.statusCodeAssert("INSUFFICIENT_FUNDS"), s.balanceAssert(initialResp.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "Balance unaffected")
}

// Verification of user making from 1 to 101(by default 101) top-up bets in round. (Transactions
// quantity configuration occurs in the config).
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_multipleDontSettlementsScenario() {
	// Call service method sid(). Get new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place bet. Verify response - response must contain status OK. Verify balance.
	gameID := rnd()
	for i := 0; i < 101; i++ {
		refID := rnd()
		client.placeBet(12, tableIDTestValue, rnd(), gameID, refID, balanceStatusAssert(s.T(), initialResp.Balance, -12))
		client.settleBet(12, tableIDTestValue, gameID, refID, balanceStatusAssert(s.T(), initialResp.Balance, 0))
	}

	// End game (method credit() call). Verify response - response must contain status OK. Verify balance.
	// Check balance. Verify response - response must contain status OK. Verify balance
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "end of DOND should match initial balance")
}

// Executed for casinos that don't use depositAfterCancel
//
//	Expected behavior: Second cancel bet request should be canceled.
//
// Note: In game-wise settlement cancels can be treated after game round closes, hence the last balance check breaks
func (s *EvolutionIntegrationTestSuite) Test_onePlayerPlaysTwoGameRoundsOnOneTableInParallel() {
	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place bet for game id = 'a'. Verify response - response must contain status OK. Verify balance.
	gameRoundA, refIDA1 := rnd(), rnd()
	client.placeBet(10, tableIDTestValue, "", gameRoundA, refIDA1, balanceStatusAssert(s.T(), initialResp.Balance, -10))

	// Place bet for game id = 'b'. Verify response - response must contain status OK. Verify balance.
	gameRoundB, refIDB1 := rnd(), rnd()
	transIDB1 := client.placeBet(3, tableIDTestValue, "", gameRoundB, refIDB1, balanceStatusAssert(s.T(), initialResp.Balance, -10-3))

	// Place bet for game id = 'b'. Verify response - response must contain status OK. Verify balance.
	refIDB2 := rnd()
	transIDB2 := client.placeBet(3, tableIDTestValue, "", gameRoundB, refIDB2, balanceStatusAssert(s.T(), initialResp.Balance, -10-3-3))

	// Place bet for game id = 'a'. Verify response - response must contain status OK. Verify balance.
	refIDA2 := rnd()
	client.placeBet(5, tableIDTestValue, "", gameRoundA, refIDA2, balanceStatusAssert(s.T(), initialResp.Balance, -10-3-3-5))

	// Cancel 2nd bet for game id = 'b'. Verify response - response must contain status OK. Verify balance.
	client.cancelBet(3, tableIDTestValue, gameRoundB, transIDB2, refIDB2, balanceStatusAssert(s.T(), initialResp.Balance, -10-3-5))

	// End game (method credit() call) for game id = 'a'. Verify response - response must contain status OK. Verify balance.
	client.settleBet(15, tableIDTestValue, gameRoundA, refIDA2, balanceStatusAssert(s.T(), initialResp.Balance, -3))

	// End game (method credit() call) for game id = 'b'. Verify response - response must contain status OK. Verify balance.
	client.settleBet(3, tableIDTestValue, gameRoundB, refIDB1, balanceStatusAssert(s.T(), initialResp.Balance, 0))

	// Cancel 1st bet for game id = 'b'. Verify response - response must contain failures (status BET_ALREADY_SETTLED). Verify balance.
	client.cancelBet(5, tableIDTestValue, gameRoundB, transIDB1, refIDB1,
		s.statusCodeAssert("BET_ALREADY_SETTLED"), s.balanceAssert(initialResp.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "Balance back to original")
}

// Verifies that bet cant be settled after cancellation.
//
//	Expected behavior: Bet payout request should be rejected.
func (s *EvolutionIntegrationTestSuite) Test_debitCancelCredit() {
	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place Bet. Verify response - response must contain status OK. Verify Balance.
	gameRound, refID := rnd(), rnd()
	transID := client.placeBet(10, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), initialResp.Balance, -10))

	// Cancel Bet. Verify response - response must contain status OK. Verify Balance.
	client.cancelBet(10, tableIDTestValue, gameRound, transID, refID, balanceStatusAssert(s.T(), initialResp.Balance, 0))

	// Settle Bet. Verify response - response must contain status BET_ALREADY_SETTLED. Verify balance.

	client.settleBet(300, tableIDTestValue, gameRound, refID, s.statusCodeAssert("BET_ALREADY_SETTLED"),
		s.balanceAssert(initialResp.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify Balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "Balance back to original")
}

// Verifies that big payouts can be processed.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_debitCreditWithBigPayout() {

	bigAmount := 63727092.34
	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place Bet. Verify response - response must contain status OK. Verify Balance.
	gameRound, refID := rnd(), rnd()
	client.placeBet(10, tableIDTestValue, "", gameRound, refID, balanceStatusAssert(s.T(), initialResp.Balance, -10))

	// End game with big amount 63727092.340000. Verify response - response must contain status OK. Verify Balance.
	client.settleBet(bigAmount, tableIDTestValue, gameRound, refID, balanceStatusAssert(s.T(), initialResp.Balance, -10+bigAmount))

	// Place Bet with big amount 63727092.340000. Verify response - response must contain status OK. Verify Balance.
	gameRound2, refID2 := rnd(), rnd()
	client.placeBet(bigAmount, tableIDTestValue, "", gameRound2, refID2, balanceStatusAssert(s.T(), initialResp.Balance, -10))

	// End game (method credit() call). Verify response - response must contain status OK. Verify Balance.
	client.settleBet(10, tableIDTestValue, gameRound2, refID2, balanceStatusAssert(s.T(), initialResp.Balance, 0))

	// Check balance. Verify response - response must contain status OK. Verify Balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "Balance back to original")
}

// Verifies that double promo payouts can be processed.
//
//	Expected behavior: First payout should affect balance, the second should not.
func (s *EvolutionIntegrationTestSuite) Test_doublePromoPayouts() {

	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Payout promo. Verify response - response must contain status OK. Verify Balance.
	gameRound, transID := rnd(), rnd()

	client.promoPayout(10, tableIDTestValue, transID, gameRound, balanceStatusAssert(s.T(), initialResp.Balance, 10))

	// Payout the same promo again. Verify response - response must contain status OK. Verify Balance.
	client.promoPayout(10, tableIDTestValue, transID, gameRound, balanceStatusAssert(s.T(), initialResp.Balance, 10))

	// Check balance. Verify response - response must contain status OK. Verify Balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 10, "Balance increased by first payout value")
}

// Verifies that promo payouts can be handled in general. Note that in this test context there is nothing that differs
// various flavours promo payouts from the general one
// This test covers JACKPOT_PROMO_PAYOUT, NEW_TYPE_PROMO_PAYOUT, REAL_TIME_MONETARY_REWARD_PROMO_PAYOUT,
// REWARD_GAME_PLAYABLE_SPENT_PROMO_PAYOUT, FROM_GAME_PROMO_PAYOUT, FREE_ROUND_PLAYABLE_SPENT_PROMO_PAYOUT,
// REWARD_GAME_MIN_BET_LIMIT_REACHED_PROMO_PAYOUT and REWARD_GAME_WIN_CAP_REACHED_PROMO_PAYOUT
//
//	Expected behavior: Payout should affect balance.
func (s *EvolutionIntegrationTestSuite) Test_promoPayout() {

	// Call service method sid(). Getting new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Payout promo. Verify response - response must contain status OK. Verify Balance.
	gameRound, transID := rnd(), rnd()

	client.promoPayout(10, tableIDTestValue, transID, gameRound, balanceStatusAssert(s.T(), initialResp.Balance, 10))

	// Check balance. Verify response - response must contain status OK. Verify Balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 10, "Balance increased by first payout value")
}

// Verifies the bet handling process in mixed type with multiple bets.
//
//	Expected behavior: All requests should be proceeded successfully.
func (s *EvolutionIntegrationTestSuite) Test_multipleDebitCreditCancelSettleTypeMixed() {
	// Call service method sid(). Get new SID.
	// Initialize (method check() call). Verify response - response must contain status OK.
	// Check balance. Verify response - response must contain status OK.
	client, initialResp := s.prepareCase(userOne)

	// Place Bet 15 times. Verify response - response must contain status OK. Verify balance.
	gameID, refID := rnd(), rnd()
	transactions := []string{}
	for i := 1; i < 16; i++ {
		t := client.placeBet(3, tableIDTestValue, "", gameID, fmt.Sprintf("%s_%v", refID, i),
			balanceStatusAssert(s.T(), initialResp.Balance, -float64(i*3), "balance after placed bets"))
		transactions = append(transactions, t)
	}

	// End game (method credit() call) for 1-7 bets with x2 debit amount.
	// 		Verify response - response must contain status OK. Verify balance.
	for i := 1; i < 8; i++ {
		client.settleBet(4, tableIDTestValue, gameID, fmt.Sprintf("%s_%v", refID, i),
			balanceStatusAssert(s.T(), initialResp.Balance, -45+float64(i*4), "balance after settled bet(s) 1"))
	}

	// End game (method credit() call) for 8-11 bets with debit amount.
	// 		Verify response must contain status OK. Verify balance.
	for i := 1; i < 4; i++ {
		client.settleBet(0, tableIDTestValue, gameID, fmt.Sprintf("%s_%v", refID, i+7),
			balanceStatusAssert(s.T(), initialResp.Balance, -45+28, "balance after settled bet(s) 2"))
	}

	client.settleBet(5, tableIDTestValue, gameID, fmt.Sprintf("%s_%v", refID, 11),
		balanceStatusAssert(s.T(), initialResp.Balance, -45+28+5, "balance after settled bet(s) 3"))

	// Cancel game(12-15 bets). Verify response - response must contain status OK. Verify balance.
	for i := 1; i < 5; i++ {
		client.cancelBet(3, tableIDTestValue, gameID, transactions[i+10], fmt.Sprintf("%s_%v", refID, i+11),
			balanceStatusAssert(s.T(), initialResp.Balance, -45+28+5+float64(i*3), "balance after cancelled bet(s)"))
	}

	// Check balance. Verify response - response must contain status OK. Verify balance.
	client.balanceWithAssert(s.T(), initialResp.Balance, 0, "back to original balance")
}

// Prepare case by getting SID and initial balance. Same for all test cases
//  1. Call service method sid(). Get new SID.
//  2. Initialize (method check() call). Verify response - response must contain status OK.
//  3. Check balance. Verify response - response must contain status OK.
func (s *EvolutionIntegrationTestSuite) prepareCase(userID string) (*EvoRGIClient, *evolution.StandardResponse) {
	client := NewEvo(s.ValkyrieURL, s.BackdoorURL, s.ProviderConfig.Auth["api_key"].(string))
	sidResponse, err := client.SID(userID, 'P')
	s.Require().NoError(err, "Preparing case by getting SID failed")

	r := evolution.CheckRequest{
		RequestBase: evolution.RequestBase{
			SID:    sidResponse.SID,
			UserID: userID,
		},
	}
	checkResp, err := client.Check(r)
	s.Require().NoError(err, "Check request should work")
	s.Assert().Equal("OK", checkResp.Status)

	balanceResp, err := client.Balance(currency)
	s.Require().NoErrorf(err, "Balance request should work: %v", balanceResp)
	s.Assert().Equal("OK", balanceResp.Status)

	// Here we pay for all the layers of type wrapping
	s.Require().True(decimal.Decimal(pam.Amt(balanceResp.Balance)).GreaterThanOrEqual(decimal.NewFromFloat(50)),
		"Starting balance needs to be >= 50")

	return client, balanceResp
}

// Type which happens to match the signature of Evo debit, credit and cancel methods
type transFunc func(curr string, game evolution.Game, trans evolution.Transaction) (*evolution.StandardResponse, error)

// Generic method that runs debit, credit or cancel, and verifies response and with arbitrary
//
//	method
func sendTrans(amt float64, tableID, roundID, transID, refID string, fn transFunc) (*evolution.StandardResponse, error) {
	game := evolution.Game{
		ID:   roundID,
		Type: "blackjack",
		Details: evolution.GameDetails{
			Table: evolution.GameTable{
				ID: tableID,
			},
		},
	}
	trans := evolution.Transaction{
		ID:     transID,
		RefID:  refID,
		Amount: fromFloat(amt),
	}
	return fn(currency, game, trans) // <- Actual invocation of the transaction method
}

type resultCheck func(*evolution.StandardResponse, error)

// Syntactic sugar for fetching and asserting balance
func (c *EvoRGIClient) balanceWithAssert(t *testing.T, expectedAmount evolution.Amount, addition float64, msg ...string) {
	c.balanceWithAnyAssert(balanceStatusAssert(t, expectedAmount, addition, msg...))
}

// Desperate attempt to reduce repetition in different balance call variants
func (c *EvoRGIClient) balanceWithAnyAssert(checks ...resultCheck) {
	balanceResp, err := c.Balance(currency)
	for _, rc := range checks {
		rc(balanceResp, err)
	}
}

func balanceStatusAssert(t *testing.T, expectedAmount evolution.Amount, addition float64, msg ...string) resultCheck {
	return func(sr *evolution.StandardResponse, err error) {
		expectedAmount = addToAmount(expectedAmount, addition)
		require.NoError(t, err, "Request should not produce hard error")
		assert.Equal(t, "OK", sr.Status)
		require.True(t, expectedAmount.Equal(sr.Balance), msg)
	}
}

func (s *EvolutionIntegrationTestSuite) balanceAssert(expectedAmount evolution.Amount, addition float64, msg ...string) resultCheck {
	return func(sr *evolution.StandardResponse, err error) {
		expectedAmount = addToAmount(expectedAmount, addition)
		s.Assert().True(expectedAmount.Equal(sr.Balance), msg)
	}
}

func (s *EvolutionIntegrationTestSuite) statusCodeAssert(expectedStatus string, msg ...string) resultCheck {
	return func(sr *evolution.StandardResponse, err error) {
		s.Require().NoError(err, "Request should not produce hard error")
		s.Assert().Equal(expectedStatus, sr.Status, msg)
	}
}

// Sugar for running a less verbose `check()` call
func (c *EvoRGIClient) checkWithAssert(sid, userID string, asserts func(*evolution.CheckResponse, error)) string {
	r := evolution.CheckRequest{
		RequestBase: evolution.RequestBase{
			SID:    sid,
			UserID: userID,
		},
	}
	checkResp, err := c.Check(r)
	asserts(checkResp, err)
	return checkResp.SID
}

func rnd() string {
	return testutils.RandomString(10)
}

// addToAmount unpacks amounts to decimals and does simple addition
func addToAmount(amt evolution.Amount, val float64) evolution.Amount {
	res := decimal.Decimal(pam.Amt(amt)).Add(decimal.NewFromFloat(val))
	return evolution.Amount(pam.Amt(res))
}

func multiply(amt evolution.Amount, times int64) float64 {
	res := decimal.Decimal(pam.Amt(amt)).Mul(decimal.NewFromInt(times))
	return res.InexactFloat64()
}

func fromFloat(val float64) evolution.Amount {
	return evolution.Amount(decimal.NewFromFloat(val))
}
