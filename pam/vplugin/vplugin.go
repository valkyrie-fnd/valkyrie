// Package vplugin contains the generic and externalized plugin interface. This
// allows closed source implementations to be used with valkyrie as plugins.
package vplugin

import (
	"encoding/gob"
	"errors"
	"net/rpc"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

func init() {
	RegisterGobs()
}

func RegisterGobs() {
	// While using gob the request types needs to be registered with
	// a name. The same name needs to be registered at the receiving end.
	gob.RegisterName("pam.GetSessionRequest", pam.GetSessionRequest{})
	gob.RegisterName("pam.RefreshSessionRequest", pam.RefreshSessionRequest{})
	gob.RegisterName("pam.GetBalanceRequest", pam.GetBalanceRequest{})
	gob.RegisterName("pam.GetTransactionsRequest", pam.GetTransactionsRequest{})
	gob.RegisterName("pam.AddTransactionRequest", pam.AddTransactionRequest{})
	gob.RegisterName("pam.GetGameRoundRequest", pam.GetGameRoundRequest{})

	gob.RegisterName("pam.Session", pam.SessionResponse{})
	gob.RegisterName("pam.Balance", pam.BalanceResponse{})

	gob.RegisterName("GetTransactionsResponse", pam.GetTransactionsResponse{})
	gob.RegisterName("pam.AddTransactionResponse", pam.AddTransactionResponse{})
	gob.RegisterName("pam.GameRoundResponse", pam.GameRoundResponse{})

	gob.RegisterName("pam.Amount", pam.Amount{})
	gob.RegisterName("vplugin.PluginInitConfig", PluginInitConfig{})
}

type VPlugin struct {
	client *rpc.Client
}

func (vp *VPlugin) Init(cfg PluginInitConfig) error {
	return callWithLogging(vp.client, "Plugin.Init", cfg, new(interface{}))
}

func (vp *VPlugin) GetSession(req pam.GetSessionRequest) *pam.SessionResponse {
	var response pam.SessionResponse
	err := callWithLogging(vp.client, "Plugin.GetSession", req, &response)
	if err != nil {
		response.Status = pam.ERROR
		response.Error = wrapError(err)
	}
	return &response
}

func (vp *VPlugin) RefreshSession(req pam.RefreshSessionRequest) *pam.SessionResponse {
	var response pam.SessionResponse
	err := callWithLogging(vp.client, "Plugin.RefreshSession", req, &response)
	if err != nil {
		response.Status = pam.ERROR
		response.Error = wrapError(err)
	}
	return &response
}

func (vp *VPlugin) GetBalance(req pam.GetBalanceRequest) *pam.BalanceResponse {
	var response pam.BalanceResponse
	err := callWithLogging(vp.client, "Plugin.GetBalance", req, &response)
	if err != nil {
		response.Status = pam.ERROR
		response.Error = wrapError(err)
	}
	return &response
}

func (vp *VPlugin) GetTransactions(req pam.GetTransactionsRequest) *pam.GetTransactionsResponse {
	var response pam.GetTransactionsResponse
	err := callWithLogging(vp.client, "Plugin.GetTransactions", req, &response)
	if err != nil {
		response.Status = pam.ERROR
		response.Error = wrapError(err)
	}
	return &response
}

func (vp *VPlugin) AddTransaction(req pam.AddTransactionRequest) *pam.AddTransactionResponse {
	var response pam.AddTransactionResponse
	err := callWithLogging(vp.client, "Plugin.AddTransaction", req, &response)
	if err != nil {
		response.Status = pam.ERROR
		response.Error = wrapError(err)
	}
	return &response
}

func (vp *VPlugin) GetGameRound(req pam.GetGameRoundRequest) *pam.GameRoundResponse {
	var response pam.GameRoundResponse
	err := callWithLogging(vp.client, "Plugin.GetGameRound", req, &response)
	if err != nil {
		response.Status = pam.ERROR
		response.Error = wrapError(err)
	}
	return &response
}

// Server func is required by plugin but implemented at the receiving side
func (p *VPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, errors.New("not for use in host mode")
}

// Client func is part of plugin.Plugin interface
func (VPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &VPlugin{client: c}, nil
}

func callWithLogging(rpc *rpc.Client, method string, params any, response any) error {
	l := log.With().Str("vplugin.call", method).Logger()

	var tt time.Time
	l.Trace().Func(func(e *zerolog.Event) {
		e.Interface("request", params)
		tt = time.Now()
	})

	err := rpc.Call(method, &params, response)

	l.Trace().Func(func(e *zerolog.Event) {
		e.Dur("timing", time.Since(tt))
	})

	if err != nil {
		l.Error().Interface("response", response).Err(err)
	} else {
		l.Trace().Interface("response", response).Msg("plugin called")
	}
	return err
}

// wrapError helps comply with the error-less interface of PAM,
// by wrapping hard errors in pamErrors
func wrapError(err error) *pam.PamError {
	return &pam.PamError{
		Code:    pam.PAMERRUNDEFINED,
		Message: err.Error(),
	}
}
