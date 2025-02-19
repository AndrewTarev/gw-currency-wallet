package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/storage/models"
)

type Auth struct {
	svc    *service.Service
	logger *logrus.Logger
	cfg    *config.AuthConfig
}

func NewAuthHandler(svc *service.Service, logger *logrus.Logger, cfg *config.AuthConfig) *Auth {
	return &Auth{
		svc:    svc,
		logger: logger,
		cfg:    cfg,
	}
}

func (h *Auth) Register(c *gin.Context) {
	var input *models.UserRegister

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	if err := input.Validate(); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			c.Error(validationErrs)
			return
		}
	}

	err := h.svc.Register(c, input)
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
