package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/storage/models/validate"
)

type Auth struct {
	svc      *service.Service
	logger   *logrus.Logger
	cfg      *config.AuthConfig
	validate *validate.Validator
}

func NewAuthHandler(
	svc *service.Service,
	logger *logrus.Logger,
	cfg *config.AuthConfig,
	validate *validate.Validator,
) *Auth {
	return &Auth{
		svc:      svc,
		logger:   logger,
		cfg:      cfg,
		validate: validate,
	}
}

func (h *Auth) Register(c *gin.Context) {
	var input *models.UserRegister

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	// Валидируем входные данные
	if err := h.validate.ValidateStruct(input); err != nil {
		c.Error(err)
		return
	}

	err := h.svc.Register(c, *input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "UserOutput registered successfully"})
}

func (h *Auth) Login(c *gin.Context) {
	var input *models.UserLogin
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	token, err := h.svc.AuthService.Login(c, input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
