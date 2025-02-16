package http

import (
	"Merch_store-Avito_test_task/internal/models"
	"Merch_store-Avito_test_task/internal/pkg/httpresponses"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	"Merch_store-Avito_test_task/internal/pkg/payments"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"strconv"
)

type PaymentsHandler struct {
	uc     payments.PaymentsUsecase
	logger *slog.Logger
}

func NewPaymentsHandler(uc payments.PaymentsUsecase, logger *slog.Logger) *PaymentsHandler {
	return &PaymentsHandler{uc: uc, logger: logger}
}

func (h *PaymentsHandler) SendCoins(w http.ResponseWriter, r *http.Request) {
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
	type Data struct {
		ToUser string `json:"toUser"`
		Amount uint   `json:"amount"`
	}
	var data Data
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to decode request body:", slog.String("err", err.Error()))
		response := httpresponses.Response{
			Message: "failed to decode request body",
		}
		httpresponses.SendJSONResponse(ctx, w, response, http.StatusBadRequest, h.logger)
		return
	}
	err = h.uc.SendCoins(ctx, data.ToUser, data.Amount)
	if err != nil {
		if errors.Is(err, models.ErrNotEnough) {
			h.logger.ErrorContext(ctx, "not enough money:", slog.String("err", err.Error()))
			response := httpresponses.Response{
				Message: "not enough money",
			}
			httpresponses.SendJSONResponse(ctx, w, response, http.StatusForbidden, h.logger)
			return
		} else if errors.Is(err, models.ErrNotFound) {
			h.logger.ErrorContext(ctx, "not found:", slog.String("err", err.Error()))
			response := httpresponses.Response{
				Message: "not found",
			}
			httpresponses.SendJSONResponse(ctx, w, response, http.StatusNotFound, h.logger)
			return
		}
		h.logger.ErrorContext(ctx, "failed to send coins:", slog.String("err", err.Error()))
		response := httpresponses.Response{
			Message: "failed to send coins",
		}
		httpresponses.SendJSONResponse(ctx, w, response, http.StatusInternalServerError, h.logger)
		return
	}
	h.logger.DebugContext(ctx, "successfully sent coins to user: %v", slog.String("toUser", data.ToUser))
	httpresponses.SendJSONResponse(ctx, w, nil, http.StatusOK, h.logger)
}

func (h *PaymentsHandler) BuyItem(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	itemId, err := strconv.Atoi(vars["item"])
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to parse item id:", slog.String("err", err.Error()))
		response := httpresponses.Response{
			Message: "failed to parse item id",
		}
		httpresponses.SendJSONResponse(ctx, w, response, http.StatusBadRequest, h.logger)
		return
	}
	err = h.uc.BuyItem(ctx, uint(itemId))
	if err != nil {
		if errors.Is(err, models.ErrNotEnough) {
			h.logger.ErrorContext(ctx, "not enough money:", slog.String("err", err.Error()))
			response := httpresponses.Response{
				Message: "not enough money",
			}
			httpresponses.SendJSONResponse(ctx, w, response, http.StatusForbidden, h.logger)
			return
		}
		h.logger.ErrorContext(ctx, "failed to buy item:", slog.String("err", err.Error()))
		response := httpresponses.Response{
			Message: "failed to buy item",
		}
		httpresponses.SendJSONResponse(ctx, w, response, http.StatusInternalServerError, h.logger)
		return
	}
	h.logger.DebugContext(ctx, "successfully buy item to user: %v", slog.Int("itemId", itemId))
	httpresponses.SendJSONResponse(ctx, w, nil, http.StatusOK, h.logger)
}
