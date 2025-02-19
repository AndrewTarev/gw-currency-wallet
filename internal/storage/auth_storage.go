package storage

import (
	"context"
	"errors"
	"fmt"

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

func (s *Auth) CreateUser(c context.Context, username, email, passwordHash string) error {
	_, err := s.db.Exec(c, "INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)", username, passwordHash, email)
	if err != nil {
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
	return err
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
