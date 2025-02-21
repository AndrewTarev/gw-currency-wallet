package app

import (
	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/delivery/rest"
	"gw-currency-wallet/internal/infrastructure/grpc"
	"gw-currency-wallet/internal/server"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/storage/models/validate"
	"gw-currency-wallet/internal/utils"
	"gw-currency-wallet/pkg/db"
	"gw-currency-wallet/pkg/redis_client"
)

func StartApplication(cfg *config.Config, logger *logrus.Logger) error {
	// Подключение к базе данных
	dbConn, err := db.ConnectPostgres(cfg.Database.Dsn, logger)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	// Создаем зависимости
	cache, err := redis_client.InitRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB) // Кэш
	if err != nil {
		panic(err)
	}
	validator := validate.NewValidator()                            // Общий валидатор входных данных
	exClient := grpc.NewUserServiceClient(cfg.ExchangeService.Addr) // grpc клиент для связи с gw-exchanger
	jwtManager := utils.NewJWTManager(cfg)                          // Генерация и парсинг JWT
	repo := storage.NewStorage(dbConn, logger)
	services := service.NewService(repo, logger, jwtManager, exClient, cache)
	handlers := rest.NewHandler(services, logger, &cfg.Auth, validator)

	// Настройка и запуск сервера
	server.SetupAndRunServer(&cfg.Server, handlers.InitRoutes(logger), logger)
	return nil
}
