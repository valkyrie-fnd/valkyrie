package provider

import (
	"sync"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/httpclient"
	"github.com/valkyrie-fnd/valkyrie/pam"
)

// ProviderArgs composes all arguments required to build a provider router
type ProviderArgs struct {
	PamClient  pam.PamClient
	HTTPClient httpclient.HTTPClientJSONInterface
	Config     configs.ProviderConf
}

type OperatorArgs struct {
	HTTPClient httpclient.HTTPClientJSONInterface
	Config     configs.ProviderConf
}

type providerFactory = configs.AbstractFactory[ProviderArgs, *Router]
type operatorFactory = configs.AbstractFactory[OperatorArgs, *Router]

var (
	providerOnce            sync.Once
	providerFactoryInstance *providerFactory

	operatorOnce            sync.Once
	operatorFactoryInstance *operatorFactory
)

// ProviderFactory returns a single instance to the provider router factory
func ProviderFactory() *providerFactory {
	// Make the providerFactory a singleton
	providerOnce.Do(func() {
		providerFactoryInstance = configs.NewAbstractFactory[ProviderArgs, *Router]()
	})

	return providerFactoryInstance
}

// OperatorFactory returns a single instance to the operator router factory
func OperatorFactory() *operatorFactory {
	// Make the operatorFactory a singleton
	operatorOnce.Do(func() {
		operatorFactoryInstance = configs.NewAbstractFactory[OperatorArgs, *Router]()
	})

	return operatorFactoryInstance
}
