package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"Merch_store-Avito_test_task/internal/models"
	"Merch_store-Avito_test_task/internal/pkg/httpresponses"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	mocks "Merch_store-Avito_test_task/internal/pkg/service/mocks"
	"log/slog"
)

func TestServiceHandler_GetUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceUsecase(ctrl)
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewServiceHandler(mockService, logger)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedBody   httpresponses.Response
		ctx            context.Context
	}{
		{
			name: "Successful user info retrieval",
			mockSetup: func() {
				mockService.EXPECT().GetUserInfo(gomock.Any()).Return(models.UserData{
					Coins: 5000,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   httpresponses.Response{Message: ""},
			ctx:            context.WithValue(context.Background(), middleware.IdKey, uint(1)),
		},
		{
			name:           "User not authorized",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   httpresponses.Response{Message: "User is not authorized"},
			ctx:            context.Background(), // Нет userID
		},
		{
			name: "Failed to retrieve user info",
			mockSetup: func() {
				mockService.EXPECT().GetUserInfo(gomock.Any()).Return(models.UserData{}, errors.New("DB error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   httpresponses.Response{Message: "Failed to get user info"},
			ctx:            context.WithValue(context.Background(), middleware.IdKey, uint(1)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req, err := http.NewRequest(http.MethodGet, "/user/info", nil)
			assert.NoError(t, err)

			req = req.WithContext(tt.ctx)
			rr := httptest.NewRecorder()

			handler.GetUserInfo(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response httpresponses.Response
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody.Message, response.Message)
		})
	}
}
