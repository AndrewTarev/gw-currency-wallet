package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"gw-currency-wallet/internal/delivery/middleware"
	"gw-currency-wallet/internal/service/mocks"
	"gw-currency-wallet/internal/storage/models"
)

func TestGetExchangeRates(t *testing.T) {
	router, mockCtrl, mockSvc, _, handler, _ := SetupTestEnv(t)
	defer mockCtrl.Finish()

	// Регистрируем маршрут
	router.GET("/exchange/rates", handler.GetExchangeRates)

	tests := []struct {
		name            string
		mockServiceResp map[string]string
		mockServiceErr  error
		expectedStatus  int
		expectedRates   map[string]string
		expectedMessage string
	}{
		{
			name: "Success - Get exchange rates",
			mockServiceResp: map[string]string{
				"USD": "92.5",
				"EUR": "100.3",
				"RUB": "1.0",
			},
			mockServiceErr: nil,
			expectedStatus: http.StatusOK,
			expectedRates: map[string]string{
				"USD": "92.5",
				"EUR": "100.3",
				"RUB": "1.0",
			},
			expectedMessage: "",
		},
		{
			name:            "Error - Service failure",
			mockServiceResp: nil,
			mockServiceErr:  errors.New("Internal server error"),
			expectedStatus:  http.StatusInternalServerError,
			expectedRates:   nil,
			expectedMessage: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCall := mockSvc.ExchangeService.(*mocks.MockExchangeService).EXPECT().
				GetRates(gomock.Any()).
				Return(tt.mockServiceResp, tt.mockServiceErr)

			// Если сервис должен вернуть ошибку, ожидание вызова мок-функции остается
			if tt.mockServiceErr != nil {
				mockCall.Times(1)
			}

			req, _ := http.NewRequest("GET", "/exchange/rates", nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			t.Logf("HTTP статус: %d", w.Code)
			t.Logf("Ответ сервера: %s", w.Body.String())

			if w.Code != tt.expectedStatus {
				t.Fatalf("Ожидался статус %d, но получили: %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var successResponse models.ExchangeRatesResponse
				if err := json.NewDecoder(w.Body).Decode(&successResponse); err != nil {
					t.Fatalf("Ошибка декодирования успешного ответа: %v. Тело ответа: %s", err, w.Body.String())
				}

				if !reflect.DeepEqual(successResponse.Rates, tt.expectedRates) {
					t.Fatalf("Ожидались курсы валют %+v, но получили: %+v", tt.expectedRates, successResponse.Rates)
				}
			} else {
				var errorResponse middleware.ValidationErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
					t.Fatalf("Ошибка декодирования ответа с ошибкой: %v. Тело ответа: %s", err, w.Body.String())
				}

				if errorResponse.Error.Message != tt.expectedMessage {
					t.Fatalf("Ожидалось сообщение ошибки '%s', но получили: '%s'", tt.expectedMessage, errorResponse.Error.Message)
				}
			}

			t.Logf("✅ Тест '%s' прошел успешно", tt.name)
		})
	}
}

