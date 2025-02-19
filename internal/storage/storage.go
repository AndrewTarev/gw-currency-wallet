package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"gw-currency-wallet/internal/storage/models"
)

type AuthStorage interface {
	CreateUser(c context.Context, username, email, passwordHash string) error
	GetUserByUsername(c context.Context, username string) (*models.UserOutput, error)
}

type Storage struct {
	AuthStorage
}

func NewStorage(db *pgxpool.Pool, logger *logrus.Logger) *Storage {
	return &Storage{
		AuthStorage: NewAuthStorage(db, logger),
	}
}
