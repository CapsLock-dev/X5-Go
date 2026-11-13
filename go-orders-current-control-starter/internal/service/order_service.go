package service

import (
	"context"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, order.ErrInvalidOrder
	}
	sum := int64(0)
	for _, item := range input.Items {
		if strings.TrimSpace(item.Name) == "" {
			return order.Order{}, order.ErrInvalidOrder
		}
		if item.Price <= 0 {
			return order.Order{}, order.ErrInvalidOrder
		}
		if item.Quantity <= 0 {
			return order.Order{}, order.ErrInvalidOrder
		}
		sum += item.Price * int64(item.Quantity)
	}
	now := time.Now()
	tempOrder := order.Order{
		Items:      input.Items,
		TotalPrice: sum,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History:    []order.StatusChange{{Status: order.StatusNew, ChangedAt: now}},
	}
	return s.repository.Create(ctx, tempOrder)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	// TODO: validate the requested transition, update current status and history,
	// then persist the changed order.
	switch status {
	case order.StatusNew, order.StatusProcessing, order.StatusCompleted, order.StatusCancelled:
		break
	default:
		return order.Order{}, order.ErrInvalidStatus
	}
	ord, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}
	var statusCorrect bool
	switch ord.Status {
	case order.StatusNew:
		statusCorrect = status == order.StatusProcessing || status == order.StatusCancelled
	case order.StatusProcessing:
		statusCorrect = status == order.StatusCompleted || status == order.StatusCancelled
	default:
		statusCorrect = false
	}
	if !statusCorrect {
		return order.Order{}, order.ErrInvalidStatusTransition
	}
	ord.History = append(ord.History, order.StatusChange{Status: status, ChangedAt: time.Now()})
	ord.Status = status
	return s.repository.Update(ctx, ord)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	// TODO: return status changes in chronological order.
	ord, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ord.History, nil
}
