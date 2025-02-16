package http

import (
	"Merch_store-Avito_test_task/internal/pkg/auth"
	httpresponse "Merch_store-Avito_test_task/internal/pkg/httpresponses"
	"Merch_store-Avito_test_task/internal/pkg/jwt"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	uc     auth.AuthUsecase
	logger *slog.Logger
	jwt    jwt.JWTInterface
}

func NewAuthHandler(uc auth.AuthUsecase, logger *slog.Logger, jwt jwt.JWTInterface) *AuthHandler {
	return &AuthHandler{uc: uc, logger: logger, jwt: jwt}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self';")
	logCtx := r.Context()
	h.logger.DebugContext(logCtx, "Handling request for log in")

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil || credentials.Username == "" || credentials.Password == "" {
		h.logger.WarnContext(logCtx, "Failed to decode credentials",
			slog.String("error", func() string {
				if err != nil {
					return err.Error()
				}
				return "empty username or password"
			}()),
		)
		response := httpresponse.Response{
			Message: "Invalid request",
		}
		httpresponse.SendJSONResponse(logCtx, w, response, http.StatusBadRequest, h.logger)
		return
	}

	credentials.Username = template.HTMLEscapeString(credentials.Username)
	credentials.Password = template.HTMLEscapeString(credentials.Password)

	user, err := h.uc.Login(r.Context(), credentials.Username, credentials.Password)
	if err != nil {
		h.logger.ErrorContext(logCtx, "Login failed: invalid username or password")

		response := httpresponse.Response{
			Message: "Invalid username or password",
		}
		httpresponse.SendJSONResponse(logCtx, w, response, http.StatusUnauthorized, h.logger)
		return
	}

	h.logger.DebugContext(logCtx, "User logged in successfully")

	token, err := h.jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.logger.ErrorContext(logCtx, "Token generation failed", slog.String("error", err.Error()))
		response := httpresponse.Response{
			Message: "Token generation failed",
		}
		httpresponse.SendJSONResponse(logCtx, w, response, http.StatusInternalServerError, h.logger)
		return
	}
	h.logger.DebugContext(logCtx, "Token generated", slog.Int("ID", int(user.ID)), slog.String("username", user.Username))
	w.Header().Set("Access-Token", token)
	h.logger.DebugContext(logCtx, "Login request completed successfully")

	httpresponse.SendJSONResponse(logCtx, w, map[string]string{"token": token}, http.StatusOK, h.logger)
}
