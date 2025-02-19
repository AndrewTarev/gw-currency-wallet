package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gw-currency-wallet/internal/utils"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем токен из cookies
		tokenString, err := extractTokenFromCookie(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Парсим и валидируем токен
		claims, err := jwtManager.ParseJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Извлекаем user_id
		userID := claims.UserID

		// Передаем user_id в контекст запроса
		c.Set("user_id", userID)
		c.Next()
	}
}

// extractTokenFromCookie извлекает access_token из cookie
func extractTokenFromCookie(c *gin.Context) (string, error) {
	cookie, err := c.Cookie("token")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(cookie), nil
}
