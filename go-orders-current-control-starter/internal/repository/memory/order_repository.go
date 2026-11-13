package memory

import (
	"context"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = r.nextID
	r.nextID++
	r.items = append(r.items, orderDeepCopy(value))
	return orderDeepCopy(value), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, item := range r.items {
		if item.ID == id {
			return orderDeepCopy(item), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	list := make([]order.Order, 0, len(r.items))
	for _, item := range r.items {
		list = append(list, orderDeepCopy(item))
	}
	return list, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i, item := range r.items {
		if item.ID == value.ID {
			r.items[i] = orderDeepCopy(value)
			return orderDeepCopy(value), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func orderDeepCopy(o order.Order) order.Order {
	if o.Items != nil {
		temp := make([]order.Item, len(o.Items))
		copy(temp, o.Items)
		o.Items = temp
	}
	if o.History != nil {
		temp := make([]order.StatusChange, len(o.History))
		copy(temp, o.History)
		o.History = temp
	}
	return o
}
