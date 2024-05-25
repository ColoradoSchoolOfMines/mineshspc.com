package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	globallog "github.com/rs/zerolog/log"
	"go.mau.fi/util/dbutil"
	"go.mau.fi/util/exerrors"
	"go.mau.fi/util/exzerolog"
	"gopkg.in/yaml.v3"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/database"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal"
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/config"
)

func main() {
	var configFilenames config.ConfigFilenames
	flag.Var(&configFilenames, "config", "config file name")
	flag.Parse()
	if len(configFilenames) == 0 {
		configFilenames = append(configFilenames, "config.yaml")
	}

	// Parse the configuration
	var config config.Configuration
	for _, filename := range configFilenames {
		f := exerrors.Must(os.Open(filename))
		exerrors.PanicIfNotNil(yaml.NewDecoder(f).Decode(&config))
		exerrors.PanicIfNotNil(f.Close())
	}

	// Setup logging
	log := exerrors.Must(config.Logging.Compile())
	defaultCtxLog := log.With().Bool("default_context_log", true).Caller().Logger()
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.CallerMarshalFunc = exzerolog.CallerWithFunctionName
	zerolog.DefaultContextLogger = &defaultCtxLog
	globallog.Logger = log.With().Bool("global_log", true).Caller().Logger()

	log.Info().Msg("mineshspc.com backend starting...")

	// Open the database
	rawDB := exerrors.Must(dbutil.NewFromConfig("mineshspc", config.Database, dbutil.ZeroLogger(*log)))
	db := database.NewDatabase(rawDB)
	if err := db.DB.Upgrade(context.TODO()); err != nil {
		log.Fatal().Err(err).Msg("failed to upgrade the mineshspc.com database")
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
			db.DB.RawDB.Close()
			os.Exit(0)
		}
	}()

	// Healthcheck loop
	healthcheckTimer := time.NewTimer(time.Second)
	healthcheckURL := config.HealthcheckURL
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
					log.Err(err).Msg("Failed to send healtheck ping")
				} else if resp.StatusCode < 200 || 300 <= resp.StatusCode {
					log.Error().Int("status", resp.StatusCode).Msg("non-200 status code from healthcheck ping")
				}
				healthcheckTimer.Reset(30 * time.Second)
			}
		}
	}(log)

	internal.NewApplication(log, config, db).Start()
}