func TestExchangeCurrency(t *testing.T) {
	router, mockCtrl, mockSvc, validator, handler, _ := SetupTestEnv(t)
	defer mockCtrl.Finish()

	// Настроим роутер с middleware
	router.POST("/exchange", middleware.ValidationMiddleware[models.ExchangeRequest](validator), handler.ExchangeCurrency)

	tests := []struct {
		name               string
		input              models.ExchangeRequest
		mockRate           string
		mockExchangeResp   models.WalletResponse
		mockServiceResp    error
		expectedStatus     int
		expectedMessage    string
		expectedNewBalance models.WalletResponse
		expectServiceCalls bool
	}{
		{
			name: "Success - Exchange RUB to USD",
			input: models.ExchangeRequest{
				FromCurrency: "RUB",
				ToCurrency:   "USD",
				Amount:       1000.00,
			},
			mockRate: "0.013",
			mockExchangeResp: models.WalletResponse{
				BalanceRub: decimal.NewFromFloat(9000.00),
				BalanceUsd: decimal.NewFromFloat(13.00),
				BalanceEur: decimal.NewFromFloat(0.00),
			},
			mockServiceResp: nil,
			expectedStatus:  http.StatusOK,
			expectedMessage: "Exchange successful",
			expectedNewBalance: models.WalletResponse{
				BalanceRub: decimal.NewFromFloat(9000.00),
				BalanceUsd: decimal.NewFromFloat(13.00),
				BalanceEur: decimal.NewFromFloat(0.00),
			},
			expectServiceCalls: true,
		},
		{
			name: "Error - Invalid currency",
			input: models.ExchangeRequest{
				FromCurrency: "RUB",
				ToCurrency:   "INVALID", // Некорректная валюта
				Amount:       500.00,
			},
			mockRate:         "",                              // Пустой курс
			mockExchangeResp: models.WalletResponse{},         // Пустой ответ
			mockServiceResp:  errors.New("Validation failed"), // Ошибка
			expectedStatus:   http.StatusBadRequest,
			expectedMessage:  "Validation failed",
			expectedNewBalance: models.WalletResponse{
				BalanceRub: decimal.Zero,
				BalanceUsd: decimal.Zero,
				BalanceEur: decimal.Zero,
			},
			expectServiceCalls: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Эмулируем конкретный userID
			userID := uuid.Must(uuid.Parse("11ff6680-c604-4231-9453-6e2fbc2c30dc"))

			// Приводим ExchangeService к MockExchangeService для использования EXPECT
			mockExchangeService := mockSvc.ExchangeService.(*mocks.MockExchangeService)

			if tt.expectServiceCalls {
				// Мокаем вызовы сервисов, используем gomock.Any() для UUID
				mockExchangeService.EXPECT().
					ExchangeCurrency(gomock.Any(), gomock.Any(), tt.input.FromCurrency, tt.input.ToCurrency, gomock.Any(), gomock.Any()).
					Return(tt.mockExchangeResp, tt.mockServiceResp).Times(1)

				mockExchangeService.EXPECT().
					GetRate(gomock.Any(), tt.input.FromCurrency, tt.input.ToCurrency).
					Return(tt.mockRate, tt.mockServiceResp).Times(1)
			}

			// Создаем запрос и вручную ставим userID в контекст запроса
			reqBody, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/exchange", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Эмулируем добавление userID в контекст через middleware
			ctx := req.Context()
			ctx = context.WithValue(ctx, "user_id", userID) // Ключ должен совпадать с тем, что используется в middleware
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			t.Logf("HTTP статус: %d", w.Code)
			t.Logf("Ответ сервера: %s", w.Body.String())

			// Проверяем статус ответа
			if w.Code != tt.expectedStatus {
				t.Fatalf("Ожидался статус %d, но получили: %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var successResponse models.ExchangeCurrencyResponse
				if err := json.NewDecoder(w.Body).Decode(&successResponse); err != nil {
					t.Fatalf("Ошибка декодирования успешного ответа: %v. Тело ответа: %s", err, w.Body.String())
				}

				if successResponse.Message != tt.expectedMessage {
					t.Fatalf("Ожидалось сообщение '%s', но получили: '%s'", tt.expectedMessage, successResponse.Message)
				}

				// Проверяем баланс
				if !successResponse.NewBalance.BalanceRub.Equal(tt.expectedNewBalance.BalanceRub) {
					t.Fatalf("Ожидался баланс RUB %s, но получили: %s",
						tt.expectedNewBalance.BalanceRub.String(), successResponse.NewBalance.BalanceRub.String())
				}
				if !successResponse.NewBalance.BalanceUsd.Equal(tt.expectedNewBalance.BalanceUsd) {
					t.Fatalf("Ожидался баланс USD %s, но получили: %s",
						tt.expectedNewBalance.BalanceUsd.String(), successResponse.NewBalance.BalanceUsd.String())
				}
				if !successResponse.NewBalance.BalanceEur.Equal(tt.expectedNewBalance.BalanceEur) {
					t.Fatalf("Ожидался баланс EUR %s, но получили: %s",
						tt.expectedNewBalance.BalanceEur.String(), successResponse.NewBalance.BalanceEur.String())
				}

			} else {
				var errorResponse middleware.ValidationErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
					t.Fatalf("Ошибка декодирования ошибки: %v. Тело ответа: %s", err, w.Body.String())
				}

				if errorResponse.Error.Message != tt.expectedMessage {
					t.Fatalf("Ожидалось сообщение ошибки '%s', но получили: '%s'", tt.expectedMessage, errorResponse.Error.Message)
				}
			}

			t.Logf("✅ Тест '%s' прошел успешно", tt.name)
		})
	}
}
