package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	config "gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/delivery/middleware"
	"gw-currency-wallet/internal/delivery/rest"
	"gw-currency-wallet/internal/errs"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/service/mocks"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/storage/models/validate"
)

func SetupTestEnv(t *testing.T) (
	*gin.Engine,
	*gomock.Controller,
	*service.Service,
	*validate.Validator,
	*rest.Handler,
	*config.Config,
) {
	t.Helper() // Помечаем функцию как вспомогательную для корректных логов при ошибках

	gin.SetMode(gin.TestMode)

	mockCtrl := gomock.NewController(t)
	mockSvc := &service.Service{
		AuthService:     mocks.NewMockAuthService(mockCtrl),
		ExchangeService: mocks.NewMockExchangeService(mockCtrl),
		WalletService:   mocks.NewMockWalletService(mockCtrl),
	}

	logger := logrus.New()
	cfg := &config.Config{
		Server:   config.ServerConfig{},
		Logging:  config.LoggerConfig{},
		Database: config.PostgresConfig{},
		Auth: config.AuthConfig{
			SecretKey: "secret",
			TokenTTl:  time.Second * 30,
		},
		Redis:           config.RedisConfig{},
		ExchangeService: config.ExchangeService{},
	}
	validator := validate.NewValidator()
	handler := rest.NewHandler(mockSvc, logger, &cfg.Auth, validator)

	// Настройка тестового роутера
	router := gin.New()
	router.Use(middleware.ErrorHandler(logger))
	router.Use(middleware.RecoverMiddleware(logger))

	return router, mockCtrl, mockSvc, validator, handler, cfg
}

func TestUserRegister(t *testing.T) {
	router, mockCtrl, mockSvc, validator, handler, _ := SetupTestEnv(t)
	defer mockCtrl.Finish()

	router.POST("/auth/register", middleware.ValidationMiddleware[models.UserRegister](validator), handler.Register)

	tests := []struct {
		name            string
		input           models.UserRegister
		mockServiceResp error
		expectedStatus  int
		expectedMessage string
		expectedFields  map[string]string // Добавляем проверку на поля ошибки
	}{
		{
			name: "Success - User registered successfully",
			input: models.UserRegister{
				Username: "test123",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockServiceResp: nil, // Нет ошибки
			expectedStatus:  http.StatusOK,
			expectedMessage: "User registered successfully",
			expectedFields:  nil,
		},
		{
			name: "Error - Username already exists",
			input: models.UserRegister{
				Username: "existingUser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockServiceResp: errs.ErrUserAlreadyExists,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Username already exists",
			expectedFields:  map[string]string{"username": "field already exists"},
		},
		{
			name: "Error - Invalid email format",
			input: models.UserRegister{
				Username: "test123",
				Email:    "invalid-email",
				Password: "password123",
			},
			mockServiceResp: nil,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Validation failed",
			expectedFields:  map[string]string{"Email": "must be a valid email address"},
		},
		{
			name: "Error - Weak password",
			input: models.UserRegister{
				Username: "test123",
				Email:    "test@example.com",
				Password: "123",
			},
			mockServiceResp: nil,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Validation failed",
			expectedFields:  map[string]string{"Password": "must be at least 8 characters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCall := mockSvc.AuthService.(*mocks.MockAuthService).EXPECT().
				Register(gomock.Any(), gomock.Any()).
				Return(tt.mockServiceResp)

			// Если тест на валидацию (не передается в сервис), ожидаем 0 вызовов
			if tt.expectedStatus == http.StatusBadRequest && tt.mockServiceResp == nil {
				mockCall.Times(0)
			}

			reqBody, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req) // Вместо ручного вызова хэндлера используем HTTP-запрос к маршруту

			t.Logf("HTTP статус: %d", w.Code)
			t.Logf("Ответ сервера: %s", w.Body.String())

			// Проверяем статус ответа
			if w.Code != tt.expectedStatus {
				t.Fatalf("Ожидался статус %d, но получили: %d", tt.expectedStatus, w.Code)
			}

			// Парсим JSON-ответ
			if tt.expectedStatus == http.StatusOK {
				// Проверяем успешный ответ
				var successResponse models.RegisterSuccessResponse
				if err := json.NewDecoder(w.Body).Decode(&successResponse); err != nil {
					t.Fatalf("Ошибка декодирования успешного ответа: %v. Тело ответа: %s", err, w.Body.String())
				}

				if successResponse.Message != tt.expectedMessage {
					t.Fatalf("Ожидалось сообщение '%s', но получили: '%s'", tt.expectedMessage, successResponse.Message)
				}
			} else {
				// Проверяем ошибочный ответ
				var errorResponse middleware.ValidationErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
					t.Fatalf("Ошибка декодирования ответа с ошибкой: %v. Тело ответа: %s", err, w.Body.String())
				}

				if errorResponse.Error.Message != tt.expectedMessage {
					t.Fatalf("Ожидалось сообщение ошибки '%s', но получили: '%s'", tt.expectedMessage, errorResponse.Error.Message)
				}

				// Проверяем поля ошибки
				if tt.expectedFields != nil {
					for key, expectedValue := range tt.expectedFields {
						if actualValue, exists := errorResponse.Error.Fields[key]; !exists || actualValue != expectedValue {
							t.Fatalf("Ожидалось поле ошибки '%s' со значением '%s', но получили: '%s'", key, expectedValue, actualValue)
						}
					}
				}
			}

			t.Logf("✅ Тест '%s' прошел успешно", tt.name)
		})
	}
}

