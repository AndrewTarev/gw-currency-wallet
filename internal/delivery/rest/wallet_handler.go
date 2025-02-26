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

// GetBalance godoc
// @Summary Получить баланс кошелька
// @Description Возвращает текущий баланс пользователя во всех валютах
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.GetBalanceResponse
// @Failure 400 {object} middleware.ValidationErrorResponse
// @Failure 401 {object} middleware.ValidationErrorResponse
// @Failure 500 {object} middleware.ValidationErrorResponse
// @Router /wallet/balance [get]
func (w *Wallet) GetBalance(c *gin.Context) {
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		c.Error(err)
	}

	response, err := w.svc.WalletService.GetBalance(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	successResponse := models.GetBalanceResponse{
		Balance: response,
	}

	c.JSON(http.StatusOK, successResponse)
}

// Deposit godoc
// @Summary Пополнить баланс
// @Description Пополняет баланс пользователя на указанную сумму в указанной валюте
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.WalletTransaction true "Данные для пополнения"
// @Success 200 {object} models.WalletOperationsResponse
// @Failure 400 {object} middleware.ValidationErrorResponse
// @Failure 401 {object} middleware.ValidationErrorResponse
// @Failure 500 {object} middleware.ValidationErrorResponse
// @Router /wallet/deposit [post]
func (w *Wallet) Deposit(c *gin.Context) {
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		c.Error(err)
	}

	input, exists := c.Get("validatedInput")
	if !exists {
		c.Error(errs.ErrValidationNotWorking)
		return
	}

	userInput := input.(models.WalletTransaction)

	balance, err := w.svc.WalletService.Deposit(c, userID, userInput.Currency, decimal.NewFromFloat(userInput.Amount))
	if err != nil {
		c.Error(err)
		return
	}

	successResponse := models.WalletOperationsResponse{
		Message: "Account topped up successfully",
		Balance: balance,
	}

	c.JSON(http.StatusOK, successResponse)
}

// Withdraw godoc
// @Summary Снять средства
// @Description Списывает указанную сумму в указанной валюте с баланса пользователя
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.WalletTransaction true "Данные для снятия средств"
// @Success 200 {object} models.WalletOperationsResponse
// @Failure 400 {object} middleware.ValidationErrorResponse
// @Failure 401 {object} middleware.ValidationErrorResponse
// @Failure 500 {object} middleware.ValidationErrorResponse
// @Router /wallet/withdraw [post]
func (w *Wallet) Withdraw(c *gin.Context) {
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		c.Error(err)
		return
	}

	input, exists := c.Get("validatedInput")
	if !exists {
		c.Error(errs.ErrValidationNotWorking)
		return
	}

	userInput := input.(models.WalletTransaction)

	balance, err := w.svc.WalletService.Withdraw(c, userID, userInput.Currency, decimal.NewFromFloat(userInput.Amount))
	if err != nil {
		c.Error(err)
		return
	}

	successResponse := models.WalletOperationsResponse{
		Message: "Withdrawal successful",
		Balance: balance,
	}

	c.JSON(http.StatusOK, successResponse)
}
