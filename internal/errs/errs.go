package errs

import "github.com/pkg/errors"

var (
	ErrUserAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyUsed  = errors.New("email already used")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserNotFound      = errors.New("user not found")
)
