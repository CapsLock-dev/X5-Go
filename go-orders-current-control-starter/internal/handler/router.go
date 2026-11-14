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
	httpSwagger "github.com/swaggo/http-swagger"
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
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)
	router.Get("/swagger/*", httpSwagger.Handler())
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

// PostOrder godoc
// @Summary      Create order
// @Description  Creates order, computes total price, assigns ID and new status
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        input body order.CreateOrderInput true "Order items"
// @Success      201  {object} 	order.Order
// @Failure      400  {object}  handler.ErrorResponse "Invalid JSON or order"
// @Failure      500  {object}  handler.ErrorResponse "Internal error"
// @Router       /orders [post]
func (h *Handler) postOrders(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var data order.CreateOrderInput
	if err := decoder.Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "Invalid JSON",
			Code:  ecInvalidJSON,
		})
		return
	}
	ord, err := h.service.Create(r.Context(), data)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.ErrorResponse)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/api/v1/orders/%d", ord.ID))
	writeJSON(w, http.StatusCreated, ord)
}

// GetOrder godoc
// @Summary      Get order
// @Description  Retrieves order by ID
// @Tags         orders
// @Produce      json
// @Param        id path int true "Order ID"
// @Success      200  {object} 	order.Order
// @Failure      400  {object}  handler.ErrorResponse "Invalid ID"
// @Failure      404  {object}  handler.ErrorResponse "Order not found"
// @Failure      500  {object}  handler.ErrorResponse "Internal error"
// @Router       /orders/{id} [get]
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Code: ecValidationFail, Error: "Invalid order ID"})
		return
	}
	ord, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.ErrorResponse)
		return
	}
	writeJSON(w, http.StatusOK, ord)
}

// GetOrderList godoc
// @Summary      Get list of orders
// @Description  Retrieves list of orders
// @Tags         orders
// @Produce      json
// @Success      200  {object} 	[]order.Order
// @Failure      500  {object}  handler.ErrorResponse "Internal error"
// @Router       /orders [get]
func (h *Handler) getOrderList(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.ErrorResponse)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

// UpdateStatus godoc
// @Summary      Update status
// @Description  Update status of order with given ID.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param 	     id path int true "Order ID"
// @Param 	     status body order.UpdateStatusInput true "Updated status"
// @Success      200  {object} 	order.Order
// @Failure      400  {object}  handler.ErrorResponse "Invalid ID"
// @Failure      400  {object}  handler.ErrorResponse "Invalid status"
// @Failure      404  {object}  handler.ErrorResponse "Order not found"
// @Failure      409  {object}  handler.ErrorResponse "Forbidden status transition"
// @Failure      500  {object}  handler.ErrorResponse "Internal error"
// @Router       /orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Code: ecValidationFail, Error: "Invalid order ID"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	var data order.UpdateStatusInput
	if err := decoder.Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "Invalid JSON",
			Code:  ecInvalidJSON,
		})
		return
	}
	ord, err := h.service.UpdateStatus(r.Context(), id, data.Status)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.ErrorResponse)
		return
	}
	writeJSON(w, http.StatusOK, ord)
}

// GetHistory godoc
// @Summary      Get history
// @Description  Get status update history for order by given ID
// @Tags         orders
// @Produce      json
// @Param 	     id path int true "Order ID"
// @Success      200  {object} 	[]order.StatusChange
// @Failure      400  {object}  handler.ErrorResponse "Invalid ID"
// @Failure      404  {object}  handler.ErrorResponse "Order not found"
// @Failure      500  {object}  handler.ErrorResponse "Internal error"
// @Router       /orders/{id}/history [get]
func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Code: ecValidationFail, Error: "Invalid order ID"})
		return
	}
	history, err := h.service.History(r.Context(), id)
	if err != nil {
		pe := parseError(err)
		writeJSON(w, pe.Status, pe.ErrorResponse)
		return
	}
	writeJSON(w, http.StatusOK, history)
}

type ErrorResponse struct {
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
	ErrorResponse
	Status int
}

func parseError(err error) parsedError {
	switch {
	case errors.Is(err, order.ErrInvalidOrder):
		return parsedError{
			ErrorResponse: ErrorResponse{Error: "Invalid order", Code: ecValidationFail},
			Status:        http.StatusBadRequest,
		}
	case errors.Is(err, order.ErrInvalidStatus):
		return parsedError{
			ErrorResponse: ErrorResponse{Error: "Invalid status", Code: ecValidationFail},
			Status:        http.StatusBadRequest,
		}
	case errors.Is(err, order.ErrOrderNotFound):
		return parsedError{
			ErrorResponse: ErrorResponse{Error: "Order not found", Code: ecNotFound},
			Status:        http.StatusNotFound,
		}
	case errors.Is(err, order.ErrInvalidStatusTransition):
		return parsedError{
			ErrorResponse: ErrorResponse{Error: "Status transition is not allowed", Code: ecInvalidStatusTransition},
			Status:        http.StatusConflict,
		}
	default:
		return parsedError{
			ErrorResponse: ErrorResponse{Error: "Internal error", Code: ecInternal},
			Status:        http.StatusInternalServerError,
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
