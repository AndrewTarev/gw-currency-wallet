package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/infrastructure/grpc"
	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/storage/models"
)

// Exchange Сервис, работающий с gRPC-клиентом обмена валют
type Exchange struct {
	exClient *grpc.ExchangeClient
	cache    *redis.Client
	logger   *logrus.Logger
	stor     *storage.Storage
}

// NewExchangeService Конструктор
func NewExchangeService(
	exClient *grpc.ExchangeClient,
	cache *redis.Client,
	logger *logrus.Logger,
	stor *storage.Storage,
) *Exchange {
	return &Exchange{
		exClient: exClient,
		cache:    cache,
		logger:   logger,
		stor:     stor,
	}
}

// GetRates Метод получения курсов обмена
func (e *Exchange) GetRates(c context.Context) (map[string]string, error) {
	cacheKey := "exchange_rates"

	// Проверяем кэш Redis
	data, err := e.cache.Get(c, cacheKey).Bytes()
	if err == nil {
		var cachedRates map[string]string
		if json.Unmarshal(data, &cachedRates) == nil {
			e.logger.Debug("✅ Курсы валют получены из кэша Redis")
			return cachedRates, nil
		}
	}

	// Данных нет в кэше — делаем запрос в сервис
	rates, err := e.exClient.GetExchangeRates(c)
	e.logger.Debug("❌ Курсы валют получены НЕ из кэша Redis")
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш на 5 минут
	data, _ = json.Marshal(rates.Rates)
	e.cache.Set(c, cacheKey, data, 5*time.Minute)
	return rates.Rates, nil
}

// GetRate Метод получения курса для конкретной валютной пары
func (e *Exchange) GetRate(c context.Context, fromCurrency, toCurrency string) (string, error) {
	cacheKey := fmt.Sprintf("exchange_rate:%s:%s", fromCurrency, toCurrency)

	// Проверяем кэш Redis
	rate, err := e.cache.Get(c, cacheKey).Result()
	if err == nil {
		e.logger.Debug("✅ Курс валюты получен из кэша Redis")
		return rate, nil
	}

	// Данных нет в кэше — делаем запрос в сервис
	rateResponse, err := e.exClient.GetExchangeRateForCurrency(c, fromCurrency, toCurrency)
	e.logger.Debug("❌ Курс валюты получены НЕ из кэша Redis")
	if err != nil {
		return "", err
	}

	// Сохраняем в кэш на 5 минут
	e.cache.Set(c, cacheKey, rateResponse.Rate, 5*time.Minute)

	return rateResponse.Rate, nil
}

// ExchangeCurrency обмен валют
func (e *Exchange) ExchangeCurrency(
	c context.Context,
	userID uuid.UUID,
	fromCurrency string,
	toCurrency string,
	amount decimal.Decimal,
	exchangedAmount decimal.Decimal,
) (models.WalletResponse, error) {
	balance, err := e.stor.WalletStorage.Exchange(c, userID, fromCurrency, toCurrency, amount, exchangedAmount)
	if err != nil {
		return models.WalletResponse{}, err
	}
	return balance, nil
}
