package middleware

import (
	"Merch_store-Avito_test_task/internal/pkg/httpresponses"
	"Merch_store-Avito_test_task/internal/pkg/jwt"
	"context"
	"log/slog"
	"net/http"
)

type ContextKey string

const (
	IdKey       ContextKey = "userID"
	UsernameKey ContextKey = "username"
)

func AuthMiddleware(jwtService jwt.JWTInterface, next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Access-Token")
		if token == "" {
			response := httpresponses.Response{
				Message: "token is missing",
			}
			logger.Error("token is missing")
			httpresponses.SendJSONResponse(r.Context(), w, response, http.StatusUnauthorized, logger)
			return
		}
		claims, err := jwtService.ParseToken(token)
		if err != nil || claims == nil {
			response := httpresponses.Response{
				Message: "token is invalid",
			}
			logger.Error("token is invalid", slog.Any("error", err.Error()))
			httpresponses.SendJSONResponse(r.Context(), w, response, http.StatusUnauthorized, logger)
			return
		}
		userIDFloat, ok1 := claims["userID"].(float64)
		username, ok2 := claims["username"].(string)
		if !ok1 || !ok2 {
			logger.Error("Invalid token claims")
			response := httpresponses.Response{
				Message: "Invalid token claims",
			}
			httpresponses.SendJSONResponse(r.Context(), w, response, http.StatusUnauthorized, logger)
			return
		}
		userID := uint(userIDFloat)

		logger.Debug("Token parsed", slog.Int("userID", int(userID)), slog.String("username", username))
		ctx := context.WithValue(r.Context(), IdKey, userID)
		ctx = context.WithValue(ctx, UsernameKey, username)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
