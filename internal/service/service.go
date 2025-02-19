package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/utils"
)

type AuthService interface {
	Register(c context.Context, input *models.UserRegister) error
	Login(c context.Context, userInput *models.UserLogin) (string, error)
}

type Service struct {
	AuthService
}

func NewService(stor *storage.Storage, logger *logrus.Logger, jwtManager *utils.JWTManager) *Service {
	return &Service{
		AuthService: NewAuthService(stor, logger, jwtManager),
	}
}
