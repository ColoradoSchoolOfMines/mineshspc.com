package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Arg parsing
	configPath := flag.String("config", "./config.yaml", "config file location")
	dbPath := flag.String("db", "./mineshspc.db", "SQLite database file location")
	flag.Parse()

	// Configure logging
	log := log.Output(os.Stdout)
	if os.Getenv("LOG_CONSOLE") != "" {
		log = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	log.Info().Msg("mineshspc.com backend starting...")

	// Open the database
	log.Info().Str("db_path", *dbPath).Msg("opening database...")
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	// Make sure to exit cleanly
	c := make(chan os.Signal, 1)
	healthcheckCtx, healthcheckCancel := context.WithCancel(context.Background())
	signal.Notify(c,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	go func() {
		for range c { // when the process is killed
			log.Info().Msg("Cleaning up")
			healthcheckCancel()
			db.Close()
			os.Exit(0)
		}
	}()

	app := NewApplication(&log, db)

	// Load configuration
	log.Info().Str("config_path", *configPath).Msg("Reading config")
	configYaml, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
	}
	if err := app.Config.Parse(configYaml); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}

	// Healthcheck loop
	healthcheckTimer := time.NewTimer(time.Second)
	healthcheckURL := app.Config.HealthcheckURL
	go func(log *zerolog.Logger) {
		if healthcheckURL == "" {
			log.Warn().Msg("Healthcheck URL not set, skipping healthcheck")
			return
		}
		for {
			select {
			case <-healthcheckCtx.Done():
				return
			case <-healthcheckTimer.C:
				log.Info().Msg("Sending healthcheck ping")
				resp, err := http.Get(healthcheckURL)
				if err != nil {
					log.Error().Err(err).Msg("Failed to send healtheck ping")
				} else if resp.StatusCode < 200 || 300 <= resp.StatusCode {
					log.Error().Int("status", resp.StatusCode).Msg("non-200 status code from healthcheck ping")
				}
				healthcheckTimer.Reset(30 * time.Second)
			}
		}
	}(&log)

	app.Start()
}
