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

type Handler struct {
	AuthHandler
}

func NewHandler(svc *service.Service, logger *logrus.Logger, cfg *config.AuthConfig) *Handler {
	return &Handler{
		AuthHandler: NewAuthHandler(svc, logger, cfg),
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
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}
	}

	return router
}
