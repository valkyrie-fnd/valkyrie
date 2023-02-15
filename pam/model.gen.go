// Package pam provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/four-fingers/oapi-codegen version v0.0.0-20221219135408-9237c9743c67 DO NOT EDIT.
package pam

import (
	"time"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Defines values for ErrorCode.
const (
	PAMERRACCNOTFOUND           ErrorCode = "PAM_ERR_ACC_NOT_FOUND"
	PAMERRAPITOKEN              ErrorCode = "PAM_ERR_API_TOKEN"
	PAMERRBETNOTALLOWED         ErrorCode = "PAM_ERR_BET_NOT_ALLOWED"
	PAMERRBONUSOVERDRAFT        ErrorCode = "PAM_ERR_BONUS_OVERDRAFT"
	PAMERRCANCELNONWITHDRAW     ErrorCode = "PAM_ERR_CANCEL_NON_WITHDRAW"
	PAMERRCANCELNOTFOUND        ErrorCode = "PAM_ERR_CANCEL_NOT_FOUND"
	PAMERRCASHOVERDRAFT         ErrorCode = "PAM_ERR_CASH_OVERDRAFT"
	PAMERRDUPLICATETRANS        ErrorCode = "PAM_ERR_DUPLICATE_TRANS"
	PAMERRGAMENOTFOUND          ErrorCode = "PAM_ERR_GAME_NOT_FOUND"
	PAMERRMISSINGPROVIDER       ErrorCode = "PAM_ERR_MISSING_PROVIDER"
	PAMERRNEGATIVESTAKE         ErrorCode = "PAM_ERR_NEGATIVE_STAKE"
	PAMERRPLAYERNOTFOUND        ErrorCode = "PAM_ERR_PLAYER_NOT_FOUND"
	PAMERRPROMOOVERDRAFT        ErrorCode = "PAM_ERR_PROMO_OVERDRAFT"
	PAMERRROUNDNOTFOUND         ErrorCode = "PAM_ERR_ROUND_NOT_FOUND"
	PAMERRSESSIONEXPIRED        ErrorCode = "PAM_ERR_SESSION_EXPIRED"
	PAMERRSESSIONNOTFOUND       ErrorCode = "PAM_ERR_SESSION_NOT_FOUND"
	PAMERRTIMEOUT               ErrorCode = "PAM_ERR_TIMEOUT"
	PAMERRTRANSALREADYCANCELLED ErrorCode = "PAM_ERR_TRANS_ALREADY_CANCELLED"
	PAMERRTRANSALREADYSETTLED   ErrorCode = "PAM_ERR_TRANS_ALREADY_SETTLED"
	PAMERRTRANSCURRENCY         ErrorCode = "PAM_ERR_TRANS_CURRENCY"
	PAMERRTRANSNOTFOUND         ErrorCode = "PAM_ERR_TRANS_NOT_FOUND"
	PAMERRUNDEFINED             ErrorCode = "PAM_ERR_UNDEFINED"
)

// Defines values for PromoType.
const (
	FREEROUNDS       PromoType = "FREEROUNDS"
	FREESPINS        PromoType = "FREESPINS"
	PROMOBONUS       PromoType = "PROMOBONUS"
	PROMOCAP         PromoType = "PROMOCAP"
	PROMOFROMGAME    PromoType = "PROMOFROMGAME"
	PROMOLIMIT       PromoType = "PROMOLIMIT"
	PROMOMONEYREWARD PromoType = "PROMOMONEYREWARD"
	PROMOTOURNAMENT  PromoType = "PROMOTOURNAMENT"
)

// Defines values for StatusCode.
const (
	ERROR StatusCode = "ERROR"
	OK    StatusCode = "OK"
)

// Defines values for TransactionType.
const (
	CANCEL        TransactionType = "CANCEL"
	DEPOSIT       TransactionType = "DEPOSIT"
	PROMOCANCEL   TransactionType = "PROMOCANCEL"
	PROMODEPOSIT  TransactionType = "PROMODEPOSIT"
	PROMOWITHDRAW TransactionType = "PROMOWITHDRAW"
	WITHDRAW      TransactionType = "WITHDRAW"
)

// AddTransactionResponse defines model for AddTransactionResponse.
type AddTransactionResponse struct {
	// Error Error details describing why PAM rejected the request
	Error             *PamError          `json:"error,omitempty"`
	Status            StatusCode         `json:"status"`
	TransactionResult *TransactionResult `json:"transactionResult,omitempty"`
}

// Amount Amount in some currency, rounded to 6 decimal places
type Amount Amt

// Balance player account balance
type Balance struct {
	// BonusAmount Amount in some currency, rounded to 6 decimal places
	BonusAmount Amount `json:"bonusAmount"`

	// CashAmount Amount in some currency, rounded to 6 decimal places
	CashAmount Amount `json:"cashAmount"`

	// PromoAmount Amount in some currency, rounded to 6 decimal places
	PromoAmount Amount `json:"promoAmount"`
}

// BalanceResponse defines model for BalanceResponse.
type BalanceResponse struct {
	// Balance player account balance
	Balance *Balance `json:"balance,omitempty"`

	// Error Error details describing why PAM rejected the request
	Error  *PamError  `json:"error,omitempty"`
	Status StatusCode `json:"status"`
}

// BaseResponse defines model for BaseResponse.
type BaseResponse struct {
	// Error Error details describing why PAM rejected the request
	Error  *PamError  `json:"error,omitempty"`
	Status StatusCode `json:"status"`
}

// BetCode metadata about what kind of bet/transaction it is
type BetCode = string

// BucketReference Jackpot bucket reference, arbitrary use
type BucketReference = string

// BucketType Type of jackpot bucket, if any. Arbitrary use
type BucketType = string

// Country ISO 3166-1 alpha-2 two letter country code
type Country = string

// Currency ISO 4217 three letter currency code
type Currency = string

// ErrorCode - `PAM_ERR_UNDEFINED` - When you need a generic error.
// - `PAM_ERR_ACC_NOT_FOUND` - When account of `playerId` is not found.
// - `PAM_ERR_GAME_NOT_FOUND` - When specified `providerGameId` is not found.
// - `PAM_ERR_ROUND_NOT_FOUND` - In getGameRound, when there is no game round with id `providerGameRoundId`.
// - `PAM_ERR_TRANS_NOT_FOUND` - In DEPOSIT transaction if the game round with id `providerRoundId` is not found.
// - `PAM_ERR_CASH_OVERDRAFT` - When user does not have enough funds on their account for a withdraw transactions.
// - `PAM_ERR_BONUS_OVERDRAFT` - When user does not have enough funds on their bonus account for a withdraw transaction.
// - `PAM_ERR_SESSION_NOT_FOUND` - When no session is found for provided `X-Player-Token`.
// - `PAM_ERR_SESSION_EXPIRED` - When session related to `X-Player-Token` has expired.
// - `PAM_ERR_MISSING_PROVIDER` - When specified query parameter `provider` is not found.
// - `PAM_ERR_TRANS_CURRENCY` - When specified `Currency` does not match that of the session.
// - `PAM_ERR_NEGATIVE_STAKE` - When transaction amount is negative.
// - `PAM_ERR_CANCEL_NOT_FOUND` - When the transaction trying to cancel doesn't exist.
// - `PAM_ERR_TRANS_ALREADY_CANCELLED` - When trying to cancel an already cancelled transaction, or when a Deposit is made toward a cancelled withdraw.
// - `PAM_ERR_CANCEL_NON_WITHDRAW` - When trying to cancel a transaction that is not a Withdraw transaction.
// - `PAM_ERR_BET_NOT_ALLOWED` - When a bet cannot be done, eg when the user is blocked.
// - `PAM_ERR_PLAYER_NOT_FOUND` - When `playerId` is not found.
// - `PAM_ERR_API_TOKEN` - When `Authorization` header api token does not match the PAM api token.
// - `PAM_ERR_TRANS_ALREADY_SETTLED` - When trying to cancel an already Deposited bet or when trying to Deposit on an already finished gameRound, finished bet.
// - `PAM_ERR_DUPLICATE_TRANS` - When a Deposit is made with an already existing `providerTransactionId` but with different `playerId`/`providerGameId`/`providerRoundId`.
// - `PAM_ERR_PROMO_OVERDRAFT` - When user does not have enough funds on their promo account for a withdraw transaction.
// - `PAM_ERR_TIMEOUT` - A timeout occurred
type ErrorCode string

// GameRound Game round object
type GameRound struct {
	// EndTime A date and time in IS0 8601 format
	EndTime *Timestamp `json:"endTime,omitempty"`

	// ProviderGameId The game identifier unique for the RGS(provider)
	ProviderGameId ProviderGameId `json:"providerGameId"`

	// ProviderRoundId The unique game round identifier for the provider
	ProviderRoundId ProviderRoundId `json:"providerRoundId"`

	// StartTime A date and time in IS0 8601 format
	StartTime Timestamp `json:"startTime"`
}

// GameRoundResponse defines model for GameRoundResponse.
type GameRoundResponse struct {
	// Error Error details describing why PAM rejected the request
	Error *PamError `json:"error,omitempty"`

	// Gameround Game round object
	Gameround *GameRound `json:"gameround,omitempty"`
	Status    StatusCode `json:"status"`
}

// GetTransactionsResponse defines model for GetTransactionsResponse.
type GetTransactionsResponse struct {
	// Error Error details describing why PAM rejected the request
	Error        *PamError      `json:"error,omitempty"`
	Status       StatusCode     `json:"status"`
	Transactions *[]Transaction `json:"transactions,omitempty"`
}

// Jackpot defines model for Jackpot.
type Jackpot struct {
	// JackpotAmount Amount in some currency, rounded to 6 decimal places
	JackpotAmount  *Amount          `json:"jackpotAmount,omitempty"`
	JackpotBuckets *[]JackpotBucket `json:"jackpotBuckets,omitempty"`

	// JackpotId Jackpot identifier
	JackpotId *JackpotId `json:"jackpotId,omitempty"`

	// JackpotReference Jackpot reference, arbitrary use
	JackpotReference *JackpotReference `json:"jackpotReference,omitempty"`
}

// JackpotBucket defines model for JackpotBucket.
type JackpotBucket struct {
	// BucketAmount Amount in some currency, rounded to 6 decimal places
	BucketAmount *Amount `json:"bucketAmount,omitempty"`

	// BucketReference Jackpot bucket reference, arbitrary use
	BucketReference *BucketReference `json:"bucketReference,omitempty"`

	// BucketType Type of jackpot bucket, if any. Arbitrary use
	BucketType *BucketType `json:"bucketType,omitempty"`

	// Currency ISO 4217 three letter currency code
	Currency *Currency `json:"currency,omitempty"`
}

// JackpotId Jackpot identifier
type JackpotId = string

// JackpotReference Jackpot reference, arbitrary use
type JackpotReference = string

// Language ISO 639-1 two letter language code
type Language = string

// PamError Error details describing why PAM rejected the request
type PamError struct {
	// Code - `PAM_ERR_UNDEFINED` - When you need a generic error.
	// - `PAM_ERR_ACC_NOT_FOUND` - When account of `playerId` is not found.
	// - `PAM_ERR_GAME_NOT_FOUND` - When specified `providerGameId` is not found.
	// - `PAM_ERR_ROUND_NOT_FOUND` - In getGameRound, when there is no game round with id `providerGameRoundId`.
	// - `PAM_ERR_TRANS_NOT_FOUND` - In DEPOSIT transaction if the game round with id `providerRoundId` is not found.
	// - `PAM_ERR_CASH_OVERDRAFT` - When user does not have enough funds on their account for a withdraw transactions.
	// - `PAM_ERR_BONUS_OVERDRAFT` - When user does not have enough funds on their bonus account for a withdraw transaction.
	// - `PAM_ERR_SESSION_NOT_FOUND` - When no session is found for provided `X-Player-Token`.
	// - `PAM_ERR_SESSION_EXPIRED` - When session related to `X-Player-Token` has expired.
	// - `PAM_ERR_MISSING_PROVIDER` - When specified query parameter `provider` is not found.
	// - `PAM_ERR_TRANS_CURRENCY` - When specified `Currency` does not match that of the session.
	// - `PAM_ERR_NEGATIVE_STAKE` - When transaction amount is negative.
	// - `PAM_ERR_CANCEL_NOT_FOUND` - When the transaction trying to cancel doesn't exist.
	// - `PAM_ERR_TRANS_ALREADY_CANCELLED` - When trying to cancel an already cancelled transaction, or when a Deposit is made toward a cancelled withdraw.
	// - `PAM_ERR_CANCEL_NON_WITHDRAW` - When trying to cancel a transaction that is not a Withdraw transaction.
	// - `PAM_ERR_BET_NOT_ALLOWED` - When a bet cannot be done, eg when the user is blocked.
	// - `PAM_ERR_PLAYER_NOT_FOUND` - When `playerId` is not found.
	// - `PAM_ERR_API_TOKEN` - When `Authorization` header api token does not match the PAM api token.
	// - `PAM_ERR_TRANS_ALREADY_SETTLED` - When trying to cancel an already Deposited bet or when trying to Deposit on an already finished gameRound, finished bet.
	// - `PAM_ERR_DUPLICATE_TRANS` - When a Deposit is made with an already existing `providerTransactionId` but with different `playerId`/`providerGameId`/`providerRoundId`.
	// - `PAM_ERR_PROMO_OVERDRAFT` - When user does not have enough funds on their promo account for a withdraw transaction.
	// - `PAM_ERR_TIMEOUT` - A timeout occurred
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// PlayerId id of player
type PlayerId = string

// Promo defines model for Promo.
type Promo struct {
	// Currency ISO 4217 three letter currency code
	Currency *Currency `json:"currency,omitempty"`

	// PromoAmount Amount in some currency, rounded to 6 decimal places
	PromoAmount *Amount `json:"promoAmount,omitempty"`

	// PromoAmountTotal Amount in some currency, rounded to 6 decimal places
	PromoAmountTotal *Amount `json:"promoAmountTotal,omitempty"`

	// PromoAwardRef Offering award reference, if any
	PromoAwardRef *PromoAwardRef `json:"promoAwardRef,omitempty"`

	// PromoCode Offering code, if any
	PromoCode *PromoCode `json:"promoCode,omitempty"`

	// PromoConfigRef Offering configuration reference, if any
	PromoConfigRef *PromoConfigRef `json:"promoConfigRef,omitempty"`

	// PromoName Name of offering, if any
	PromoName *PromoName `json:"promoName,omitempty"`

	// PromoReference Reference to the offering, if any
	PromoReference *PromoReference `json:"promoReference,omitempty"`

	// PromoStatus Offering status, if any
	PromoStatus *PromoStatus `json:"promoStatus,omitempty"`

	// PromoType Promo types according to:
	// * `PROMOBONUS` - bonus promotion
	// * `PROMOTOURNAMENT` - tournament related promotion
	// * `PROMOFROMGAME` - promotion awarded from a game
	// * `PROMOCAP` - capped promotion
	// * `PROMOLIMIT` - limited promotion
	// * `PROMOMONEYREWARD` - real money promotion award
	// * `FREEROUNDS` - extra game rounds for free
	// * `FREESPINS` - extra spins for free
	PromoType *PromoType `json:"promoType,omitempty"`
}

// PromoAwardRef Offering award reference, if any
type PromoAwardRef = string

// PromoCode Offering code, if any
type PromoCode = string

// PromoConfigRef Offering configuration reference, if any
type PromoConfigRef = string

// PromoName Name of offering, if any
type PromoName = string

// PromoReference Reference to the offering, if any
type PromoReference = string

// PromoStatus Offering status, if any
type PromoStatus = string

// PromoType Promo types according to:
// * `PROMOBONUS` - bonus promotion
// * `PROMOTOURNAMENT` - tournament related promotion
// * `PROMOFROMGAME` - promotion awarded from a game
// * `PROMOCAP` - capped promotion
// * `PROMOLIMIT` - limited promotion
// * `PROMOMONEYREWARD` - real money promotion award
// * `FREEROUNDS` - extra game rounds for free
// * `FREESPINS` - extra spins for free
type PromoType string

// Provider Game provider identity known by the PAM and Valkyrie
type Provider = string

// ProviderBetRef Provider bet reference for grouping or matching transactions. Either this or `providerTransactionId` is required. This one is prioritized if both are present. It is used for RGS:s that encapsulate many transactions in a wrapper transaction.
type ProviderBetRef = string

// ProviderGameId The game identifier unique for the RGS(provider)
type ProviderGameId = string

// ProviderRoundId The unique game round identifier for the provider
type ProviderRoundId = string

// ProviderTransactionId The RGS transaction identifier. Unique for each provider. Either this or `providerBetRef` is required. `providerBetRef` will be prioritized if both are present.
type ProviderTransactionId = string

// RoundTransaction A transaction that's part of a game round. It has a limited set of fields as its intended use is when doing gamewise settlement.
type RoundTransaction struct {
	// CashAmount Amount in some currency, rounded to 6 decimal places
	CashAmount *Amount `json:"cashAmount,omitempty"`
	IsGameOver *bool   `json:"isGameOver,omitempty"`

	// JackpotContribution Amount in some currency, rounded to 6 decimal places
	JackpotContribution *Amount `json:"jackpotContribution,omitempty"`

	// ProviderBetRef Provider bet reference for grouping or matching transactions. Either this or `providerTransactionId` is required. This one is prioritized if both are present. It is used for RGS:s that encapsulate many transactions in a wrapper transaction.
	ProviderBetRef *ProviderBetRef `json:"providerBetRef,omitempty"`

	// ProviderTransactionId The RGS transaction identifier. Unique for each provider. Either this or `providerBetRef` is required. `providerBetRef` will be prioritized if both are present.
	ProviderTransactionId *ProviderTransactionId `json:"providerTransactionId,omitempty"`

	// TransactionDateTime A date and time in IS0 8601 format
	TransactionDateTime *Timestamp `json:"transactionDateTime,omitempty"`

	// TransactionType Transaction types according to:
	// * `DEPOSIT` - for adding funds
	// * `WITHDRAW` - subtract funds from an account balance. Generally for placing bets
	// * `CANCEL` - reverting a previous transaction
	// * `PROMODEPOSIT` - payout from promo and similar offerings programs
	// * `PROMOWITHDRAW` - buyin to promo and similar offerings programs
	// * `PROMOCANCEL` - reverting a previous promo transaction
	TransactionType TransactionType `json:"transactionType"`
}

// Session defines model for Session.
type Session struct {
	// Country ISO 3166-1 alpha-2 two letter country code
	Country Country `json:"country"`

	// Currency ISO 4217 three letter currency code
	Currency Currency `json:"currency"`

	// GameId The game identifier unique for the RGS(provider)
	GameId *ProviderGameId `json:"gameId,omitempty"`

	// Language ISO 639-1 two letter language code
	Language Language `json:"language"`

	// PlayerId id of player
	PlayerId PlayerId `json:"playerId"`

	// Token Player game session identifier
	Token SessionToken `json:"token"`
}

// SessionResponse defines model for SessionResponse.
type SessionResponse struct {
	// Error Error details describing why PAM rejected the request
	Error   *PamError  `json:"error,omitempty"`
	Session *Session   `json:"session,omitempty"`
	Status  StatusCode `json:"status"`
}

// SessionToken Player game session identifier
type SessionToken = string

// StatusCode defines model for StatusCode.
type StatusCode string

// Timestamp A date and time in IS0 8601 format
type Timestamp = time.Time

// Tip defines model for Tip.
type Tip struct {
	// TipAmount Amount in some currency, rounded to 6 decimal places
	TipAmount *Amount `json:"tipAmount,omitempty"`
}

// Transaction defines model for Transaction.
type Transaction struct {
	// BetCode metadata about what kind of bet/transaction it is
	BetCode *BetCode `json:"betCode,omitempty"`

	// BonusAmount Amount in some currency, rounded to 6 decimal places
	BonusAmount Amount `json:"bonusAmount"`

	// CashAmount Amount in some currency, rounded to 6 decimal places
	CashAmount Amount `json:"cashAmount"`

	// Currency ISO 4217 three letter currency code
	Currency   Currency   `json:"currency"`
	IsGameOver *bool      `json:"isGameOver,omitempty"`
	Jackpots   *[]Jackpot `json:"jackpots,omitempty"`

	// PromoAmount Amount in some currency, rounded to 6 decimal places
	PromoAmount Amount   `json:"promoAmount"`
	Promos      *[]Promo `json:"promos,omitempty"`

	// Provider Game provider identity known by the PAM and Valkyrie
	Provider Provider `json:"provider"`

	// ProviderBetRef Provider bet reference for grouping or matching transactions. Either this or `providerTransactionId` is required. This one is prioritized if both are present. It is used for RGS:s that encapsulate many transactions in a wrapper transaction.
	ProviderBetRef *ProviderBetRef `json:"providerBetRef,omitempty"`

	// ProviderGameId The game identifier unique for the RGS(provider)
	ProviderGameId *ProviderGameId `json:"providerGameId,omitempty"`

	// ProviderRoundId The unique game round identifier for the provider
	ProviderRoundId *ProviderRoundId `json:"providerRoundId,omitempty"`

	// ProviderTransactionId The RGS transaction identifier. Unique for each provider. Either this or `providerBetRef` is required. `providerBetRef` will be prioritized if both are present.
	ProviderTransactionId ProviderTransactionId `json:"providerTransactionId"`
	RoundTransactions     *[]RoundTransaction   `json:"roundTransactions,omitempty"`
	Tip                   *Tip                  `json:"tip,omitempty"`

	// TransactionDateTime A date and time in IS0 8601 format
	TransactionDateTime Timestamp `json:"transactionDateTime"`

	// TransactionType Transaction types according to:
	// * `DEPOSIT` - for adding funds
	// * `WITHDRAW` - subtract funds from an account balance. Generally for placing bets
	// * `CANCEL` - reverting a previous transaction
	// * `PROMODEPOSIT` - payout from promo and similar offerings programs
	// * `PROMOWITHDRAW` - buyin to promo and similar offerings programs
	// * `PROMOCANCEL` - reverting a previous promo transaction
	TransactionType TransactionType `json:"transactionType"`
}

// TransactionId Unique transaction identifier from the PAM system
type TransactionId = string

// TransactionResult defines model for TransactionResult.
type TransactionResult struct {
	// Balance player account balance
	Balance *Balance `json:"balance,omitempty"`

	// TransactionId Unique transaction identifier from the PAM system
	TransactionId *TransactionId `json:"transactionId,omitempty"`
}

// TransactionType Transaction types according to:
// * `DEPOSIT` - for adding funds
// * `WITHDRAW` - subtract funds from an account balance. Generally for placing bets
// * `CANCEL` - reverting a previous transaction
// * `PROMODEPOSIT` - payout from promo and similar offerings programs
// * `PROMOWITHDRAW` - buyin to promo and similar offerings programs
// * `PROMOCANCEL` - reverting a previous promo transaction
type TransactionType string

// CorrelationId defines model for correlationId.
type CorrelationId = string

// Traceparent defines model for traceparent.
type Traceparent = string

// Tracestate defines model for tracestate.
type Tracestate = string

// Unauthorized defines model for Unauthorized.
type Unauthorized struct {
	// Error Error details describing why PAM rejected the request
	Error struct {
		// Code Pam Error code "PAM_ERR_SESSION_NOT_FOUND" or "PAM_ERR_UNDEFINED"
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`

	// Status Status is Error
	Status string `json:"status"`
}

// GetSessionParams defines parameters for GetSession.
type GetSessionParams struct {
	// Provider Name of the game provider associated with the session
	Provider Provider `form:"provider" json:"provider"`

	// XPlayerToken Player game session identifier
	XPlayerToken SessionToken `json:"X-Player-Token"`

	// XCorrelationID Header for correlating requests between the services for debugging purposes and request tracing. The value will originate from the game providers that support request identification. Otherwise Valkyrie will generate a value.
	XCorrelationID CorrelationId `json:"X-Correlation-ID"`

	// Traceparent Describes the position of the incoming request in its trace graph. Further specified in https://www.w3.org/TR/trace-context
	Traceparent *Traceparent `json:"traceparent,omitempty"`

	// Tracestate Extends traceparent with vendor-specific data represented by a set of name/value pairs. Further specified in https://www.w3.org/TR/trace-context
	Tracestate *Tracestate `json:"tracestate,omitempty"`
}

// RefreshSessionParams defines parameters for RefreshSession.
type RefreshSessionParams struct {
	// Provider Name of the game provider associated with the session
	Provider Provider `form:"provider" json:"provider"`

	// XPlayerToken Player game session identifier
	XPlayerToken SessionToken `json:"X-Player-Token"`

	// XCorrelationID Header for correlating requests between the services for debugging purposes and request tracing. The value will originate from the game providers that support request identification. Otherwise Valkyrie will generate a value.
	XCorrelationID CorrelationId `json:"X-Correlation-ID"`

	// Traceparent Describes the position of the incoming request in its trace graph. Further specified in https://www.w3.org/TR/trace-context
	Traceparent *Traceparent `json:"traceparent,omitempty"`

	// Tracestate Extends traceparent with vendor-specific data represented by a set of name/value pairs. Further specified in https://www.w3.org/TR/trace-context
	Tracestate *Tracestate `json:"tracestate,omitempty"`
}

// GetBalanceParams defines parameters for GetBalance.
type GetBalanceParams struct {
	// Provider Name of the game provider associated with the session
	Provider Provider `form:"provider" json:"provider"`

	// XPlayerToken Player game session identifier
	XPlayerToken SessionToken `json:"X-Player-Token"`

	// XCorrelationID Header for correlating requests between the services for debugging purposes and request tracing. The value will originate from the game providers that support request identification. Otherwise Valkyrie will generate a value.
	XCorrelationID CorrelationId `json:"X-Correlation-ID"`

	// Traceparent Describes the position of the incoming request in its trace graph. Further specified in https://www.w3.org/TR/trace-context
	Traceparent *Traceparent `json:"traceparent,omitempty"`

	// Tracestate Extends traceparent with vendor-specific data represented by a set of name/value pairs. Further specified in https://www.w3.org/TR/trace-context
	Tracestate *Tracestate `json:"tracestate,omitempty"`
}

// GetGameRoundParams defines parameters for GetGameRound.
type GetGameRoundParams struct {
	// Provider Name of the game provider associated with the session
	Provider Provider `form:"provider" json:"provider"`

	// XPlayerToken Player game session identifier
	XPlayerToken SessionToken `json:"X-Player-Token"`

	// XCorrelationID Header for correlating requests between the services for debugging purposes and request tracing. The value will originate from the game providers that support request identification. Otherwise Valkyrie will generate a value.
	XCorrelationID CorrelationId `json:"X-Correlation-ID"`

	// Traceparent Describes the position of the incoming request in its trace graph. Further specified in https://www.w3.org/TR/trace-context
	Traceparent *Traceparent `json:"traceparent,omitempty"`

	// Tracestate Extends traceparent with vendor-specific data represented by a set of name/value pairs. Further specified in https://www.w3.org/TR/trace-context
	Tracestate *Tracestate `json:"tracestate,omitempty"`
}

// GetTransactionsParams defines parameters for GetTransactions.
type GetTransactionsParams struct {
	// Provider Name of the game provider associated with the session
	Provider              Provider               `form:"provider" json:"provider"`
	ProviderTransactionId *ProviderTransactionId `form:"providerTransactionId,omitempty" json:"providerTransactionId,omitempty"`
	ProviderBetRef        *ProviderBetRef        `form:"providerBetRef,omitempty" json:"providerBetRef,omitempty"`

	// XPlayerToken Player game session identifier
	XPlayerToken SessionToken `json:"X-Player-Token"`

	// XCorrelationID Header for correlating requests between the services for debugging purposes and request tracing. The value will originate from the game providers that support request identification. Otherwise Valkyrie will generate a value.
	XCorrelationID CorrelationId `json:"X-Correlation-ID"`

	// Traceparent Describes the position of the incoming request in its trace graph. Further specified in https://www.w3.org/TR/trace-context
	Traceparent *Traceparent `json:"traceparent,omitempty"`

	// Tracestate Extends traceparent with vendor-specific data represented by a set of name/value pairs. Further specified in https://www.w3.org/TR/trace-context
	Tracestate *Tracestate `json:"tracestate,omitempty"`
}

// AddTransactionParams defines parameters for AddTransaction.
type AddTransactionParams struct {
	// Provider Name of the game provider associated with the session
	Provider Provider `form:"provider" json:"provider"`

	// XPlayerToken Player game session identifier
	XPlayerToken SessionToken `json:"X-Player-Token"`

	// XCorrelationID Header for correlating requests between the services for debugging purposes and request tracing. The value will originate from the game providers that support request identification. Otherwise Valkyrie will generate a value.
	XCorrelationID CorrelationId `json:"X-Correlation-ID"`

	// Traceparent Describes the position of the incoming request in its trace graph. Further specified in https://www.w3.org/TR/trace-context
	Traceparent *Traceparent `json:"traceparent,omitempty"`

	// Tracestate Extends traceparent with vendor-specific data represented by a set of name/value pairs. Further specified in https://www.w3.org/TR/trace-context
	Tracestate *Tracestate `json:"tracestate,omitempty"`
}

// AddTransactionJSONRequestBody defines body for AddTransaction for application/json ContentType.
type AddTransactionJSONRequestBody = Transaction