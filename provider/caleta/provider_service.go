package caleta

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider/caleta/auth"
	"github.com/valkyrie-fnd/valkyrie/rest"
)

type caletaService struct {
	client       rest.HTTPClientJSONInterface
	headerSigner headerSigner
	authConfig   AuthConf
	config       configs.ProviderConf
	caletaConfig caletaConf
}

func NewCaletaService(config configs.ProviderConf, client rest.HTTPClientJSONInterface) (*caletaService, error) {
	authConfig, err := getAuthConf(config)
	if err != nil {
		return nil, err
	}

	caletaConfig, err := getCaletaConf(config)
	if err != nil {
		return nil, err
	}

	hs, err := newHeaderSigner(authConfig)
	if err != nil {
		return nil, err
	}

	return &caletaService{
		config:       config,
		client:       client,
		authConfig:   authConfig,
		headerSigner: hs,
		caletaConfig: caletaConfig,
	}, nil
}

type headerSigner interface {
	sign(body any, headers map[string]string) error
}

func newHeaderSigner(authConfig AuthConf) (headerSigner, error) {
	if authConfig.SigningKey != "" {
		sig, err := NewSigner([]byte(authConfig.SigningKey))
		if err != nil {
			return nil, err
		}

		return &authHeaderSigner{
			signer: sig,
		}, nil
	} else {
		log.Warn().Msg("Missing Caleta provider 'signing_key' config, skipping header signing")
		return &noopHeaderSigner{}, nil
	}
}

type authHeaderSigner struct {
	signer auth.Signer
}

func (s *authHeaderSigner) sign(body any, headers map[string]string) error {
	byteBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	signature, err := s.signer.Sign(byteBody)
	if err != nil {
		return err
	}

	headers["X-Auth-Signature"] = string(signature)

	return nil
}

type noopHeaderSigner struct{}

func (_ *noopHeaderSigner) sign(_ any, _ map[string]string) error {
	return nil
}
