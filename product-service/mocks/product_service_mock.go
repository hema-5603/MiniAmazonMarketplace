package mocks

import (
	"context"
	"product-service/models"
	"github.com/stretchr/testify/mock"
)

type MockProductService struct{
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, sellerID string, req models.CreateProductRequest)(*models.Product, error){
	args := m.Called(ctx, sellerID, req)
	if args.Get(0) != nil{
		return args.Get(0).(*models.Product), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, productID, sellerID string, req models.UpdateProductRequest)(*models.Product, error){
	args := m.Called(ctx, productID, sellerID, req)
	if args.Get(0) != nil{
		return args.Get(0).(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductService) UpdateStock(ctx context.Context, productID, sellerID string, req models.UpdateProductStockRequest)error{
	return m.Called(ctx, productID, sellerID, req).Error(0)
}

func (m *MockProductService) UpdateProductStatus(ctx context.Context, productID, sellerID string, req models.UpdateStatusRequest) error{
	return m.Called(ctx, productID, sellerID, req).Error(0)
}

func (m *MockProductService) GetProducts(ctx context.Context, page, limit int, search, category string)(*models.PaginatedProductResponse, error){
	args := m.Called(ctx, page, limit, search, category)
	if args.Get(0) != nil{
		return args.Get(0).(*models.PaginatedProductResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductService) GetProductDetail(ctx context.Context, productID string) (*models.Product, error){
	args := m.Called(ctx, productID)
	if args.Get(0) != nil{
		return args.Get(0).(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductService) ValidateStock(ctx context.Context, req models.StockCheckRequest)([]models.StockCheckResult, bool, error){
	args := m.Called(ctx, req)
	return args.Get(0).([]models.StockCheckResult), args.Bool(1), args.Error(1)
}

func (m *MockProductService) ReserveStock(ctx context.Context, req models.ReserveStockRequest)error{
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockProductService) ReleaseStock(ctx context.Context, req models.ReserveStockRequest) error{
	args := m.Called(ctx, req)
	return args.Error(0)
}