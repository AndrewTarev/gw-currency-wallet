package errs

import "github.com/pkg/errors"

// auth
var (
	ErrUserAlreadyExists  = errors.New("username already exists")
	ErrEmailAlreadyUsed   = errors.New("email already used")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// wallets
var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInvalidUserId       = errors.New("invalid user id")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrInvalidAmount       = errors.New("invalid amount, must be greater than zero")
	ErrUnsupportedCurrency = errors.New("unsupported currency")
)

var (
	ErrValidationNotWorking = errors.New("Validation middleware not working")
)
