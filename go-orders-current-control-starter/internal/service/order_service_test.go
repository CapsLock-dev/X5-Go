package service_test

import (
	"context"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, value order.Order) (order.Order, error) {
	args := m.Called(ctx, value)
	var res order.Order
	if args.Get(0) != nil {
		res = args.Get(0).(order.Order)
	}
	return res, args.Error(1)
}
func (m *MockOrderRepository) GetByID(ctx context.Context, id int64) (order.Order, error) {
	args := m.Called(ctx, id)
	var res order.Order
	if args.Get(0) != nil {
		res = args.Get(0).(order.Order)
	}
	return res, args.Error(1)
}

func (m *MockOrderRepository) List(ctx context.Context) ([]order.Order, error) {
	args := m.Called(ctx)
	var res []order.Order
	if args.Get(0) != nil {
		res = args.Get(0).([]order.Order)
	}
	return res, args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, value order.Order) (order.Order, error) {
	args := m.Called(ctx, value)
	var res order.Order
	if args.Get(0) != nil {
		res = args.Get(0).(order.Order)
	}
	return res, args.Error(1)
}

var _ order.Repository = (*MockOrderRepository)(nil)

func TestCreateSuccess(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	repository := new(MockOrderRepository)
	service := service.NewOrderService(repository)

	var created order.Order

	repository.On("Create", mock.Anything, mock.AnythingOfType("order.Order")).Run(func(args mock.Arguments) {
		created = args.Get(1).(order.Order)
	}).Return(order.Order{
		ID:         1,
		Items:      items,
		TotalPrice: 1500,
		Status:     order.StatusNew,
	}, nil).Once()

	res, err := service.Create(context.Background(), order.CreateOrderInput{Items: items})

	require.NoError(t, err)

	assert.Equal(t, int64(1), res.ID)
	assert.Equal(t, items, res.Items)
	assert.Equal(t, int64(1500), res.TotalPrice)
	assert.Equal(t, order.StatusNew, res.Status)

	assert.Equal(t, items, created.Items)
	assert.Equal(t, int64(1500), created.TotalPrice)
	assert.Equal(t, order.StatusNew, created.Status)

	require.Len(t, created.History, 1)
	assert.Equal(t, order.StatusNew, created.History[0].Status)
	assert.Equal(t, created.CreatedAt, created.History[0].ChangedAt)

	repository.AssertExpectations(t)
}

func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name  string
		input order.CreateOrderInput
	}{
		{"EmptyItems",
			order.CreateOrderInput{Items: []order.Item{}},
		},
		{"EmptyName",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 1, Price: 100}}},
		},
		{"ZeroQuantity",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 0, Price: 100}}},
		},
		{"NegativeQuantity",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: -1, Price: 100}}},
		},
		{"ZeroPrice",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 1, Price: 0}}},
		},
		{"NegativePrice",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 1, Price: -100}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := new(MockOrderRepository)
			orderService := service.NewOrderService(repository)
			_, err := orderService.Create(context.Background(), tt.input)

			require.ErrorIs(t, err, order.ErrInvalidOrder)
			repository.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			repository.AssertExpectations(t)
		})
	}
}

func TestGetByID(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	tests := []struct {
		name    string
		input   int64
		want    order.Order
		wantErr error
	}{
		{"OrderFound",
			1,
			order.Order{ID: 1, Items: items, TotalPrice: 1500, Status: order.StatusNew},
			nil},
		{"OrderNotFound",
			2,
			order.Order{ID: 2, Items: items, TotalPrice: 1500, Status: order.StatusNew},
			order.ErrOrderNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := new(MockOrderRepository)
			repository.On("GetByID", mock.Anything, tt.input).Return(tt.want, tt.wantErr).Once()

			orderService := service.NewOrderService(repository)
			res, err := orderService.GetByID(context.TODO(), tt.input)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want.ID, res.ID)
				assert.Equal(t, tt.want.Items, res.Items)
				assert.Equal(t, tt.want.Status, res.Status)
				assert.Equal(t, tt.want.TotalPrice, res.TotalPrice)
			}
		})
	}
}

