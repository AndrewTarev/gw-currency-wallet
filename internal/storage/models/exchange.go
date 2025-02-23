package models

import "github.com/shopspring/decimal"

// ExchangeRequest структура запроса на обмен валют
type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" validate:"required,len=3,alpha"`
	ToCurrency   string  `json:"to_currency" validate:"required,len=3,alpha"`
	Amount       float64 `json:"amount" validate:"required,number,gt=0"`
}

type ExchangeRatesResponse struct {
	Rates map[string]string `json:"rates"`
}

type ExchangeCurrencyResponse struct {
	Message         string          `json:"message"`
	ExchangedAmount decimal.Decimal `json:"exchanged_amount"`
	NewBalance      WalletResponse  `json:"new_balance"`
}
