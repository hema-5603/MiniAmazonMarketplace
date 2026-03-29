package mocks

import (
	"context"
	"order-service/models"

	"github.com/stretchr/testify/mock"
)

type MockOrderService struct{
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID string, req models.CheckoutRequest) (*models.Order, error){
	args := m.Called(ctx, userID, req)
	if args.Get(0) != nil{
		return args.Get(0).(*models.Order), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockOrderService) GetOrderDetail(ctx context.Context, orderID string, userID string) (*models.Order, error){
	args := m.Called(ctx, orderID, userID)
	if args.Get(0) != nil{
		return args.Get(0).(*models.Order), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockOrderService) GetOrderHistory(ctx context.Context, userID string, page, limit int, status string) (*models.PaginatedOrderResponse, error){
	args := m.Called(ctx, userID, page, limit, status)
	if args.Get(0) != nil{
		return args.Get(0).(*models.PaginatedOrderResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockOrderService) ExpireUnpaidOrders(ctx context.Context) error{
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOrderService) CancelOrder(ctx context.Context, orderID string, userID string) error{
	args := m.Called(ctx, orderID, userID)
	return args.Error(0)
}