func TestGetList(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	orders := []order.Order{
		{ID: 1, Items: items, TotalPrice: 1500, Status: order.StatusNew},
		{ID: 2, Items: items, TotalPrice: 1500, Status: order.StatusNew},
	}
	t.Run("EmptyList", func(t *testing.T) {
		repository := new(MockOrderRepository)
		repository.On("List", mock.Anything).Return([]order.Order{}, nil).Once()
		orderService := service.NewOrderService(repository)

		res, err := orderService.List(context.TODO())
		require.ErrorIs(t, err, nil)
		assert.Equal(t, len(res), 0)
	})
	t.Run("ValidList", func(t *testing.T) {
		repository := new(MockOrderRepository)
		repository.On("List", mock.Anything).Return(orders, nil).Once()
		orderService := service.NewOrderService(repository)

		res, err := orderService.List(context.TODO())
		require.ErrorIs(t, err, nil)
		assert.Equal(t, res, orders)
	})
}

func TestUpdateStatusValid(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	tests := []struct {
		name   string
		status order.Status
		target order.Status
	}{
		{
			name:   "New to Processing",
			status: order.StatusNew,
			target: order.StatusProcessing,
		},
		{
			name:   "New to Cancelled",
			status: order.StatusNew,
			target: order.StatusCancelled,
		},
		{
			name:   "Processing to Completed",
			status: order.StatusProcessing,
			target: order.StatusCompleted,
		},
		{
			name:   "Processing to Cancelled",
			status: order.StatusProcessing,
			target: order.StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := new(MockOrderRepository)
			orderService := service.NewOrderService(repository)

			initialOrder := order.Order{
				ID:         1,
				Status:     tt.status,
				Items:      items,
				TotalPrice: 1500,
				History: []order.StatusChange{
					{Status: tt.status, ChangedAt: time.Now()},
				},
			}

			updatedOrder := initialOrder
			updatedOrder.Status = tt.target
			updatedOrder.History = append(updatedOrder.History, order.StatusChange{Status: tt.target, ChangedAt: time.Now()})

			repository.On("GetByID", mock.Anything, int64(1)).Return(initialOrder, nil).Once()
			repository.On("Update", mock.Anything, mock.AnythingOfType("order.Order")).Return(updatedOrder, nil).Once()

			res, err := orderService.UpdateStatus(context.Background(), 1, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.target, res.Status)
			assert.Len(t, res.History, 2)

			repository.AssertExpectations(t)
		})
	}
}

func TestUpdateStatusInvalidTransition(t *testing.T) {
	tests := []struct {
		name   string
		status order.Status
		target order.Status
	}{
		{"New to New", order.StatusNew, order.StatusNew},
		{"New to Completed", order.StatusNew, order.StatusCompleted},
		{"Processing to New", order.StatusProcessing, order.StatusNew},
		{"Processing to Processing", order.StatusProcessing, order.StatusProcessing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := new(MockOrderRepository)
			orderService := service.NewOrderService(repository)

			initialOrder := order.Order{
				ID:     1,
				Status: tt.status,
			}

			repository.On("GetByID", mock.Anything, int64(1)).
				Return(initialOrder, nil).Once()

			_, err := orderService.UpdateStatus(context.Background(), 1, tt.target)

			require.ErrorIs(t, err, order.ErrInvalidStatusTransition)
			repository.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		})
	}
}

func TestGetHistory(t *testing.T) {
	t.Run("HistoryNotFound", func(t *testing.T) {
		repository := new(MockOrderRepository)
		orderService := service.NewOrderService(repository)

		repository.On("GetByID", mock.Anything, int64(999)).Return(order.Order{}, order.ErrOrderNotFound).Once()

		_, err := orderService.History(context.Background(), 999)

		require.ErrorIs(t, err, order.ErrOrderNotFound)
		repository.AssertExpectations(t)
	})

	t.Run("HistoryFound", func(t *testing.T) {
		repository := new(MockOrderRepository)
		orderService := service.NewOrderService(repository)

		now := time.Now()
		history := []order.StatusChange{
			{Status: order.StatusNew, ChangedAt: now},
			{Status: order.StatusProcessing, ChangedAt: now.Add(time.Hour)},
			{Status: order.StatusCompleted, ChangedAt: now.Add(time.Hour).Add(time.Hour)},
		}

		repository.On("GetByID", mock.Anything, int64(1)).Return(order.Order{ID: 1, History: history}, nil).Once()

		res, err := orderService.History(context.Background(), 1)

		require.NoError(t, err)
		require.Len(t, res, 3)

		assert.Equal(t, order.StatusNew, res[0].Status)
		assert.Equal(t, order.StatusProcessing, res[1].Status)
		assert.Equal(t, order.StatusCompleted, res[2].Status)

		assert.True(t, res[0].ChangedAt.Before(res[1].ChangedAt))
		assert.True(t, res[1].ChangedAt.Before(res[2].ChangedAt))

		repository.AssertExpectations(t)
	})
}
