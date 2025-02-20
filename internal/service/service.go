package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/infrastructure/grpc"
	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/utils"
)

type AuthService interface {
	Register(c context.Context, input *models.UserRegister) error
	Login(c context.Context, userInput *models.UserLogin) (string, error)
}

type ExchangeService interface {
	GetRates(c context.Context) (map[string]string, error)
	GetRate(c context.Context, fromCurrency, toCurrency string) (string, error)
}

type Service struct {
	AuthService
	ExchangeService
}

func NewService(
	stor *storage.Storage,
	logger *logrus.Logger,
	jwtManager *utils.JWTManager,
	exClient *grpc.ExchangeClient,
) *Service {
	return &Service{
		AuthService:     NewAuthService(stor, logger, jwtManager),
		ExchangeService: NewExchangeService(exClient),
	}
}
