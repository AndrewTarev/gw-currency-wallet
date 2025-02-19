package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
)

func SetupAndRunServer(cfg *config.ServerConfig, handler http.Handler, logger *logrus.Logger) {
	// Создаем HTTP-сервер
	server := &http.Server{
		Addr:           cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Handler:        handler,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}

	// Канал для обработки сигналов завершения работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Infof("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Could not start server: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-stop
	logger.Info("Shutting down server...")

	// Контекст для завершения активных соединений
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server gracefully stopped")
}
