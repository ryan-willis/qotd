package startup

import (
	"os"

	"github.com/ryan-willis/qotd/app"
	"github.com/ryan-willis/qotd/app/logger"
	"github.com/ryan-willis/qotd/app/storage"
)

func LoadContext() *app.Context {
	_, isProduction := os.LookupEnv("IS_PROD") // why tho
	log := logger.Get(isProduction)

	log.Info("Connecting to cache...")
	storageProvider, cacheErr := storage.Connect()
	if cacheErr != nil {
		log.Error("error while connecting to cache", logger.Field("err", cacheErr.Error()))
		os.Exit(1)
	}

	log.Info("Connected to " + storageProvider.Name() + " cache.")

	log.Info("Checking for Sentry...")
	sentryDSN, hasSentry := os.LookupEnv("SENTRY_DSN")
	if hasSentry {
		log.Info("Sentry is enabled.")
	} else {
		log.Warn("Sentry is disabled.")
	}

	return &app.Context{
		Log:          log,
		Storage:      storageProvider,
		IsProduction: isProduction,
		SentryDSN:    sentryDSN,
	}
}
