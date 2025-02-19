package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

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
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var statusCode int
			var message string
			var fieldErrors map[string]string

			var validationErrs validator.ValidationErrors

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

			// Проверяем, является ли err ошибкой валидации
			case errors.As(err, &validationErrs):
				statusCode = http.StatusBadRequest
				message = "Validation error"
				fieldErrors = make(map[string]string)
				for _, fieldErr := range validationErrs {
					fieldErrors[fieldErr.Field()] = validationErrorMessage(fieldErr)
				}

			default:
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
			}

			// Логируем критические ошибки
			if statusCode == http.StatusInternalServerError {
				// Логируем только неизвестные ошибки
				logger.WithFields(logrus.Fields{
					"method":      c.Request.Method + " " + c.Request.URL.Path,
					"error":       fmt.Sprintf("%v", err),
					"stack_trace": string(debug.Stack()), // Получаем stack trace
				}).Error("❌ Unhandled RestAPI error")
			}

			// Формируем JSON-ответ
			errorResponse := ValidationErrorResponse{}
			errorResponse.Error.Code = statusCode
			errorResponse.Error.Message = message
			if len(fieldErrors) > 0 {
				errorResponse.Error.Fields = fieldErrors
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
		return "must be at least " + fe.Param()
	case "max":
		return "must be at most " + fe.Param()
	case "email":
		return "must be a valid email address"
	default:
		return "is invalid"
	}
}
