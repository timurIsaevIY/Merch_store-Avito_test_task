package http

import (
	"Merch_store-Avito_test_task/internal/models"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	mocks "Merch_store-Avito_test_task/internal/pkg/payments/mocks"
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"log/slog"
	h "net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSendCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockPaymentsUsecase(ctrl)
	logger := slog.Default()
	handler := NewPaymentsHandler(mockUsecase, logger)

	t.Run("successful coin transfer", func(t *testing.T) {
		mockUsecase.EXPECT().SendCoins(gomock.Any(), "user2", uint(100)).Return(nil)

		data := map[string]interface{}{
			"toUser": "user2",
			"amount": 100,
		}
		jsonData, _ := json.Marshal(data)
		req := httptest.NewRequest(h.MethodPost, "/sendCoin", bytes.NewBuffer(jsonData))
		req = req.WithContext(context.WithValue(req.Context(), middleware.IdKey, uint(1)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SendCoins(w, req)
		assert.Equal(t, h.StatusOK, w.Code)
	})

	t.Run("not enough money", func(t *testing.T) {
		mockUsecase.EXPECT().SendCoins(gomock.Any(), "user2", uint(100)).Return(models.ErrNotEnough)

		data := map[string]interface{}{
			"toUser": "user2",
			"amount": 100,
		}
		jsonData, _ := json.Marshal(data)
		req := httptest.NewRequest(h.MethodPost, "/sendCoin", bytes.NewBuffer(jsonData))
		req = req.WithContext(context.WithValue(req.Context(), middleware.IdKey, uint(1)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SendCoins(w, req)
		assert.Equal(t, h.StatusForbidden, w.Code)
	})
}

func TestBuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockPaymentsUsecase(ctrl)
	logger := slog.Default()
	handler := NewPaymentsHandler(mockUsecase, logger)

	t.Run("successful item purchase", func(t *testing.T) {
		mockUsecase.EXPECT().BuyItem(gomock.Any(), uint(1)).Return(nil)

		req := httptest.NewRequest(h.MethodGet, "/buy/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.IdKey, uint(1)))
		req = mux.SetURLVars(req, map[string]string{"item": "1"})
		w := httptest.NewRecorder()

		handler.BuyItem(w, req)
		assert.Equal(t, h.StatusOK, w.Code)
	})

	t.Run("not enough money to buy item", func(t *testing.T) {
		mockUsecase.EXPECT().BuyItem(gomock.Any(), uint(1)).Return(models.ErrNotEnough)

		req := httptest.NewRequest(h.MethodGet, "/buy/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.IdKey, uint(1)))
		req = mux.SetURLVars(req, map[string]string{"item": "1"})
		w := httptest.NewRecorder()

		handler.BuyItem(w, req)
		assert.Equal(t, h.StatusForbidden, w.Code)
	})
}
