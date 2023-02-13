package vplugin

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"github.com/valkyrie-fnd/valkyrie/pam/routine"
)

type PluginControl interface {
	// Init passes configuration to plugins which are expected to report
	// any startup issues as errors.
	Init(PluginInitConfig) error
}

var (
	MagicCookieKey = "BASIC_VPLUGIN"
	// MagicCookieValue is not sensitive, it's used to do a basic handshake between
	// a plugin and host. If the handshake fails, a user-friendly error is shown.
	// It is a UX feature, not a security feature.
	MagicCookieValue = "ff912fb194609a432cfe8d504951f133a0a21b18"
)

var PluginConfig func(string, string) plugin.ClientConfig = func(name, path string) plugin.ClientConfig {
	return plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   MagicCookieKey,
			MagicCookieValue: MagicCookieValue,
		},
		Plugins: map[string]plugin.Plugin{
			name: &VPlugin{},
		},
		Cmd:    exec.Command(path),
		Logger: NewZerologAdapter(),
	}
}

func start(ctx context.Context, name, path string) (PAM, error) {
	clientConfig := PluginConfig(name, path)

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&clientConfig)
	killOnDone(ctx, client)

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(name)
	if err != nil {
		return nil, err
	}

	plugin, ok := raw.(PAM)
	if !ok {
		return nil, fmt.Errorf("Vplugin [%s] at [%s] does not fullfil PAM interface", name, path)
	}

	return plugin, nil
}

func killOnDone(ctx context.Context, client *plugin.Client) {
	routine.Go(func() {
		defer client.Kill()
		<-ctx.Done()
	})
}
