package mocks

import (
	"context"
	"order-service/models"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct{
	mock.Mock
}


func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *models.Order) error{
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error){
	args := m.Called(ctx, orderID)
	if args.Get(0) != nil{
		return args.Get(0).(*models.Order), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockOrderRepository) CountOrdersByUserID(ctx context.Context, userID string, status string) (int64, error){
	args := m.Called(ctx, userID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByUserID(ctx context.Context, userID string, status string, limit, offset int) ([]models.Order, error){
	args := m.Called(ctx, userID, status, limit, offset)
	if args.Get(0) != nil{
		return args.Get(0).([]models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) GetPendingOrdersOlderThan(ctx context.Context, threshold time.Time) ([]string, error){
	args := m.Called(ctx, threshold)
	if args.Get(0) != nil{
		return args.Get(0).([]string), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockOrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error{
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}