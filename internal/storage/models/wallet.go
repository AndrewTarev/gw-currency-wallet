package models

import (
	"github.com/shopspring/decimal"
)

type WalletResponse struct {
	BalanceRub decimal.Decimal `json:"balance_rub"`
	BalanceUsd decimal.Decimal `json:"balance_usd"`
	BalanceEur decimal.Decimal `json:"balance_eur"`
}

type WalletTransaction struct {
	Currency string  `json:"currency" validate:"required,len=3,alpha"`
	Amount   float64 `json:"amount" validate:"required,number,gt=0"`
}

type GetBalanceResponse struct {
	Balance WalletResponse `json:"balance"`
}

type WalletOperationsResponse struct {
	Message string         `json:"message"`
	Balance WalletResponse `json:"new_balance"`
}
