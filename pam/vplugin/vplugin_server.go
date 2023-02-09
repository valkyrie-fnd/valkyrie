package vplugin

import (
	"fmt"

	"github.com/valkyrie-fnd/valkyrie/pam"
)

type VPluginRPCServer struct {
	Impl PAM
}

func (pp *VPluginRPCServer) Init(args any, resp *any) error {
	cfg, ok := args.(PluginInitConfig)
	if !ok {
		return fmt.Errorf("Init config not valid, %v", args)
	}

	return pp.Impl.Init(cfg)
}

func (pp *VPluginRPCServer) GetBalance(args any, response *pam.BalanceResponse) error {
	req, ok := args.(pam.GetBalanceRequest)
	if !ok {
		return fmt.Errorf("invalid request, %v", args)
	}
	balance := pp.Impl.GetBalance(req)
	*response = *balance
	return nil
}

func (pp *VPluginRPCServer) GetSession(args any, resp *pam.SessionResponse) error {
	req, ok := args.(pam.GetSessionRequest)
	if !ok {
		return fmt.Errorf("invalid request, %v", args)
	}
	result := pp.Impl.GetSession(req)
	*resp = *result
	return nil
}

func (pp *VPluginRPCServer) RefreshSession(args any, resp *pam.SessionResponse) error {
	req, ok := args.(pam.RefreshSessionRequest)
	if !ok {
		return fmt.Errorf("invalid request, %v", args)
	}
	result := pp.Impl.RefreshSession(req)
	*resp = *result
	return nil
}

func (pp *VPluginRPCServer) GetTransactions(args any, resp *pam.GetTransactionsResponse) error {
	req, ok := args.(pam.GetTransactionsRequest)
	if !ok {
		return fmt.Errorf("invalid request, %v", args)
	}
	result := pp.Impl.GetTransactions(req)
	*resp = *result
	return nil
}

func (pp *VPluginRPCServer) AddTransaction(args any, resp *pam.AddTransactionResponse) error {
	req, ok := args.(pam.AddTransactionRequest)
	if !ok {
		return fmt.Errorf("invalid request, %v", args)
	}
	result := pp.Impl.AddTransaction(req)
	*resp = *result
	return nil
}

func (pp *VPluginRPCServer) GetGameRound(args any, resp *pam.GameRoundResponse) error {
	req, ok := args.(pam.GetGameRoundRequest)
	if !ok {
		return fmt.Errorf("invalid request, %v", args)
	}
	result := pp.Impl.GetGameRound(req)
	*resp = *result
	return nil
}

func (pp *VPluginRPCServer) GetSettlementType(args any, resp *pam.SettlementType) error {
	result := pp.Impl.GetSettlementType()
	*resp = result
	return nil
}
