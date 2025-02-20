package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gw-currency-wallet/internal/service"
)

type ExchangeHandler struct {
	svc *service.Service
}

func NewExchangeHandler(svc *service.Service) *ExchangeHandler {
	return &ExchangeHandler{
		svc: svc,
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
