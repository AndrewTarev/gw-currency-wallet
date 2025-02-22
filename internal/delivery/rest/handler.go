package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/delivery/middleware"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/storage/models/validate"
	"gw-currency-wallet/internal/utils"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type Exchange interface {
	GetExchangeRates(c *gin.Context)
	ExchangeCurrency(c *gin.Context)
}

type WalletHandler interface {
	GetBalance(c *gin.Context)
	Deposit(c *gin.Context)
	Withdraw(c *gin.Context)
}

type Handler struct {
	AuthHandler
	Exchange
	WalletHandler
}

func NewHandler(
	svc *service.Service,
	logger *logrus.Logger,
	cfg *config.AuthConfig,
	validate *validate.Validator,
) *Handler {
	return &Handler{
		AuthHandler:   NewAuthHandler(svc, logger, cfg, validate),
		Exchange:      NewExchangeHandler(svc, validate),
		WalletHandler: NewWalletHandler(svc, validate),
	}
}

func (h *Handler) InitRoutes(logger *logrus.Logger, jwtManager *utils.JWTManager) *gin.Engine {
	router := gin.New()

	// Обработчик ошибок и паник
	router.Use(middleware.ErrorHandler(logger))
	router.Use(middleware.RecoverMiddleware(logger))

	// docs.SwaggerInfo.BasePath = "/api/v1"
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")

	// Группа маршрутов без авторизации
	auth := apiV1.Group("")
	{
		auth.POST("/register", h.AuthHandler.Register)
		auth.POST("/login", h.AuthHandler.Login)
	}

	// Группа маршрутов с авторизацией
	protected := apiV1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		wallet := protected.Group("/wallet")
		{
			wallet.GET("/balance", h.WalletHandler.GetBalance)
			wallet.POST("/deposit", h.WalletHandler.Deposit)
			wallet.POST("/withdraw", h.WalletHandler.Withdraw)
		}
		exchange := protected.Group("/exchange")
		{
			exchange.GET("/rates", h.Exchange.GetExchangeRates)
			exchange.POST("/", h.Exchange.ExchangeCurrency)
		}
	}

	return router
}
