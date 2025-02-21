package errs

import "github.com/pkg/errors"

// auth
var (
	ErrUserAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyUsed  = errors.New("email already used")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserNotFound      = errors.New("user not found")
)

// wallets
var (
	ErrWalletNotFound = errors.New("wallet not found")
	ErrInvalidUserId  = errors.New("invalid user id")
)
