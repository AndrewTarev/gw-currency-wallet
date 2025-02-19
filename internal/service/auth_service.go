package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/errs"
	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/utils"
)

type Auth struct {
	stor   *storage.Storage
	logger *logrus.Logger
	jwt    *utils.JWTManager
}

func NewAuthService(stor *storage.Storage, logger *logrus.Logger, jwtManager *utils.JWTManager) *Auth {
	return &Auth{
		stor:   stor,
		logger: logger,
		jwt:    jwtManager,
	}
}

func (a *Auth) Register(c context.Context, user *models.UserRegister) error {
	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	err = a.stor.CreateUser(c, user.Username, user.Email, passwordHash)
	if err != nil {
		return err
	}
	return nil
}

func (a *Auth) Login(c context.Context, userInput *models.UserLogin) (string, error) {
	user, err := a.stor.AuthStorage.GetUserByUsername(c, userInput.Username)
	if err != nil {
		return "", err
	}

	if !utils.CheckPassword(userInput.Password, user.PasswordHash) {
		return "", errs.ErrInvalidPassword
	}

	token, err := a.jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}
