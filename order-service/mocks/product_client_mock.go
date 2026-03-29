package mocks

import (
	"context"
	"order-service/client"
	"order-service/models"

	"github.com/stretchr/testify/mock"
)

type MockProductClient struct{
	mock.Mock
}

func (m *MockProductClient) ValidateCart(ctx context.Context, items []models.CheckoutItem) (*client.ProductValidationResponse, error){
	args := m.Called(ctx, items)
	if args.Get(0) != nil{
		return args.Get(0).(*client.ProductValidationResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductClient) ReserveStock(ctx context.Context, items []models.CheckoutItem) error{
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockProductClient) ReleaseStock(ctx context.Context, items []models.OrderItem) error{
	args := m.Called(ctx, items)
	return args.Error(0)
}