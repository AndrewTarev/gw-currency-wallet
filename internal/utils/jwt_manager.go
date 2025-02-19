package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/storage/models"
)

type JWTManager struct {
	cfg *config.Config
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{cfg: cfg}
}

func (m *JWTManager) GenerateToken(userID uuid.UUID, username string) (string, error) {
	claims := models.Claims{
		UserID:   userID.String(),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.Auth.TokenTTl)), // Срок действия
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.cfg.Auth.SecretKey))
}

// ParseJWT парсит и проверяет токен
func (m *JWTManager) ParseJWT(tokenString string) (*models.Claims, error) {
	secret := []byte(m.cfg.Auth.SecretKey)
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что алгоритм подписи корректен
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	// Извлекаем claims
	claims, ok := token.Claims.(*models.Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
