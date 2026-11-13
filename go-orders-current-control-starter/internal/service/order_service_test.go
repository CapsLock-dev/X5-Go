package service_test

import (
	"context"
	"testing"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/repository/memory"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSuccess(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	orderInput := order.CreateOrderInput{Items: items}
	want := order.Order{ID: 1, Items: items, TotalPrice: 1500, Status: order.StatusNew}
	repository := memory.NewOrderRepository()
	orderService := service.NewOrderService(repository)
	t.Run("ValidOrder", func(t *testing.T) {
		res, err := orderService.Create(context.TODO(), orderInput)
		require.ErrorIs(t, err, nil)
		assert.Equal(t, res.Items, want.Items)
		assert.Equal(t, res.ID, want.ID)
		assert.Equal(t, res.TotalPrice, want.TotalPrice)
		require.Equal(t, len(res.History), 1)
		require.Equal(t, res.Status, want.Status)
		require.Equal(t, res.CreatedAt, res.History[0].ChangedAt)
	})
	t.Run("ID_Increment", func(t *testing.T) {
		res, err := orderService.Create(context.TODO(), orderInput)
		require.ErrorIs(t, err, nil)
		assert.Equal(t, res.ID, int64(2))
	})
}

func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    order.CreateOrderInput
		want     order.Order
		want_err error
	}{
		{"EmptyItems",
			order.CreateOrderInput{Items: []order.Item{}},
			order.Order{},
			order.ErrInvalidOrder},
		{"EmptyName",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 1, Price: 100}}},
			order.Order{},
			order.ErrInvalidOrder},
		{"ZeroQuantity",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 0, Price: 100}}},
			order.Order{},
			order.ErrInvalidOrder},
		{"NegativeQuantity",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: -1, Price: 100}}},
			order.Order{},
			order.ErrInvalidOrder},
		{"ZeroPrice",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 1, Price: 0}}},
			order.Order{},
			order.ErrInvalidOrder},
		{"NegativePrice",
			order.CreateOrderInput{Items: []order.Item{{Name: "", Quantity: 1, Price: -100}}},
			order.Order{},
			order.ErrInvalidOrder},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := memory.NewOrderRepository()
			orderService := service.NewOrderService(repository)
			res, err := orderService.Create(context.TODO(), tt.input)
			require.ErrorIs(t, err, tt.want_err)
			if tt.want_err == nil {
				assert.Equal(t, res.Items, tt.want.Items)
				assert.Equal(t, res.ID, tt.want.ID)
				assert.Equal(t, res.TotalPrice, tt.want.TotalPrice)
				assert.Equal(t, res.Status, tt.want.Status)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	tests := []struct {
		name     string
		input    int64
		want     order.Order
		want_err error
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
	repository := memory.NewOrderRepository()
	orderService := service.NewOrderService(repository)
	orderService.Create(context.TODO(), order.CreateOrderInput{Items: items})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := orderService.GetByID(context.TODO(), tt.input)
			require.ErrorIs(t, err, tt.want_err)
			if tt.want_err == nil {
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
	repository := memory.NewOrderRepository()
	orderService := service.NewOrderService(repository)
	t.Run("EmptyList", func(t *testing.T) {
		res, err := orderService.List(context.TODO())
		require.ErrorIs(t, err, nil)
		assert.Equal(t, len(res), 0)
	})
	ord1, _ := orderService.Create(context.TODO(), order.CreateOrderInput{Items: items})
	ord2, _ := orderService.Create(context.TODO(), order.CreateOrderInput{Items: items})
	orders := []order.Order{ord1, ord2}
	t.Run("ValidList", func(t *testing.T) {
		res, err := orderService.List(context.TODO())
		require.ErrorIs(t, err, nil)
		assert.Equal(t, res, orders)
	})
}

func TestUpdateStatus(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	tests := []struct {
		name         string
		initialSetup []order.Status
		orderID      int64
		targetStatus order.Status
		wantStatus   order.Status
		wantErr      error
	}{
		{
			name:         "Valid: New to Processing",
			initialSetup: nil,
			orderID:      1,
			targetStatus: order.StatusProcessing,
			wantStatus:   order.StatusProcessing,
			wantErr:      nil,
		},
		{
			name:         "Valid: New to Cancelled",
			initialSetup: nil,
			orderID:      1,
			targetStatus: order.StatusCancelled,
			wantStatus:   order.StatusCancelled,
			wantErr:      nil,
		},
		{
			name:         "Valid: Processing to Completed",
			initialSetup: []order.Status{order.StatusProcessing},
			orderID:      1,
			targetStatus: order.StatusCompleted,
			wantStatus:   order.StatusCompleted,
			wantErr:      nil,
		},
		{
			name:         "Valid: Processing to Cancelled",
			initialSetup: []order.Status{order.StatusProcessing},
			orderID:      1,
			targetStatus: order.StatusCancelled,
			wantStatus:   order.StatusCancelled,
			wantErr:      nil,
		},

		{
			name:         "Invalid: New to New",
			initialSetup: nil,
			orderID:      1,
			targetStatus: order.StatusNew,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Invalid: New to Completed",
			initialSetup: nil,
			orderID:      1,
			targetStatus: order.StatusCompleted,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Invalid: Processing to New",
			initialSetup: []order.Status{order.StatusProcessing},
			orderID:      1,
			targetStatus: order.StatusNew,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Invalid: Processing to Processing",
			initialSetup: []order.Status{order.StatusProcessing},
			orderID:      1,
			targetStatus: order.StatusProcessing,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Invalid: Completed to Processing",
			initialSetup: []order.Status{order.StatusProcessing, order.StatusCompleted},
			orderID:      1,
			targetStatus: order.StatusProcessing,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Invalid: Completed to Cancelled",
			initialSetup: []order.Status{order.StatusProcessing, order.StatusCompleted},
			orderID:      1,
			targetStatus: order.StatusCancelled,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Invalid: Cancelled to Processing",
			initialSetup: []order.Status{order.StatusCancelled},
			orderID:      1,
			targetStatus: order.StatusProcessing,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Invalid: Cancelled to Completed",
			initialSetup: []order.Status{order.StatusCancelled},
			orderID:      1,
			targetStatus: order.StatusCompleted,
			wantErr:      order.ErrInvalidStatusTransition,
		},
		{
			name:         "Error: Order Not Found",
			initialSetup: nil,
			orderID:      999,
			targetStatus: order.StatusProcessing,
			wantErr:      order.ErrOrderNotFound,
		},
		{
			name:         "Error: Unknown Status Value",
			initialSetup: nil,
			orderID:      1,
			targetStatus: order.Status("unknown_status"),
			wantErr:      order.ErrInvalidStatus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := memory.NewOrderRepository()
			orderService := service.NewOrderService(repository)

			ord, err := orderService.Create(context.TODO(), order.CreateOrderInput{Items: items})
			require.NoError(t, err)

			for _, stat := range tt.initialSetup {
				_, err := orderService.UpdateStatus(context.TODO(), ord.ID, stat)
				require.NoError(t, err)
			}
			targetID := tt.orderID
			if tt.orderID == 1 {
				targetID = ord.ID
			}
			res, err := orderService.UpdateStatus(context.TODO(), targetID, tt.targetStatus)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, targetID, res.ID)
				assert.Equal(t, tt.wantStatus, res.Status)
				assert.Equal(t, items, res.Items)
				assert.Equal(t, int64(1500), res.TotalPrice)
				expectedHistoryLen := 1 + len(tt.initialSetup) + 1
				require.Len(t, res.History, expectedHistoryLen)
				assert.Equal(t, tt.wantStatus, res.History[expectedHistoryLen-1].Status)
			}
		})
	}
}

func TestGetHistory(t *testing.T) {
	items := []order.Item{
		{Name: "Item1", Quantity: 5, Price: 100},
		{Name: "Item2", Quantity: 1, Price: 1000},
	}
	repository := memory.NewOrderRepository()
	orderService := service.NewOrderService(repository)
	t.Run("HistoryNotFound", func(t *testing.T) {
		_, err := orderService.History(context.TODO(), 999)
		require.ErrorIs(t, err, order.ErrOrderNotFound)
	})
	ord, _ := orderService.Create(context.TODO(), order.CreateOrderInput{Items: items})
	ord, _ = orderService.UpdateStatus(context.TODO(), ord.ID, order.StatusProcessing)
	ord, _ = orderService.UpdateStatus(context.TODO(), ord.ID, order.StatusCancelled)
	t.Run("ValidHistory", func(t *testing.T) {
		res, err := orderService.History(context.TODO(), ord.ID)
		require.ErrorIs(t, err, nil)
		assert.Equal(t, res, ord.History)
	})
}
