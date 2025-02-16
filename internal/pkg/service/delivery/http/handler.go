package http

import (
	"Merch_store-Avito_test_task/internal/pkg/httpresponses"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	"Merch_store-Avito_test_task/internal/pkg/service"
	"log/slog"
	"net/http"
)

type ServiceHandler struct {
	uc     service.ServiceUsecase
	logger *slog.Logger
}

func NewServiceHandler(uc service.ServiceUsecase, logger *slog.Logger) *ServiceHandler {
	return &ServiceHandler{uc: uc, logger: logger}
}

func (h *ServiceHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := ctx.Value(middleware.IdKey).(uint)
	if !ok {
		h.logger.ErrorContext(ctx, "failed to retrieve user id from context")
		response := httpresponses.Response{
			Message: "User is not authorized",
		}
		httpresponses.SendJSONResponse(ctx, w, response, http.StatusUnauthorized, h.logger)
		return
	}
	info, err := h.uc.GetUserInfo(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to retrieve user info from context")
		response := httpresponses.Response{
			Message: "Failed to get user info",
		}
		httpresponses.SendJSONResponse(ctx, w, response, http.StatusInternalServerError, h.logger)
		return
	}
	httpresponses.SendJSONResponse(ctx, w, info, http.StatusOK, h.logger)
}
