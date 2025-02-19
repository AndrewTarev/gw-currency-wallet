package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func ConnectPostgres(dsn string, logger *logrus.Logger) (*pgxpool.Pool, error) {
	// Настраиваем контекст для подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Создаем пул соединений
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Errorf("Error parsing PostgreSQL DSN: %v", err)
		return nil, err
	}

	// Применяем дополнительные настройки пула
	poolConfig.MaxConns = 50                       // Максимальное количество соединений
	poolConfig.MinConns = 5                        // Минимальное количество соединений
	poolConfig.HealthCheckPeriod = 1 * time.Minute // Период проверки соединений

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Errorf("Failed to connect to PostgreSQL: %v", err)
		return nil, err
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		logger.Errorf("PostgreSQL ping failed: %v", err)
		return nil, err
	}

	logger.Debug("Successfully connected to PostgreSQL")
	return pool, nil
}
