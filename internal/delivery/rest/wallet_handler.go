package rest

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/storage/models/validate"
)

type Wallet struct {
	svc      *service.Service
	validate *validate.Validator
}

func NewWalletHandler(svc *service.Service, validate *validate.Validator) *Wallet {
	return &Wallet{
		svc:      svc,
		validate: validate,
	}
}

// GetUserUUID отдает id юзера и превращает в формат uuid
func GetUserUUID(c *gin.Context) (uuid.UUID, error) {
	// Получаем userID из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.UUID{}, fmt.Errorf("user_id is missing in context")
	}

	// Преобразуем userID в UUID
	userUUID, err := uuid.Parse(userID.(string)) // userID приведен к string
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid user_id format")
	}

	return userUUID, nil
}

func (w *Wallet) GetBalance(c *gin.Context) {
	userID, err := GetUserUUID(c)
	if err != nil {
		c.Error(err)
	}

	response, err := w.svc.WalletService.GetBalance(c, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": response})
}

func (w *Wallet) Deposit(c *gin.Context) {
	userID, err := GetUserUUID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var input models.WalletTransaction

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	if err := w.validate.ValidateStruct(input); err != nil {
		c.Error(err)
		return
	}

	balance, err := w.svc.WalletService.Deposit(c, userID, input.Currency, decimal.NewFromFloat(input.Amount))
	if err != nil {
		c.Error(err)
		return
	}

	successResponse := struct {
		Message string                `json:"message"`
		Balance models.WalletResponse `json:"new_balance"`
	}{
		Message: "Account topped up successfully",
		Balance: balance,
	}

	c.JSON(http.StatusOK, successResponse)
}

func (w *Wallet) Withdraw(c *gin.Context) {
	userID, err := GetUserUUID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var input models.WalletTransaction
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	if err := w.validate.ValidateStruct(input); err != nil {
		c.Error(err)
		return
	}
	balance, err := w.svc.WalletService.Withdraw(c, userID, input.Currency, decimal.NewFromFloat(input.Amount))
	if err != nil {
		c.Error(err)
		return
	}

	successResponse := struct {
		Message string                `json:"message"`
		Balance models.WalletResponse `json:"new_balance"`
	}{
		Message: "Withdrawal successful",
		Balance: balance,
	}

	c.JSON(http.StatusOK, successResponse)
}
