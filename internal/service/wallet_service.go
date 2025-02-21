package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/errs"
	"gw-currency-wallet/internal/storage"
	"gw-currency-wallet/internal/storage/models"
)

type Wallet struct {
	stor   *storage.Storage
	logger *logrus.Logger
}

func NewWalletService(stor *storage.Storage, logger *logrus.Logger) *Wallet {
	return &Wallet{
		stor:   stor,
		logger: logger,
	}
}

func (w *Wallet) GetBalance(c context.Context, userID string) (models.WalletResponse, error) {
	// Преобразуем userID в UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return models.WalletResponse{}, errs.ErrInvalidUserId
	}

	wallet, err := w.stor.WalletStorage.GetBalance(c, userUUID)
	if err != nil {
		return models.WalletResponse{}, err
	}

	return wallet, nil
}
