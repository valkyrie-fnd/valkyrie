package provider

import (
	"sync"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/valkhttp"
)

// ProviderArgs composes all arguments required to build a provider router
type ProviderArgs struct {
	PamClient  pam.PamClient
	HTTPClient valkhttp.HTTPClient
	Config     configs.ProviderConf
}

type OperatorArgs struct {
	HTTPClient valkhttp.HTTPClient
	Config     configs.ProviderConf
}

type providerFactory = internal.AbstractFactory[ProviderArgs, *Router]
type operatorFactory = internal.AbstractFactory[OperatorArgs, *Router]

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
		providerFactoryInstance = internal.NewAbstractFactory[ProviderArgs, *Router]()
	})

	return providerFactoryInstance
}

// OperatorFactory returns a single instance to the operator router factory
func OperatorFactory() *operatorFactory {
	// Make the operatorFactory a singleton
	operatorOnce.Do(func() {
		operatorFactoryInstance = internal.NewAbstractFactory[OperatorArgs, *Router]()
	})

	return operatorFactoryInstance
}
