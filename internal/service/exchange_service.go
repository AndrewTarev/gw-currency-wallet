package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/infrastructure/grpc"
)

// Exchange Сервис, работающий с gRPC-клиентом обмена валют
type Exchange struct {
	exClient *grpc.ExchangeClient
	cache    *redis.Client
	logger   *logrus.Logger
}

// NewExchangeService Конструктор
func NewExchangeService(
	exClient *grpc.ExchangeClient,
	cache *redis.Client,
	logger *logrus.Logger,
) *Exchange {
	return &Exchange{
		exClient: exClient,
		cache:    cache,
		logger:   logger,
	}
}

// GetRates Метод получения курсов обмена
func (e *Exchange) GetRates(ctx context.Context) (map[string]string, error) {
	cacheKey := "exchange_rates"

	// Проверяем кэш Redis
	data, err := e.cache.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var cachedRates map[string]string
		if json.Unmarshal(data, &cachedRates) == nil {
			e.logger.Debug("✅ Курсы валют получены из кэша Redis")
			return cachedRates, nil
		}
	}

	// Данных нет в кэше — делаем запрос в сервис
	rates, err := e.exClient.GetExchangeRates(ctx)
	e.logger.Debug("❌ Курсы валют получены НЕ из кэша Redis")
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш на 5 минут
	data, _ = json.Marshal(rates.Rates)
	e.cache.Set(ctx, cacheKey, data, 5*time.Minute)

	return rates.Rates, nil
}

// GetRate Метод получения курса для конкретной валютной пары
func (e *Exchange) GetRate(ctx context.Context, fromCurrency, toCurrency string) (string, error) {
	cacheKey := fmt.Sprintf("exchange_rate:%s:%s", fromCurrency, toCurrency)

	// Проверяем кэш Redis
	rate, err := e.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		e.logger.Debug("✅ Курс валюты получен из кэша Redis")
		return rate, nil
	}

	// Данных нет в кэше — делаем запрос в сервис
	rateResponse, err := e.exClient.GetExchangeRateForCurrency(ctx, fromCurrency, toCurrency)
	e.logger.Debug("❌ Курс валюты получены НЕ из кэша Redis")
	if err != nil {
		return "", err
	}

	// Сохраняем в кэш на 5 минут
	e.cache.Set(ctx, cacheKey, rateResponse.Rate, 5*time.Minute)

	return rateResponse.Rate, nil
}
