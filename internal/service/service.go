package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/infrastructure/grpc"
	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/utils"
)

type AuthService interface {
	Register(c context.Context, input models.UserRegister) error
	Login(c context.Context, userInput *models.UserLogin) (string, error)
}

type ExchangeService interface {
	GetRates(c context.Context) (map[string]string, error)
	GetRate(c context.Context, fromCurrency, toCurrency string) (string, error)
	ExchangeCurrency(c context.Context, userID uuid.UUID, fromCurrency string, toCurrency string, amount decimal.Decimal, exchangedAmount decimal.Decimal) (models.WalletResponse, error)
}

type WalletService interface {
	GetBalance(c context.Context, userID uuid.UUID) (models.WalletResponse, error)
	Deposit(c context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (models.WalletResponse, error)
	Withdraw(c context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (models.WalletResponse, error)
}

type Service struct {
	AuthService
	ExchangeService
	WalletService
}

func NewService(
	stor *storage.Storage,
	logger *logrus.Logger,
	jwtManager *utils.JWTManager,
	exClient *grpc.ExchangeClient,
	cache *redis.Client,
) *Service {
	return &Service{
		AuthService:     NewAuthService(stor, logger, jwtManager),
		ExchangeService: NewExchangeService(exClient, cache, logger, stor),
		WalletService:   NewWalletService(stor, logger),
	}
}
