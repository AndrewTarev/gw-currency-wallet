package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

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

func (h *ExchangeHandler) GetExchangeRates(c *gin.Context) {
	rates, err := h.svc.GetRates(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"rates": rates})
}

func (h *ExchangeHandler) ExchangeCurrency(c *gin.Context) {
	var input models.ExchangeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	// Валидируем входные данные
	if err := h.validate.ValidateStruct(input); err != nil {
		c.Error(err)
		return
	}

	// Преобразуем float64 в decimal.Decimal
	amountDecimal := decimal.NewFromFloat(input.Amount)
	amountDecimal = amountDecimal.Round(2)

	rate, err := h.svc.GetRate(c, input.FromCurrency, input.ToCurrency)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rate": rate})
}
