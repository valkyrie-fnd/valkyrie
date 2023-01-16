package pam

import (
	"context"
)

// RefreshSessionRequest bundles everything needed to make a request
type RefreshSessionRequest struct {
	Params RefreshSessionParams
}

// GetBalanceRequest bundles everything needed to make a request
type GetBalanceRequest struct {
	Params   GetBalanceParams
	PlayerID PlayerId
}

// GetTransactionsRequest bundles everything needed to make a request
type GetTransactionsRequest struct {
	Params   GetTransactionsParams
	PlayerID PlayerId
}

// AddTransactionRequest bundles everything needed to make a request
type AddTransactionRequest struct {
	Params   AddTransactionParams
	Body     AddTransactionJSONRequestBody
	PlayerID PlayerId
}

// GetGameRoundRequest bundles everything needed to make a request
type GetGameRoundRequest struct {
	ProviderRoundID ProviderRoundId
	Params          GetGameRoundParams
	PlayerID        PlayerId
}

// GetSessionRequest bundles everything needed to make a get session request
type GetSessionRequest struct {
	Params GetSessionParams
}

// RefreshSessionRequestMapper Returns context and request used by PAM
type RefreshSessionRequestMapper func() (context.Context, RefreshSessionRequest, error)

// GetBalanceRequestMapper Returns context and request used by PAM
type GetBalanceRequestMapper func() (context.Context, GetBalanceRequest, error)

// GetTransactionsRequestMapper Returns context and request used by PAM
type GetTransactionsRequestMapper func() (context.Context, GetTransactionsRequest, error)

// AddTransactionRequestMapper Returns context and request used by PAM
type AddTransactionRequestMapper func(AmountRounder) (context.Context, *AddTransactionRequest, error)

// GetGameRoundRequestMapper Returns context and request used by PAM
type GetGameRoundRequestMapper func() (context.Context, GetGameRoundRequest, error)

// GetSessionRequestMapper Returns context and request used by PAM
type GetSessionRequestMapper func() (context.Context, GetSessionRequest, error)

// PamClient Interface describing available PAM operations. The Mapper methods are indicating
// that explicit conversion is required for Provider data to work with the PAM.
type PamClient interface {
	// GetSession Return session
	GetSession(GetSessionRequestMapper) (*Session, error)
	// RefreshSession returns a new session token
	RefreshSession(RefreshSessionRequestMapper) (*Session, error)
	// GetBalance get balance from PAM
	GetBalance(GetBalanceRequestMapper) (*Balance, error)
	// GetTransactions get transactions from pam
	GetTransactions(GetTransactionsRequestMapper) ([]Transaction, error)
	// AddTransaction returns transactionId and balance. When transaction fails balance can still be returned. On failure error will be returned
	AddTransaction(AddTransactionRequestMapper) (*TransactionResult, error)
	// GetGameRound gets gameRound from PAM
	GetGameRound(GetGameRoundRequestMapper) (*GameRound, error)
}

// AmountRounder provides rounding requirements and is used for verifying
// that amounts passed to PAM clients are within acceptable precision.
//
// CheckPrecision rounds the supplied amount to the acceptable precision. If
// precision will be lost in the process a RoundingError will be returned instead.
type AmountRounder func(amt Amt) (*Amount, error)
