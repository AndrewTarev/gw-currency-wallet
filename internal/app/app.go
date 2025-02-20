package app

import (
	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/delivery/rest"
	"gw-currency-wallet/internal/infrastructure/grpc"
	"gw-currency-wallet/internal/server"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/utils"
	"gw-currency-wallet/pkg/db"
)

func StartApplication(cfg *config.Config, logger *logrus.Logger) error {
	// Подключение к базе данных
	dbConn, err := db.ConnectPostgres(cfg.Database.Dsn, logger)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	// Создаем зависимости
	exClient := grpc.NewUserServiceClient(cfg.ExchangeService.Addr)
	jwtManager := utils.NewJWTManager(cfg)
	repo := storage.NewStorage(dbConn, logger)
	services := service.NewService(repo, logger, jwtManager, exClient)
	handlers := rest.NewHandler(services, logger, &cfg.Auth)

	// Настройка и запуск сервера
	server.SetupAndRunServer(&cfg.Server, handlers.InitRoutes(logger), logger)
	return nil
}
