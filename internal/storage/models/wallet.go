package models

import (
	"github.com/shopspring/decimal"
)

type WalletResponse struct {
	BalanceRub decimal.Decimal `json:"balance_rub"`
	BalanceUsd decimal.Decimal `json:"balance_usd"`
	BalanceEur decimal.Decimal `json:"balance_eur"`
}
