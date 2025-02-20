package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/delivery/middleware"
	"gw-currency-wallet/internal/service"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type WalletHandler interface{}

type Exchange interface {
	GetExchangeRates(c *gin.Context)
}

type Handler struct {
	AuthHandler
	// WalletHandler
	Exchange
}

func NewHandler(svc *service.Service, logger *logrus.Logger, cfg *config.AuthConfig) *Handler {
	return &Handler{
		AuthHandler: NewAuthHandler(svc, logger, cfg),
		Exchange:    NewExchangeHandler(svc),
	}
}

func (h *Handler) InitRoutes(logger *logrus.Logger) *gin.Engine {
	router := gin.New()

	// Обработчик ошибок
	router.Use(middleware.ErrorHandler(logger))
	// Обработчик паник
	router.Use(middleware.RecoverMiddleware(logger))

	// docs.SwaggerInfo.BasePath = "/api/v1"
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV1 := router.Group("/api/v1")
	{
		auth := apiV1.Group("")
		{
			auth.POST("/register", h.AuthHandler.Register)
			auth.POST("/login", h.AuthHandler.Login)
		}
		wallet := apiV1.Group("/wallet")
		{
			wallet.GET("/balance")
			wallet.POST("/deposit")
			wallet.POST("/withdraw")
		}
		exchange := apiV1.Group("/exchange")
		{
			exchange.GET("/rates", h.Exchange.GetExchangeRates)
			exchange.POST("/")
		}
	}

	return router
}
