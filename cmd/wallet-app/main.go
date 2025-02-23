package main

import (
	"gw-currency-wallet/internal/app"
	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/pkg/logging"
)

// @title My API
// @version 1.0
// @description This is an example API that demonstrates Swagger documentation integration.
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {your_token}
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
