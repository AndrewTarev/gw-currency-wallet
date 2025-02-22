package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/errs"
	"gw-currency-wallet/internal/storage/models"
)

type Wallet struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewWalletStorage(db *pgxpool.Pool, logger *logrus.Logger) *Wallet {
	return &Wallet{
		db:     db,
		logger: logger,
	}
}

// GetBalance извлекает информацию о балансе пользователя
func (w *Wallet) GetBalance(c context.Context, userID uuid.UUID) (models.WalletResponse, error) {
	// SQL запрос для получения данных баланса по userID
	query := `
		SELECT balance_rub, balance_usd, balance_eur
		FROM wallets
		WHERE user_id = $1
		LIMIT 1
	`

	// Выполняем запрос и извлекаем данные в структуру
	var response models.WalletResponse
	err := w.db.QueryRow(c, query, userID).Scan(
		&response.BalanceRub,
		&response.BalanceUsd,
		&response.BalanceEur,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Если не нашли кошелек для пользователя, возвращаем ошибку
			return models.WalletResponse{}, errs.ErrWalletNotFound
		}
		return models.WalletResponse{}, err
	}

	return response, nil
}

// Deposit Пополнение баланса и возврат нового состояния кошелька
func (w *Wallet) Deposit(ctx context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (models.WalletResponse, error) {
	query := fmt.Sprintf(
		`UPDATE wallets 
		SET balance_%s = balance_%s + $1 
		WHERE user_id = $2 
		RETURNING balance_rub, balance_usd, balance_eur`,
		strings.ToLower(currency), strings.ToLower(currency),
	)

	var response models.WalletResponse
	err := w.db.QueryRow(ctx, query, amount, userID).Scan(&response.BalanceRub, &response.BalanceUsd, &response.BalanceEur)
	if err != nil {
		return models.WalletResponse{}, err
	}
	return response, nil
}

// Withdraw Списание средств и возврат нового состояния кошелька
func (w *Wallet) Withdraw(ctx context.Context, userID uuid.UUID, currency string, amount decimal.Decimal) (models.WalletResponse, error) {
	query := fmt.Sprintf(
		`UPDATE wallets 
		SET balance_%s = balance_%s - $1 
		WHERE user_id = $2 AND balance_%s >= $1 
		RETURNING balance_rub, balance_usd, balance_eur`,
		strings.ToLower(currency), strings.ToLower(currency), strings.ToLower(currency),
	)

	var response models.WalletResponse
	err := w.db.QueryRow(ctx, query, amount, userID).Scan(&response.BalanceRub, &response.BalanceUsd, &response.BalanceEur)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.WalletResponse{}, errs.ErrInsufficientFunds
		}
		return models.WalletResponse{}, err
	}
	return response, nil
}