func TestUserLogin(t *testing.T) {
	router, mockCtrl, mockSvc, validator, handler, _ := SetupTestEnv(t)
	defer mockCtrl.Finish()

	// Регистрируем маршрут
	router.POST("/auth/login", middleware.ValidationMiddleware[models.UserLogin](validator), handler.Login)

	tests := []struct {
		name            string
		input           models.UserLogin
		mockServiceResp string
		mockServiceErr  error
		expectedStatus  int
		expectedMessage string
	}{
		{
			name: "Success - User logged in",
			input: models.UserLogin{
				Username: "username",
				Password: "password123",
			},
			mockServiceResp: "valid-token",
			mockServiceErr:  nil,
			expectedStatus:  http.StatusOK,
			expectedMessage: "valid-token",
		},
		{
			name: "Error - Invalid password",
			input: models.UserLogin{
				Username: "username",
				Password: "wrongpassword",
			},
			mockServiceResp: "",
			mockServiceErr:  errs.ErrInvalidPassword,
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "invalid credentials",
		},
		{
			name: "Error - User not found",
			input: models.UserLogin{
				Username: "username",
				Password: "password123",
			},
			mockServiceResp: "",
			mockServiceErr:  errs.ErrUserNotFound,
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "invalid credentials",
		},
		{
			name: "Error - Empty password",
			input: models.UserLogin{
				Username: "username",
				Password: "",
			},
			mockServiceResp: "",
			mockServiceErr:  nil,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCall := mockSvc.AuthService.(*mocks.MockAuthService).EXPECT().
				Login(gomock.Any(), gomock.Any()).
				Return(tt.mockServiceResp, tt.mockServiceErr)

			// Если тест на валидацию (не передается в сервис), ожидаем 0 вызовов
			if tt.expectedStatus == http.StatusBadRequest {
				mockCall.Times(0)
			}

			reqBody, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			t.Logf("HTTP статус: %d", w.Code)
			t.Logf("Ответ сервера: %s", w.Body.String())

			if w.Code != tt.expectedStatus {
				t.Fatalf("Ожидался статус %d, но получили: %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var successResponse models.LoginSuccessResponse
				if err := json.NewDecoder(w.Body).Decode(&successResponse); err != nil {
					t.Fatalf("Ошибка декодирования успешного ответа: %v. Тело ответа: %s", err, w.Body.String())
				}

				if successResponse.Token != tt.expectedMessage {
					t.Fatalf("Ожидался токен '%s', но получили: '%s'", tt.expectedMessage, successResponse.Token)
				}
			} else {
				var errorResponse middleware.ValidationErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
					t.Fatalf("Ошибка декодирования ответа с ошибкой: %v. Тело ответа: %s", err, w.Body.String())
				}

				if errorResponse.Error.Message != tt.expectedMessage && tt.expectedStatus != http.StatusBadRequest {
					t.Fatalf("Ожидалось сообщение ошибки '%s', но получили: '%s'", tt.expectedMessage, errorResponse.Error.Message)
				}
			}

			t.Logf("✅ Тест '%s' прошел успешно", tt.name)
		})
	}
}
