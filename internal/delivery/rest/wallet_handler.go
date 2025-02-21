package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gw-currency-wallet/internal/service"
)

type Wallet struct {
	svc *service.Service
}

func NewWalletHandler(svc *service.Service) *Wallet {
	return &Wallet{svc: svc}
}

func (w *Wallet) GetBalance(c *gin.Context) {
	// // Получаем userID из контекста
	// userID, exists := c.Get("user_id")
	// if !exists {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is missing in context"})
	// 	return
	// }
	//
	// // Преобразуем userID в UUID
	// userUUID, err := uuid.Parse(userID.(string)) // userID приведен к string
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
	// 	return
	// }
	userID := "45460f1e-1e37-4fe2-9881-4398bb784452"

	response, err := w.svc.WalletService.GetBalance(c, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": response})
}
