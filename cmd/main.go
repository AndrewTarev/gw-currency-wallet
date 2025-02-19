package main

import (
	"gw-currency-wallet/internal/app"
	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/pkg/logging"
)

func main() {
	cfg, err := config.LoadConfig("./internal/config")
	if err != nil {
		panic(err)
	}

	logger, err := logging.SetupLogger(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.OutputFile,
		cfg.Logging.KafkaTopic,
		cfg.Logging.KafkaBroker,
	)
	if err != nil {
		panic(err)
	}

	err = app.StartApplication(cfg, logger)
	if err != nil {
		panic(err)
	}
}
