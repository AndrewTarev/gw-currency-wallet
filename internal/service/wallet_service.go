package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

func (w *Wallet) GetBalance(c context.Context, userID uuid.UUID) (models.WalletResponse, error) {
	wallet, err := w.stor.WalletStorage.GetBalance(c, userID)
	if err != nil {
		return models.WalletResponse{}, err
	}

	return wallet, nil
}

// Deposit – пополнение баланса
func (w *Wallet) Deposit(c context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (models.WalletResponse, error) {
	// Проверим, что сумма больше нуля
	if amount.IsNegative() || amount.IsZero() {
		return models.WalletResponse{}, errs.ErrInvalidAmount
	}

	balance, err := w.stor.WalletStorage.Deposit(c, userID, currency, amount)
	if err != nil {
		return models.WalletResponse{}, err
	}
	w.logger.Debugf("Deposit succeeded")
	return balance, nil
}

// Withdraw – создаем Kafka-событие на списание
func (w *Wallet) Withdraw(c context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (models.WalletResponse, error) {
	// Проверим, что сумма больше нуля
	if amount.IsNegative() || amount.IsZero() {
		return models.WalletResponse{}, errs.ErrInvalidAmount
	}

	// Пытаемся снять средства
	balance, err := w.stor.WalletStorage.Withdraw(c, userID, currency, amount)
	if err != nil {
		return models.WalletResponse{}, err
	}

	w.logger.Debugf("Successfully withdrew %s %s from user %v", amount, currency, userID)
	return balance, nil
}
