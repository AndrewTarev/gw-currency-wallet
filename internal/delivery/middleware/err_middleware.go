package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

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
		c.Next() // Выполняем все обработчики

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var statusCode int
			var message string
			var fieldErrors map[string]string

			// Если ошибка — это валидация, обрабатываем ее отдельно
			var validationErrs validator.ValidationErrors
			if errors.As(err, &validationErrs) {
				statusCode = http.StatusBadRequest
				message = "Validation error"
				fieldErrors = make(map[string]string)

				for _, fieldErr := range validationErrs {
					fieldErrors[fieldErr.Field()] = validationErrorMessage(fieldErr)
				}
			} else {
				// Обрабатываем кастомные ошибки
				switch {
				case errors.Is(err, errs.ErrUserAlreadyExists):
					statusCode = http.StatusBadRequest
					message = "username already exists"
					fieldErrors = map[string]string{"username": "field already exists"}
				case errors.Is(err, errs.ErrEmailAlreadyUsed):
					statusCode = http.StatusBadRequest
					message = "email already used"
					fieldErrors = map[string]string{"email": "field already exists"}
				case errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrInvalidPassword):
					statusCode = http.StatusUnauthorized
					message = "Invalid username or password"
				case errors.Is(err, errs.ErrWalletNotFound):
					statusCode = http.StatusNotFound
					message = "Wallet not found"
				case errors.Is(err, errs.ErrInsufficientFunds):
					statusCode = http.StatusBadRequest
					message = "Insufficient funds"
				case errors.Is(err, errs.ErrInvalidAmount):
					statusCode = http.StatusBadRequest
					message = "invalid amount, must be greater than zero"
				case errors.Is(err, errs.ErrInvalidUserId):
					statusCode = http.StatusBadRequest
					message = "invalid user id"
				case errors.Is(err, errs.ErrUnsupportedCurrency):
					statusCode = http.StatusBadRequest
					message = "unsupported currency"

				// Обработка ошибок от gRPC
				case isGRPCError(err):
					statusCode = http.StatusBadRequest
					message = err.Error()
				default:
					statusCode = http.StatusInternalServerError
					message = "Internal server error"
				}
			}

			// Логируем критические ошибки (500)
			if statusCode == http.StatusInternalServerError {
				logger.WithFields(logrus.Fields{
					"method":      c.Request.Method + " " + c.Request.URL.Path,
					"error":       err.Error(),
					"stack_trace": string(debug.Stack()), // Stack trace для дебага
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

			// Отправляем JSON-ответ
			c.JSON(statusCode, errorResponse)
		}
	}
}

// validationErrorMessage формирует читаемое сообщение ошибки
func validationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "email":
		return "must be a valid email address"
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	default:
		return "is invalid"
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
