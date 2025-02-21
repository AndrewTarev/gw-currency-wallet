package models

// ExchangeRequest структура запроса на обмен валют
type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" validate:"required,len=3,alpha"`
	ToCurrency   string  `json:"to_currency" validate:"required,len=3,alpha"`
	Amount       float64 `json:"amount" validate:"required,number,gt=0"`
}
