package vplugin

import (
	"context"
	"fmt"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal"
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
	// GetTransactionSupplier return the type of transaction supplier the PAM supports
	GetTransactionSupplier() pam.TransactionSupplier
	PluginControl
}

func init() {
	pam.ClientFactory().
		Register("vplugin", func(args pam.ClientArgs) (pam.PamClient, error) {
			return Create(args.Context, getPamConf(args))
		})
}

var Pipeline = internal.NewPipeline[any]()

type PluginPAM struct {
	plugin              PAM
	transactionSupplier pam.TransactionSupplier
}

func Create(ctx context.Context, cfg configs.PamConf) (*PluginPAM, error) {
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

	// Call the server and get the transaction supplier, this needs to be done only once.
	transactionSupplier := plugin.GetTransactionSupplier()
	if transactionSupplier == "" {
		return nil, fmt.Errorf("Could not get PAM transaction supplier")
	}
	return &PluginPAM{plugin: plugin, transactionSupplier: transactionSupplier}, nil
}

func (vp *PluginPAM) GetSession(rm pam.GetSessionRequestMapper) (*pam.Session, error) {
	ctx, req, err := rm()
	if err != nil {
		return nil, err
	}

	var resp *pam.SessionResponse
	err = Pipeline.Execute(ctx, &req,
		func(pc internal.PipelineContext[any]) error {
			resp = vp.plugin.GetSession(req)
			return handleErrors(resp.Error, err, resp.Session)
		})
	if err != nil {
		return nil, err
	}

	return resp.Session, nil
}

func (vp *PluginPAM) RefreshSession(rm pam.RefreshSessionRequestMapper) (*pam.Session, error) {
	ctx, req, err := rm()
	if err != nil {
		return nil, err
	}

	var resp *pam.SessionResponse
	err = Pipeline.Execute(ctx, &req,
		func(pc internal.PipelineContext[any]) error {
			resp = vp.plugin.RefreshSession(req)
			return handleErrors(resp.Error, err, resp.Session)
		})
	if err != nil {
		return nil, err
	}

	return resp.Session, nil
}

func (vp *PluginPAM) GetBalance(rm pam.GetBalanceRequestMapper) (*pam.Balance, error) {
	ctx, req, err := rm()

	var resp *pam.BalanceResponse
	err = Pipeline.Execute(ctx, &req,
		func(pc internal.PipelineContext[any]) error {
			resp = vp.plugin.GetBalance(req)
			return handleErrors(resp.Error, err, resp.Balance)
		})
	if err != nil {
		return nil, err
	}

	return resp.Balance, nil
}

func (vp *PluginPAM) GetTransactions(rm pam.GetTransactionsRequestMapper) ([]pam.Transaction, error) {
	ctx, req, err := rm()

	var resp *pam.GetTransactionsResponse
	err = Pipeline.Execute(ctx, &req,
		func(pc internal.PipelineContext[any]) error {
			resp = vp.plugin.GetTransactions(req)
			return handleErrors(resp.Error, err, resp.Transactions)
		})
	if err != nil {
		return nil, err
	}

	return *resp.Transactions, nil
}

func (vp *PluginPAM) AddTransaction(rm pam.AddTransactionRequestMapper) (*pam.TransactionResult, error) {
	ctx, req, err := rm(pam.SixDecimalRounder)
	if err != nil {
		return nil, err
	}

	var resp *pam.AddTransactionResponse
	err = Pipeline.Execute(ctx, req,
		func(pc internal.PipelineContext[any]) error {
			resp = vp.plugin.AddTransaction(*req)
			return handleErrors(resp.Error, err, resp.TransactionResult)
		})
	if err != nil {
		return nil, err
	}

	return resp.TransactionResult, nil
}

func (vp *PluginPAM) GetGameRound(rm pam.GetGameRoundRequestMapper) (*pam.GameRound, error) {
	ctx, req, err := rm()
	if err != nil {
		return nil, err
	}

	var resp *pam.GameRoundResponse
	err = Pipeline.Execute(ctx, &req,
		func(pc internal.PipelineContext[any]) error {
			resp = vp.plugin.GetGameRound(req)
			return handleErrors(resp.Error, err, resp.Gameround)
		})
	if err != nil {
		return nil, err
	}

	return resp.Gameround, nil
}

func (vp *PluginPAM) GetTransactionSupplier() pam.TransactionSupplier {
	return vp.transactionSupplier // This has been initialized in the Init() call
}

// getPamConf adds logging and tracing configuration to PamConf if missing.
func getPamConf(args pam.ClientArgs) configs.PamConf {
	cfg := args.Config

	if _, found := cfg["logging"]; !found {
		cfg["logging"] = args.LogConfig
	}

	if _, found := cfg["tracing"]; !found {
		cfg["tracing"] = args.TraceConfig
	}

	return cfg
}
