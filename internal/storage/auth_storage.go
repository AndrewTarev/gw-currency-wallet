package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/errs"
	"gw-currency-wallet/internal/storage/models"
)

const (
	DuplicateValue = "23505"
)

type Auth struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewAuthStorage(db *pgxpool.Pool, logger *logrus.Logger) *Auth {
	return &Auth{
		db:     db,
		logger: logger,
	}
}

// handlePgError обрабатывает ошибки PostgreSQL
func handlePgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == DuplicateValue {
			if pgErr.ConstraintName == "users_username_key" {
				return errs.ErrUserAlreadyExists
			}
			if pgErr.ConstraintName == "users_email_key" {
				return errs.ErrEmailAlreadyUsed
			}
		}
		return fmt.Errorf("database error: %v", pgErr.Message)
	}
	return err
}

func (s *Auth) CreateUser(c context.Context, username, email, passwordHash string) error {
	tx, err := s.db.Begin(c)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c)

	// Вставляем пользователя
	_, err = tx.Exec(c, "INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)", username, passwordHash, email)
	if err != nil {
		return handlePgError(err)
	}

	// Получаем ID вставленного пользователя
	var userID uuid.UUID
	err = tx.QueryRow(c, "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		return err
	}

	// Создаем кошелек для пользователя
	_, err = tx.Exec(c, "INSERT INTO wallets (user_id) VALUES ($1)", userID)
	if err != nil {
		return err
	}

	// Фиксируем транзакцию
	if err := tx.Commit(c); err != nil {
		return err
	}

	return nil
}

func (s *Auth) GetUserByUsername(c context.Context, username string) (*models.UserOutput, error) {
	var user models.UserOutput

	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username = $1`
	err := s.db.QueryRow(c, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
