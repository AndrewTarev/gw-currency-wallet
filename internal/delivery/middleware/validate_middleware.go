package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gw-currency-wallet/internal/storage/models/validate"
)

// ValidationMiddleware проверяет входные данные перед выполнением хендлера
func ValidationMiddleware[T any](v *validate.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input T

		// Парсим JSON в структуру
		if err := c.ShouldBindJSON(&input); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, ValidationErrorResponse{
				Error: struct {
					Code    int               `json:"code"`
					Message string            `json:"message"`
					Fields  map[string]string `json:"fields,omitempty"`
				}{
					Code:    http.StatusBadRequest,
					Message: "Invalid request format",
				},
			})
			return
		}

		// Валидируем структуру
		if err := v.ValidateStruct(input); err != nil {
			validationErrors := make(map[string]string)
			for _, fieldErr := range err.(validator.ValidationErrors) {
				validationErrors[fieldErr.Field()] = validationErrorMessage(fieldErr)
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, ValidationErrorResponse{
				Error: struct {
					Code    int               `json:"code"`
					Message string            `json:"message"`
					Fields  map[string]string `json:"fields,omitempty"`
				}{
					Code:    http.StatusBadRequest,
					Message: "Validation failed",
					Fields:  validationErrors,
				},
			})
			return
		}

		// Передаём данные дальше
		c.Set("validatedInput", input)
		c.Next()
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
