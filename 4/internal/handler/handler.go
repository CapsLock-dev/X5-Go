package handler

import "example.com/orders/internal/model"

type orderService interface {
	Create(customer string) (model.Order, error)
}

type Handler struct {
	orders orderService
}

func New(orders orderService) *Handler {
	return &Handler{orders: orders}
}

func (h *Handler) Create(customer string) (model.Order, error) {
	return h.orders.Create(customer)
}
