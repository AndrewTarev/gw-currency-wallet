package validate

import "github.com/go-playground/validator/v10"

// Validator — общий валидатор, который можно переиспользовать
type Validator struct {
	validate *validator.Validate
}

// NewValidator создает и настраивает валидатор
func NewValidator() *Validator {
	v := validator.New()

	// Добавляем кастомные правила, если нужно
	// v.RegisterValidation("currency", ValidateCurrency)

	return &Validator{validate: v}
}

// ValidateStruct выполняет валидацию структуры
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}
