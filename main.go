// Package main
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/server"

	_ "github.com/joho/godotenv/autoload" // load .env file automatically
)

var appVersion = "devel"

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

func main() {
	os.Exit(mainReal(listenForSignal(), os.Stdout))
}
func mainReal(ctx context.Context, out io.Writer) int {

	// Load .env.local if found
	_ = godotenv.Load(".env.local")

	versionFlag := flag.Bool("version", false, "Print the version")

	// Read config location
	configFilePath := flag.String("config",
		"",
		"Path to Valkyrie configuration yaml file")
	flag.Parse()

	if *versionFlag {
		_, _ = fmt.Fprintf(out, "Version: %s\n", appVersion)
		return 0
	}

	cfg, err := configs.Read(configFilePath)
	if err != nil {
		log.Err(err).Msg("Failed to read config")
		return 1
	}
	cfg.Version = appVersion
	// Print banner
	_, _ = fmt.Fprintf(out, "%s\n", banner)
	v, err := server.NewValkyrie(ctx, cfg)
	if err != nil {
		return 1
	}

	v.Run(func() {})
	return 0
}

func listenForSignal() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down")
		cancel()
	}()

	return ctx
}
