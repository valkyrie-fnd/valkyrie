// Package main
package main

import (
	"context"
	"flag"
	"fmt"
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

func main() {
	// Load .env.local if found
	_ = godotenv.Load(".env.local")

	versionFlag := flag.Bool("version", false, "Print the version")

	// Read config location
	configFilePath := flag.String("config",
		"",
		"Path to Valkyrie configuration yaml file")
	flag.Parse()

	if *versionFlag {
		_, _ = fmt.Fprintf(os.Stdout, "Version: %s\n", appVersion)
		return
	}

	cfg, err := configs.Read(configFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
	}
	cfg.Version = appVersion

	mainCtx := listenForSignal()
	v := server.NewValkyrie(mainCtx, cfg)

	v.Run(func() {})
}

func listenForSignal() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Info().Msg("Shutting down")
		cancel()
	}()

	return ctx
}
