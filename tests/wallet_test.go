package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"gw-currency-wallet/internal/delivery/middleware"
	"gw-currency-wallet/internal/service/mocks"
	"gw-currency-wallet/internal/storage/models"
	"gw-currency-wallet/internal/utils"
)

func TestGetBalance(t *testing.T) {
	router, mockCtrl, mockSvc, _, handler, cfg := SetupTestEnv(t)
	defer mockCtrl.Finish()

	jwtManager := utils.NewJWTManager(cfg)
	// Настроим роутер с middleware
	router.GET("/wallet/balance", middleware.AuthMiddleware(jwtManager), handler.GetBalance)

	tests := []struct {
		name              string
		mockBalanceResp   models.WalletResponse
		mockServiceErr    error
		expectedStatus    int
		expectedMessage   string
		expectedBalance   models.WalletResponse
		token             string
		expectServiceCall bool
	}{
		{
			name: "Success - Get Balance",
			mockBalanceResp: models.WalletResponse{
				BalanceRub: decimal.NewFromFloat(10000.00),
				BalanceUsd: decimal.NewFromFloat(150.00),
				BalanceEur: decimal.NewFromFloat(200.00),
			},
			mockServiceErr:  nil,
			expectedStatus:  http.StatusOK,
			expectedMessage: "",
			expectedBalance: models.WalletResponse{
				BalanceRub: decimal.NewFromFloat(10000.00),
				BalanceUsd: decimal.NewFromFloat(150.00),
				BalanceEur: decimal.NewFromFloat(200.00),
			},
			token:             generateToken(t, jwtManager, "11ff6680-c604-4231-9453-6e2fbc2c30dc", "testuser"),
			expectServiceCall: true,
		},
		{
			name:              "Error - Unauthorized",
			mockBalanceResp:   models.WalletResponse{},
			mockServiceErr:    errors.New("Unauthorized"),
			expectedStatus:    http.StatusUnauthorized,
			expectedMessage:   "token is malformed: token contains an invalid number of segments",
			expectedBalance:   models.WalletResponse{},
			token:             "invalid-token",
			expectServiceCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Эмулируем конкретный userID
			userID := uuid.Must(uuid.Parse("11ff6680-c604-4231-9453-6e2fbc2c30dc"))

			// Приводим WalletService к MockWalletService для использования EXPECT
			mockWalletService := mockSvc.WalletService.(*mocks.MockWalletService)

			if tt.expectServiceCall {
				mockWalletService.EXPECT().
					GetBalance(gomock.Any(), userID).
					Return(tt.mockBalanceResp, tt.mockServiceErr).Times(1)
			}

			// Создаем запрос и добавляем JWT токен в заголовок Authorization
			req, _ := http.NewRequest("GET", "/wallet/balance", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			t.Logf("HTTP статус: %d", w.Code)
			t.Logf("Ответ сервера: %s", w.Body.String())

			// Проверяем статус ответа
			if w.Code != tt.expectedStatus {
				t.Fatalf("Ожидался статус %d, но получили: %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var successResponse models.GetBalanceResponse
				if err := json.NewDecoder(w.Body).Decode(&successResponse); err != nil {
					t.Fatalf("Ошибка декодирования успешного ответа: %v. Тело ответа: %s", err, w.Body.String())
				}

				// Проверяем баланс
				if !successResponse.Balance.BalanceRub.Equal(tt.expectedBalance.BalanceRub) {
					t.Fatalf("Ожидался баланс RUB %s, но получили: %s",
						tt.expectedBalance.BalanceRub.String(), successResponse.Balance.BalanceRub.String())
				}
				if !successResponse.Balance.BalanceUsd.Equal(tt.expectedBalance.BalanceUsd) {
					t.Fatalf("Ожидался баланс USD %s, но получили: %s",
						tt.expectedBalance.BalanceUsd.String(), successResponse.Balance.BalanceUsd.String())
				}
				if !successResponse.Balance.BalanceEur.Equal(tt.expectedBalance.BalanceEur) {
					t.Fatalf("Ожидался баланс EUR %s, но получили: %s",
						tt.expectedBalance.BalanceEur.String(), successResponse.Balance.BalanceEur.String())
				}

			} else {
				var errorResponse map[string]string
				if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
					t.Fatalf("Ошибка декодирования ошибки: %v. Тело ответа: %s", err, w.Body.String())
				}

				if errorResponse["error"] != tt.expectedMessage {
					t.Fatalf("Ожидалось сообщение ошибки '%s', но получили: '%s'", tt.expectedMessage, errorResponse["error"])
				}
			}

			t.Logf("✅ Тест '%s' прошел успешно", tt.name)
		})
	}
}

func generateToken(t *testing.T, jwtManager *utils.JWTManager, userID, username string) string {
	uuidUserID, _ := uuid.Parse(userID)
	token, err := jwtManager.GenerateToken(uuidUserID, username)
	if err != nil {
		t.Fatalf("Ошибка генерации токена: %v", err)
	}
	return token
}
