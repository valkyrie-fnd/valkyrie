package pam

import (
	"context"
	"fmt"
	"sync"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/httpclient"
)

// ClientArgs composes all arguments required to build a pam client
type ClientArgs struct {
	Context     context.Context
	Client      httpclient.HTTPClientJSONInterface
	Config      configs.PamConf
	TraceConfig configs.TraceConfig
	LogConfig   configs.LogConfig
}

type clientFactory = configs.AbstractFactory[ClientArgs, PamClient]

var (
	once    sync.Once
	factory *clientFactory
)

// ClientFactory returns a single instance to the pam client factory
func ClientFactory() *clientFactory {
	// Make the factory a singleton
	once.Do(func() {
		factory = configs.NewAbstractFactory[ClientArgs, PamClient]()
	})

	return factory
}

func GetPamClient(args ClientArgs) (PamClient, error) {
	pamName, err := GetName(args.Config)
	if err != nil {
		return nil, fmt.Errorf("unknown pam client: %w", err)
	}
	pamClient, err := ClientFactory().Build(pamName, args)
	if err != nil {
		return nil, fmt.Errorf("unable to build pam client: %w", err)
	}
	return pamClient, nil
}
