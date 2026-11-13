package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(recoveryMiddleware(logger))
	router.Use(requestLoggingMiddleware(logger))

	router.Get("/health", h.health)
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.postOrders)
		r.Get("/", h.getOrderList)
		r.Get("/{id}", h.getOrder)
		r.Patch("/{id}/status", h.updateStatus)
		r.Get("/{id}/history", h.getHistory)
	})

	return router
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) postOrders(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var data order.CreateOrderInput
	err := decoder.Decode(&data)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.errorResponse)
		return
	}
	ord, err := h.service.Create(r.Context(), data)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.errorResponse)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/api/v1/orders/%d", ord.ID))
	writeJSON(w, http.StatusCreated, ord)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Code: ecValidationFail, Error: "Invalid order ID"})
		return
	}
	ord, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.errorResponse)
		return
	}
	writeJSON(w, http.StatusOK, ord)
}

func (h *Handler) getOrderList(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.errorResponse)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Code: ecValidationFail, Error: "Invalid order ID"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	var data order.UpdateStatusInput
	err = decoder.Decode(&data)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.errorResponse)
		return
	}
	ord, err := h.service.UpdateStatus(r.Context(), id, data.Status)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.errorResponse)
		return
	}
	writeJSON(w, http.StatusCreated, ord)
}

func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Code: ecValidationFail, Error: "Invalid order ID"})
		return
	}
	history, err := h.service.History(r.Context(), id)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.errorResponse)
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (h *Handler) notImplemented(w http.ResponseWriter, _ *http.Request) {
	// TODO: replace this placeholder with separate handlers for every endpoint.
	writeJSON(w, http.StatusNotImplemented, errorResponse{
		Error: "endpoint is not implemented",
		Code:  "NOT_IMPLEMENTED",
	})
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

const (
	ecInvalidJSON             = "INVALID_JSON"
	ecValidationFail          = "VALIDATION_ERROR"
	ecNotFound                = "NOT_FOUND"
	ecInternal                = "INTERNAL_ERROR"
	ecInvalidStatusTransition = "INVALID_STATUS_TRANSITION"
)

type parsedError struct {
	errorResponse
	Status int
}

func parseError(err error) parsedError {
	switch {
	case errors.Is(err, order.ErrInvalidOrder):
		return parsedError{
			errorResponse: errorResponse{Error: "Invalid order", Code: ecValidationFail},
			Status:        http.StatusBadRequest,
		}
	case errors.Is(err, order.ErrInvalidStatus):
		return parsedError{
			errorResponse: errorResponse{Error: "Invalid status", Code: ecValidationFail},
			Status:        http.StatusBadRequest,
		}
	case errors.Is(err, order.ErrOrderNotFound):
		return parsedError{
			errorResponse: errorResponse{Error: "Order not found", Code: ecNotFound},
			Status:        http.StatusNotFound,
		}
	case errors.Is(err, order.ErrInvalidStatusTransition):
		return parsedError{
			errorResponse: errorResponse{Error: "Status transition is not allowed", Code: ecInvalidStatusTransition},
			Status:        http.StatusBadRequest,
		}
	default:
		return parsedError{
			errorResponse: errorResponse{Error: "Internal error", Code: ecInternal},
			Status:        http.StatusInternalServerError,
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
