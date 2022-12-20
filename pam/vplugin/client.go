package vplugin

import (
	"context"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

// PamClient Interface describing available PAM operations. The implementing plugins
// are expected to fulfill this interface.
type PAM interface {
	// GetSession Return session
	GetSession(pam.GetSessionRequest) *pam.SessionResponse
	// RefreshSession returns a new session token
	RefreshSession(pam.RefreshSessionRequest) *pam.SessionResponse
	// GetBalance get balance from PAM
	GetBalance(pam.GetBalanceRequest) *pam.BalanceResponse
	// GetTransactions get transactions from pam
	GetTransactions(pam.GetTransactionsRequest) *pam.GetTransactionsResponse
	// AddTransaction returns transactionId and balance. When transaction fails balance
	// can still be returned. On failure error will be returned
	AddTransaction(pam.AddTransactionRequest) *pam.AddTransactionResponse
	// GetGameRound gets gameRound from PAM
	GetGameRound(pam.GetGameRoundRequest) *pam.GameRoundResponse

	PluginControl
}

func init() {
	pam.ClientFactory().
		Register("vplugin", func(args pam.ClientArgs) (pam.PamClient, error) {
			return Create(args.Context, args.Config)
		})
}

type PAMPlugin struct {
	plugin PAM
}

func Create(ctx context.Context, cfg map[string]any) (*PAMPlugin, error) {
	config, err := pam.GetConfig[pluginConfig](cfg)
	if err != nil {
		return nil, err
	}

	plugin, err := start(ctx, config.Type, config.PluginPath)
	if err != nil {
		return nil, err
	}

	err = plugin.Init(config.Init)
	if err != nil {
		return nil, err
	}

	return &PAMPlugin{plugin: plugin}, nil
}

func (vp *PAMPlugin) GetSession(rm pam.GetSessionRequestMapper) (*pam.Session, error) {
	_, req, err := rm()
	if err != nil {
		return nil, err
	}
	resp := vp.plugin.GetSession(req)

	if err = handleErrors(resp.Error, err, resp.Session); err != nil {
		return nil, err
	}

	return resp.Session, nil
}

func (vp *PAMPlugin) RefreshSession(rm pam.RefreshSessionRequestMapper) (*pam.Session, error) {
	_, req, err := rm()
	if err != nil {
		return nil, err
	}
	resp := vp.plugin.RefreshSession(req)
	if err = handleErrors(resp.Error, err, resp.Session); err != nil {
		return nil, err
	}
	return resp.Session, nil
}

func (vp *PAMPlugin) GetBalance(rm pam.GetBalanceRequestMapper) (*pam.Balance, error) {
	_, req, err := rm()
	if err != nil {
		return nil, err
	}
	resp := vp.plugin.GetBalance(req)
	if err = handleErrors(resp.Error, err, resp.Balance); err != nil {
		return nil, err
	}
	return resp.Balance, nil
}

func (vp *PAMPlugin) GetTransactions(rm pam.GetTransactionsRequestMapper) ([]pam.Transaction, error) {
	_, req, err := rm()
	if err != nil {
		return nil, err
	}
	resp := vp.plugin.GetTransactions(req)
	if err = handleErrors(resp.Error, err, resp.Transactions); err != nil {
		return nil, err
	}
	return *resp.Transactions, nil
}

func (vp *PAMPlugin) AddTransaction(rm pam.AddTransactionRequestMapper) (*pam.TransactionResult, error) {
	_, req, err := rm(pam.SixDecimalRounder)
	if err != nil {
		return nil, err
	}
	resp := vp.plugin.AddTransaction(*req)
	if err = handleErrors(resp.Error, err, resp.TransactionResult); err != nil {
		if resp.TransactionResult != nil {
			return resp.TransactionResult, err
		}
		return nil, err
	}
	return resp.TransactionResult, nil
}

func (vp *PAMPlugin) GetGameRound(rm pam.GetGameRoundRequestMapper) (*pam.GameRound, error) {
	_, req, err := rm()
	if err != nil {
		return nil, err
	}
	resp := vp.plugin.GetGameRound(req)
	if err = handleErrors(resp.Error, err, resp.Gameround); err != nil {
		return nil, err
	}
	return resp.Gameround, nil
}
