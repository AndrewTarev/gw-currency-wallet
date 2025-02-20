package service

import (
	"context"

	"gw-currency-wallet/internal/infrastructure/grpc"
)

// Exchange Сервис, работающий с gRPC-клиентом обмена валют
type Exchange struct {
	exClient *grpc.ExchangeClient
}

// NewExchangeService Конструктор
func NewExchangeService(exClient *grpc.ExchangeClient) *Exchange {
	return &Exchange{exClient: exClient}
}

// GetRates Метод получения курсов обмена
func (e *Exchange) GetRates(ctx context.Context) (map[string]string, error) {
	rates, err := e.exClient.GetExchangeRates(ctx)
	if err != nil {
		return nil, err
	}
	return rates.Rates, nil
}

// GetRate Метод получения курса для конкретной валютной пары
func (e *Exchange) GetRate(ctx context.Context, fromCurrency, toCurrency string) (string, error) {
	rate, err := e.exClient.GetExchangeRateForCurrency(ctx, fromCurrency, toCurrency)
	if err != nil {
		return "", err
	}
	return rate.Rate, nil
}
