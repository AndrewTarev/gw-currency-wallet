package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gw-currency-wallet/internal/utils"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Ожидаем формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ParseJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Сохраняем user_id в контексте запроса
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// GetUserUUID отдает id юзера и превращает в формат uuid
func GetUserUUID(c *gin.Context) (uuid.UUID, error) {
	// Получаем userID из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.UUID{}, fmt.Errorf("user_id is missing in context")
	}

	// Преобразуем userID в UUID
	userUUID, err := uuid.Parse(userID.(string)) // userID приведен к string
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid user_id format")
	}

	return userUUID, nil
}
