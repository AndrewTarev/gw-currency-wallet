package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gw-currency-wallet/internal/errs"
)

// ValidationErrorResponse структура для JSON-ответа
type ValidationErrorResponse struct {
	Error struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields,omitempty"` // Поля с ошибками
	} `json:"error"`
}

// ErrorHandler глобальный middleware для обработки ошибок
func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var statusCode int
			var message string
			var fieldErrors map[string]string

			switch {
			case errors.Is(err, errs.ErrUserAlreadyExists):
				statusCode = http.StatusBadRequest
				message = "Username already exists"
				fieldErrors = map[string]string{"username": "field already exists"}
			case errors.Is(err, errs.ErrEmailAlreadyUsed):
				statusCode = http.StatusBadRequest
				message = "Email already used"
				fieldErrors = map[string]string{"email": "field already exists"}
			case errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrInvalidPassword):
				statusCode = http.StatusUnauthorized
				message = errs.ErrInvalidCredentials.Error()
			case errors.Is(err, errs.ErrWalletNotFound):
				statusCode = http.StatusNotFound
				message = "Wallet not found"
			case errors.Is(err, errs.ErrInsufficientFunds):
				statusCode = http.StatusBadRequest
				message = "Insufficient funds"
			case errors.Is(err, errs.ErrInvalidAmount):
				statusCode = http.StatusBadRequest
				message = "Invalid amount, must be greater than zero"
			case errors.Is(err, errs.ErrInvalidUserId):
				statusCode = http.StatusBadRequest
				message = "Invalid user ID"
			case errors.Is(err, errs.ErrUnsupportedCurrency):
				statusCode = http.StatusBadRequest
				message = "Unsupported currency"
			case isGRPCError(err):
				// Проверяем, если ошибка gRPC имеет код NotFound
				st, ok := status.FromError(err)
				if ok && st.Code() == codes.NotFound {
					statusCode = http.StatusNotFound
					message = st.Message() // Используем описание из gRPC ошибки
				} else {
					// Для других типов ошибок, связанных с gRPC
					statusCode = http.StatusBadRequest
					message = err.Error()
				}
			default:
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
				logger.WithFields(logrus.Fields{
					"method":      c.Request.Method + " " + c.Request.URL.Path,
					"error":       err.Error(),
					"stack_trace": string(debug.Stack()),
				}).Error("❌ Unhandled RestAPI error")
			}

			// Формируем JSON-ответ
			errorResponse := ValidationErrorResponse{
				Error: struct {
					Code    int               `json:"code"`
					Message string            `json:"message"`
					Fields  map[string]string `json:"fields,omitempty"`
				}{
					Code:    statusCode,
					Message: message,
					Fields:  fieldErrors,
				},
			}

			c.JSON(statusCode, errorResponse)
		}
	}
}

// isGRPCError проверяет, является ли ошибка gRPC
func isGRPCError(err error) bool {
	// Здесь можно проверять ошибку по типу или содержимому, в зависимости от того, как возвращаются ошибки gRPC
	if strings.Contains(err.Error(), "rpc error") {
		return true
	}
	return false
}
