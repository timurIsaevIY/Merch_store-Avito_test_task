package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"Merch_store-Avito_test_task/internal/models"
	mocks "Merch_store-Avito_test_task/internal/pkg/auth/mocks"
	"Merch_store-Avito_test_task/internal/pkg/httpresponses"
	jwtmock "Merch_store-Avito_test_task/internal/pkg/jwt/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log/slog"
)

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)
	mockJWT := jwtmock.NewMockJWTInterface(ctrl)
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	handler := NewAuthHandler(mockUsecase, logger, mockJWT)

	tests := []struct {
		name           string
		requestBody    map[string]string
		mockSetup      func()
		expectedStatus int
		expectedBody   httpresponses.Response
	}{
		{
			name: "Successful login",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password",
			},
			mockSetup: func() {
				mockUsecase.EXPECT().Login(gomock.Any(), "testuser", "password").Return(models.User{ID: 1, Username: "testuser"}, nil)
				mockJWT.EXPECT().GenerateToken(uint(1), "testuser").Return("valid_token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   httpresponses.Response{Message: ""},
		},
		{
			name: "Invalid username or password",
			requestBody: map[string]string{
				"username": "wronguser",
				"password": "wrongpass",
			},
			mockSetup: func() {
				mockUsecase.EXPECT().Login(gomock.Any(), "wronguser", "wrongpass").Return(models.User{}, errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   httpresponses.Response{Message: "Invalid username or password"},
		},
		{
			name: "Token generation failure",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password",
			},
			mockSetup: func() {
				mockUsecase.EXPECT().Login(gomock.Any(), "testuser", "password").Return(models.User{ID: 1, Username: "testuser"}, nil)
				mockJWT.EXPECT().GenerateToken(uint(1), "testuser").Return("", errors.New("token error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   httpresponses.Response{Message: "Token generation failed"},
		},
		{
			name:           "Invalid JSON request",
			requestBody:    map[string]string{},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   httpresponses.Response{Message: "Invalid request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			reqBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/auth", bytes.NewReader(reqBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, response, "token")
			} else {
				assert.Equal(t, tt.expectedBody.Message, response["message"])
			}
		})
	}
}
