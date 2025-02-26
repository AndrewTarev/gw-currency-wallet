package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"gw-currency-wallet/internal/delivery/middleware"
	"gw-currency-wallet/internal/errs"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/storage/models/validate"
)

type ExchangeHandler struct {
	svc      *service.Service
	validate *validate.Validator
}

func NewExchangeHandler(svc *service.Service, validate *validate.Validator) *ExchangeHandler {
	return &ExchangeHandler{
		svc:      svc,
		validate: validate,
	}
}

// GetExchangeRates godoc
// @Summary Получить текущие курсы валют
// @Description Возвращает список текущих курсов обмена валют
// @Tags exchange
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ExchangeRatesResponse
// @Failure 500 {object} middleware.ValidationErrorResponse
// @Router /exchange/rates [get]
func (h *ExchangeHandler) GetExchangeRates(c *gin.Context) {
	rates, err := h.svc.GetRates(c)
	if err != nil {
		c.Error(err)
		return
	}

	successResponse := models.ExchangeRatesResponse{
		Rates: rates,
	}

	c.JSON(http.StatusOK, successResponse)
}

// ExchangeCurrency godoc
// @Summary Обмен валют
// @Description Обмен валюты с использованием заданного количества и курсов валют
// @Tags exchange
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.ExchangeRequest true "Данные для обмена валюты"
// @Success 200 {object} models.ExchangeCurrencyResponse
// @Failure 400 {object} middleware.ValidationErrorResponse
// @Failure 500 {object} middleware.ValidationErrorResponse
// @Router /exchange [post]
func (h *ExchangeHandler) ExchangeCurrency(c *gin.Context) {
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		c.Error(err)
	}

	input, exists := c.Get("validatedInput")
	if !exists {
		c.Error(errs.ErrValidationNotWorking)
		return
	}

	userInput := input.(models.ExchangeRequest)

	// Валидируем входные данные
	if err := h.validate.ValidateStruct(&userInput); err != nil {
		c.Error(err)
		return
	}

	// Преобразуем float64 в decimal.Decimal
	amountDecimal := decimal.NewFromFloat(userInput.Amount)
	amountDecimal = amountDecimal.Round(2)

	rateStr, err := h.svc.GetRate(c, userInput.FromCurrency, userInput.ToCurrency)
	if err != nil {
		c.Error(err)
		return
	}

	// Преобразуем курс из строки в decimal.Decimal
	rate, err := decimal.NewFromString(rateStr)
	if err != nil {
		c.Error(err)
		return
	}

	exchangedAmount := amountDecimal.Mul(rate)

	newBalance, err := h.svc.ExchangeService.ExchangeCurrency(c, userID, userInput.FromCurrency, userInput.ToCurrency, amountDecimal, exchangedAmount)
	if err != nil {
		c.Error(err)
		return
	}

	successResponse := models.ExchangeCurrencyResponse{
		Message:         "Exchange successful",
		ExchangedAmount: exchangedAmount,
		NewBalance:      newBalance,
	}

	c.JSON(http.StatusOK, successResponse)
}
