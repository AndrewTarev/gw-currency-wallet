package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RecoverMiddleware — middleware для обработки паник
func RecoverMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Логируем панику
				logger.WithFields(logrus.Fields{
					"method":      c.Request.Method + " " + c.Request.URL.Path,
					"panic":       fmt.Sprintf("%v", r),
					"stack_trace": string(debug.Stack()), // Stack trace паники
				}).Fatal("🔥 Panic recovered")

				// Возвращаем клиенту 500 Internal Server Error
				c.JSON(http.StatusInternalServerError, ValidationErrorResponse{
					Error: struct {
						Code    int               `json:"code"`
						Message string            `json:"message"`
						Fields  map[string]string `json:"fields,omitempty"`
					}{
						Code:    http.StatusInternalServerError,
						Message: "Internal server error",
					},
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
