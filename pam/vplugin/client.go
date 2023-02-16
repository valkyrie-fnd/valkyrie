package vplugin

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/ops"
	"github.com/valkyrie-fnd/valkyrie/pam"
)

const (
	tracerName = "vplugin-client"
	RPCSystem  = "net/rpc"
	RPCService = "vplugin.PluginPAM"
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

	ctx, span := startSpan(ctx, "GetSession")
	defer span.End()
	req.Params.Traceparent, req.Params.Tracestate = getTracingFromContext(ctx)

	resp := vp.plugin.GetSession(req)

	if err = handleErrors(resp.Error, err, resp.Session); err != nil {
		return nil, err
	}

	return resp.Session, nil
}

func (vp *PluginPAM) RefreshSession(rm pam.RefreshSessionRequestMapper) (*pam.Session, error) {
	ctx, req, err := rm()
	if err != nil {
		return nil, err
	}

	ctx, span := startSpan(ctx, "RefreshSession")
	defer span.End()
	req.Params.Traceparent, req.Params.Tracestate = getTracingFromContext(ctx)

	resp := vp.plugin.RefreshSession(req)
	if err = handleErrors(resp.Error, err, resp.Session); err != nil {
		return nil, err
	}
	return resp.Session, nil
}

func (vp *PluginPAM) GetBalance(rm pam.GetBalanceRequestMapper) (*pam.Balance, error) {
	ctx, req, err := rm()

	ctx, span := startSpan(ctx, "GetBalance")
	defer span.End()
	req.Params.Traceparent, req.Params.Tracestate = getTracingFromContext(ctx)

	if err != nil {
		return nil, err
	}
	resp := vp.plugin.GetBalance(req)
	if err = handleErrors(resp.Error, err, resp.Balance); err != nil {
		return nil, err
	}
	return resp.Balance, nil
}

func (vp *PluginPAM) GetTransactions(rm pam.GetTransactionsRequestMapper) ([]pam.Transaction, error) {
	ctx, req, err := rm()

	ctx, span := startSpan(ctx, "GetTransactions")
	defer span.End()
	req.Params.Traceparent, req.Params.Tracestate = getTracingFromContext(ctx)

	if err != nil {
		return nil, err
	}
	resp := vp.plugin.GetTransactions(req)
	if err = handleErrors(resp.Error, err, resp.Transactions); err != nil {
		return nil, err
	}
	return *resp.Transactions, nil
}

func (vp *PluginPAM) AddTransaction(rm pam.AddTransactionRequestMapper) (*pam.TransactionResult, error) {
	ctx, req, err := rm(pam.SixDecimalRounder)
	if err != nil {
		return nil, err
	}

	ctx, span := startSpan(ctx, "AddTransaction")
	defer span.End()
	req.Params.Traceparent, req.Params.Tracestate = getTracingFromContext(ctx)

	resp := vp.plugin.AddTransaction(*req)
	if err = handleErrors(resp.Error, err, resp.TransactionResult); err != nil {
		if resp.TransactionResult != nil {
			return resp.TransactionResult, err
		}
		return nil, err
	}
	return resp.TransactionResult, nil
}

func (vp *PluginPAM) GetGameRound(rm pam.GetGameRoundRequestMapper) (*pam.GameRound, error) {
	ctx, req, err := rm()
	if err != nil {
		return nil, err
	}

	ctx, span := startSpan(ctx, "GetGameRound")
	defer span.End()
	req.Params.Traceparent, req.Params.Tracestate = getTracingFromContext(ctx)

	resp := vp.plugin.GetGameRound(req)
	if err = handleErrors(resp.Error, err, resp.Gameround); err != nil {
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

func getTracingFromContext(ctx context.Context) (traceparent *pam.Traceparent, tracestate *pam.Tracestate) {
	tracingHeaders := ops.GetTracingHeaders(ctx)

	if value, found := tracingHeaders["traceparent"]; found {
		traceparent = &value
	}

	if value, found := tracingHeaders["tracestate"]; found {
		tracestate = &value
	}

	return traceparent, tracestate
}

func startSpan(ctx context.Context, fnName string) (context.Context, trace.Span) {
	// attributes from https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/rpc.md#common-remote-procedure-call-conventions
	return otel.Tracer(tracerName).Start(ctx, fmt.Sprintf("%s/%s", RPCService, fnName), trace.WithAttributes(
		semconv.RPCMethodKey.String(fnName),
		semconv.RPCSystemKey.String(RPCSystem),
		semconv.RPCServiceKey.String(RPCService),
	))
}
