package grpc

import (
	"context"
	"log"

	exchange "github.com/AndrewTarev/proto-repo/gen/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ExchangeClient struct {
	client exchange.ExchangeServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient(grpcAddr string) *ExchangeClient {
	// creds := credentials.NewTLS(&tls.Config{
	// 	InsecureSkipVerify: true,
	// })

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials())) // TODO Для продакш, подставь grpc.WithTransportCredentials(creds)
	if err != nil {
		log.Fatalf("Failed to connect to UserService: %v", err)

	}

	client := exchange.NewExchangeServiceClient(conn)
	return &ExchangeClient{client: client, conn: conn}
}

// GetExchangeRates Получить все курсы обмена валют
func (e *ExchangeClient) GetExchangeRates(c context.Context) (*exchange.ExchangeRatesResponse, error) {
	resp, err := e.client.GetExchangeRates(c, &exchange.Empty{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetExchangeRateForCurrency Получить курс обмена для конкретной валюты
func (e *ExchangeClient) GetExchangeRateForCurrency(c context.Context, fromCurrency, toCurrency string) (*exchange.ExchangeRateResponse, error) {
	req := &exchange.CurrencyRequest{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
	}

	resp, err := e.client.GetExchangeRateForCurrency(c, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
