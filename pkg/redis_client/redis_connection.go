package redis_client

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	logger "github.com/sirupsen/logrus"
)

// InitRedisClient инициализирует и возвращает клиента Redis
func InitRedisClient(addr, password string, db int) (*redis.Client, error) {
	// Создаем новый клиент Redis
	client := redis.NewClient(&redis.Options{
		Addr:     addr,     // Адрес Redis сервера
		Password: password, // Пароль, если установлен
		DB:       db,       // Номер базы данных
		PoolSize: 10,
	})

	// Проверяем соединение с Redis
	err := client.Ping(context.Background()).Err()
	if err != nil {
		logger.Errorf("redis client init err: %v", err)
		return nil, fmt.Errorf("could not connect to Redis: %w", err)
	}
	logger.Debugf("Successfully connected to Redis at %s", addr)
	return client, nil
}
