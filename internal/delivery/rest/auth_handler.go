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

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя с предоставленными данными
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.UserRegister true "Данные для регистрации пользователя"
// @Success 200 {object} models.RegisterSuccessResponse
// @Failure 400 {object} middleware.ValidationErrorResponse
// @Failure 500 {object} middleware.ValidationErrorResponse
// @Router /auth/register [post]
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

	successResponse := models.RegisterSuccessResponse{
		Message: "UserOutput registered successfully",
	}

	c.JSON(http.StatusOK, successResponse)
}

// Login godoc
// @Summary Вход пользователя в систему
// @Description Авторизует пользователя и возвращает токен
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.UserLogin true "Данные для входа пользователя"
// @Success 200 {object} models.LoginSuccessResponse
// @Failure 400 {object} middleware.ValidationErrorResponse
// @Failure 401 {object} middleware.ValidationErrorResponse
// @Failure 500 {object} middleware.ValidationErrorResponse
// @Router /auth/login [post]
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

	successResponse := models.LoginSuccessResponse{
		Token: token,
	}

	c.JSON(http.StatusOK, successResponse)
}
