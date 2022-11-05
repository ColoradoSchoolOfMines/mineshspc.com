package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Arg parsing
	configPath := flag.String("config", "./config.yaml", "config file location")
	flag.Parse()

	// Configure logging
	log := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("mineshspc.com backend starting...")

	app := NewApplication(&log)

	// Load configuration
	log.Info().Str("config_path", *configPath).Msg("Reading config")
	configYaml, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
	}
	if err := app.Configuration.Parse(configYaml); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}

	app.Start()
}
