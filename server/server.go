package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/valkyrie-fnd/valkyrie/internal/routine"
	"github.com/valkyrie-fnd/valkyrie/rest"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/ops"
	"github.com/valkyrie-fnd/valkyrie/pam"
	"github.com/valkyrie-fnd/valkyrie/routes"

	_ "github.com/valkyrie-fnd/valkyrie/pam/genericpam" // init generic pam
	_ "github.com/valkyrie-fnd/valkyrie/pam/vplugin"    // init pam plugins
)

// Valkyrie struct containing information and configuration on configured providers and operator
type Valkyrie struct {
	provider *fiber.App
	operator *fiber.App
	config   *configs.ValkyrieConfig
	ctx      context.Context
	cancel   context.CancelFunc
}

// banner ascii art generated using https://textkool.com/en/ascii-art-generator?hl=default&vl=default&font=Bloody&text=Valkyrie
const banner = `
 ██▒   █▓ ▄▄▄       ██▓     ██ ▄█▀▓██   ██▓ ██▀███   ██▓▓█████
▓██░   █▒▒████▄    ▓██▒     ██▄█▒  ▒██  ██▒▓██ ▒ ██▒▓██▒▓█   ▀
 ▓██  █▒░▒██  ▀█▄  ▒██░    ▓███▄░   ▒██ ██░▓██ ░▄█ ▒▒██▒▒███
  ▒██ █░░░██▄▄▄▄██ ▒██░    ▓██ █▄   ░ ▐██▓░▒██▀▀█▄  ░██░▒▓█  ▄
   ▒▀█░   ▓█   ▓██▒░██████▒▒██▒ █▄  ░ ██▒▓░░██▓ ▒██▒░██░░▒████▒
   ░ ▐░   ▒▒   ▓▒█░░ ▒░▓  ░▒ ▒▒ ▓▒   ██▒▒▒ ░ ▒▓ ░▒▓░░▓  ░░ ▒░ ░
   ░ ░░    ▒   ▒▒ ░░ ░ ▒  ░░ ░▒ ▒░ ▓██ ░▒░   ░▒ ░ ▒░ ▒ ░ ░ ░  ░
     ░░    ░   ▒     ░ ░   ░ ░░ ░  ▒ ▒ ░░    ░░   ░  ▒ ░   ░
      ░        ░  ░    ░  ░░  ░    ░ ░        ░      ░     ░  ░
     ░                             ░ ░
`

// NewValkyrie use provided cfg to create a Valkyrie instance
func NewValkyrie(ctx context.Context, cfg *configs.ValkyrieConfig) *Valkyrie {
	// Print banner
	_, _ = fmt.Fprintf(os.Stdout, "%s\n", banner)

	// Define Fiber config
	fiberCfg := fiberConfig(cfg.HTTPServer)

	// Setup context with cancel for shutdown
	cc, cancel := context.WithCancel(ctx)

	// Define a new Fiber app with config
	v := &Valkyrie{
		provider: fiber.New(fiberCfg),
		operator: fiber.New(fiberCfg),
		config:   cfg,
		ctx:      cc,
		cancel:   cancel,
	}

	configureOps(cfg, v)

	// Http client
	httpClient := rest.Create(cfg.HTTPClient)

	// PAM client.
	pamClient, err := pam.GetPamClient(pam.ClientArgs{
		Context: cc,
		Client:  httpClient,
		Config:  cfg.Pam,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting pam client")
	}

	// Provider routes.
	if err = routes.ProviderRoutes(v.provider, cfg.Providers, pamClient); err != nil {
		log.Fatal().Err(err).Msg("Unable to setup the intended provider routes")
	}
	if err = routes.OperatorRoutes(v.operator, cfg.Providers, httpClient); err != nil {
		log.Fatal().Err(err).Msg("Unable to setup the intended operator routes")
	}

	// Swagger
	err = ops.ConfigureSwagger(v.config, v.provider, v.operator)
	if err != nil {
		log.Error().Err(err).Msg("Failed to configure swagger")
	}

	return v
}

func configureOps(cfg *configs.ValkyrieConfig, v *Valkyrie) {
	// Configure logging
	ops.ConfigureLogging(cfg.Logging)

	// Get tracing config
	tracing := ops.Tracing(cfg.Tracing)

	// Middlewares
	fiberMiddleware(v.provider, v.operator)

	// Setup tracing and logging
	ops.TracingMiddleware(tracing, v.provider, v.operator)
	ops.LoggingMiddleware(v.provider, v.operator)

	// Routes
	routes.MonitoringRoutes(v.operator)
}

// Run Starts provider and operator servers. Returns only when
// listeners are ready.
func (v *Valkyrie) Start() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		v.Run(func() { wg.Done() })
	}()
	wg.Wait()
}

// Run starts the server and hangs until it's context gets cancelled. The `ready` callback
// gets fired when the server is ready for accepting connections.
func (v *Valkyrie) Run(ready func()) {
	var wg sync.WaitGroup

	// wait for listeners to start before returning
	wg.Add(2)
	v.operator.Hooks().OnListen(func() error {
		log.Info().Msgf("Operator server listening on '%v'", v.config.HTTPServer.OperatorAddress)
		wg.Done()
		return nil
	})

	v.provider.Hooks().OnListen(func() error {
		log.Info().Msgf("Provider server listening on '%v'", v.config.HTTPServer.ProviderAddress)
		wg.Done()
		return nil
	})

	errs := make(chan error)
	go func() {
		errs <- v.provider.Listen(v.config.HTTPServer.ProviderAddress)
	}()
	go func() {
		errs <- v.operator.Listen(v.config.HTTPServer.OperatorAddress)
	}()

	go func() {
		waitForOr(&wg, 3*time.Second, func() {
			errs <- errors.New("startup timeout")
		})
		ready()
	}()

	v.operator.Hooks().OnShutdown(func() error {
		log.Info().Msgf("Operator server '%v' shutting down", v.config.HTTPServer.OperatorAddress)
		return nil
	})

	v.provider.Hooks().OnShutdown(func() error {
		log.Info().Msgf("Provider server '%v' shutting down", v.config.HTTPServer.ProviderAddress)
		return nil
	})

	select {
	case <-v.ctx.Done():
	case e := <-errs:
		log.Error().Err(e).Msg("listener failed")
		// FIXME: child contexts(like dwh_writer) will not always be given enough time to shutdown
		v.cancel()
	}

	_ = v.provider.Shutdown()
	_ = v.operator.Shutdown()
}

func waitForOr(wg *sync.WaitGroup, dur time.Duration, timeoutFn func()) {
	done := make(chan struct{})
	defer close(done)

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-time.After(dur):
		timeoutFn()
	case <-done:
		return
	}
}

// Stop stops provider and operator servers
func (v *Valkyrie) Stop() {
	v.cancel()
	routine.WaitForFinishWithTimeout(3 * time.Second)
}
