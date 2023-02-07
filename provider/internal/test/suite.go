// Package test contains shared test code for providers
package test

import (
	"context"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
	"github.com/valkyrie-fnd/valkyrie-stubs/backdoors"
	"github.com/valkyrie-fnd/valkyrie-stubs/datastore"
	"github.com/valkyrie-fnd/valkyrie-stubs/genericpam"
	"github.com/valkyrie-fnd/valkyrie-stubs/memorydatastore"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
	"github.com/valkyrie-fnd/valkyrie/server"
)

const (
	baseAddr = "localhost:%d"
	baseURL  = "http://" + baseAddr
)

type IntegrationTestSuite struct {
	suite.Suite
	ProviderConfigFn func(store datastore.ExtendedDatastore) configs.ProviderConf
	valkyrie         *server.Valkyrie
	backdoorServer   *fiber.App
	pamServer        *fiber.App
	ProviderConfig   configs.ProviderConf
	ValkyrieURL      string
	BackdoorURL      string
}

func (s *IntegrationTestSuite) SetupSuite() {
	config, _ := memorydatastore.ReadConfig("testdata/datastore.config.yaml")

	dataStore := memorydatastore.NewMapDataStore(config)

	if v, found := os.LookupEnv("PAM_TOKEN"); found {
		dataStore.AddPamApiToken(v)
	}

	var pamURL string
	if v, found := os.LookupEnv("PAM_URL"); found {
		pamURL = v
	} else {
		pamPort, _ := testutils.GetFreePort()
		s.pamServer = genericpam.RunServer(dataStore, genericpam.Config{
			PamApiKey:      dataStore.GetPamApiToken(),
			ProviderTokens: dataStore.GetProviderTokens(),
			Address:        fmt.Sprintf(baseAddr, pamPort)})
		pamURL = fmt.Sprintf(baseURL, pamPort)
	}

	if v, found := os.LookupEnv("BACKDOOR_URL"); found {
		s.BackdoorURL = v
	} else {
		backdoorPort, _ := testutils.GetFreePort()
		s.backdoorServer, s.BackdoorURL = backdoors.BackdoorServer(dataStore, fmt.Sprintf(baseAddr, backdoorPort))
	}

	if v, found := os.LookupEnv("VALKYRIE_URL"); found {
		s.ValkyrieURL = v
	} else {
		providerPort, _ := testutils.GetFreePort()
		operatorPort, _ := testutils.GetFreePort()

		s.ProviderConfig = s.ProviderConfigFn(dataStore)
		valkyrieConfig := configs.ValkyrieConfig{
			Logging: configs.LogConfig{
				Level: "fatal",
			},
			ProviderBasePath: "/providers",
			Providers: []configs.ProviderConf{
				s.ProviderConfig,
			},
			Pam: configs.PamConf{
				"name":    "generic",
				"url":     pamURL,
				"api_key": dataStore.GetPamApiToken(),
			},
			HTTPServer: configs.HTTPServerConfig{
				ProviderAddress: fmt.Sprintf(baseAddr, providerPort),
				OperatorAddress: fmt.Sprintf(baseAddr, operatorPort),
			},
		}

		s.valkyrie = server.NewValkyrie(context.TODO(), &valkyrieConfig)

		s.valkyrie.Start()

		s.ValkyrieURL = fmt.Sprintf(baseURL+"%s%s", providerPort, valkyrieConfig.ProviderBasePath, s.ProviderConfig.BasePath)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.valkyrie != nil {
		s.valkyrie.Stop()
	}
	if s.backdoorServer != nil {
		_ = s.backdoorServer.Shutdown()
	}
	if s.pamServer != nil {
		_ = s.pamServer.Shutdown()
	}
}
