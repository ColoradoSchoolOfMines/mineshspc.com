package main

import (
	"flag"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Arg parsing
	configPath := flag.String("config", "./config.yaml", "config file location")
	logLevelStr := flag.String("loglevel", "debug", "the log level")
	logFilename := flag.String("logfile", "", "the log file to use (defaults to '' meaning no log file)")
	flag.Parse()

	// Configure logging
	if *logFilename != "" {
		logFile, err := os.OpenFile(*logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			mw := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(mw)
		} else {
			log.Errorf("Failed to open logging file; using default stderr: %s", err)
		}
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	logLevel, err := log.ParseLevel(*logLevelStr)
	if err == nil {
		log.SetLevel(logLevel)
	} else {
		log.Errorf("Invalid loglevel '%s'. Using default 'debug'.", logLevel)
	}

	log.Info("mineshspc.com backend starting...")

	app := NewApplication()

	// Load configuration
	log.Infof("Reading config from %s...", *configPath)
	configYaml, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Could not read config from %s: %s", *configPath, err)
	}
	if err := app.Configuration.Parse(configYaml); err != nil {
		log.Fatal("Failed to read config!")
	}

	app.Authenticate()
	app.RegisterHandlers()
	app.ConfigureSheets()

	http.ListenAndServe(":8090", nil)
}
