package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type UserRegister struct {
	Username string `json:"username" validate:"required,min=5,max=16"`
	Password string `json:"password" validate:"required,min=8,max=16"`
	Email    string `json:"email" validate:"required,email"`
}

func (u *UserRegister) Validate() error {
	return validate.Struct(u)
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserOutput struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}
