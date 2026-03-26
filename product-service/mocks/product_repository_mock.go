package mocks

import(
	"context"
	"product-service/models"
	"github.com/stretchr/testify/mock"
)

type MockProductRepository struct{
	mock.Mock
}

func (m *MockProductRepository) CreateProduct(ctx context.Context, p *models.Product)error{
	return m.Called(ctx, p).Error(0)
}

func (m *MockProductRepository) GetProductByID(ctx context.Context, id string)(*models.Product, error){
	args := m.Called(ctx, id)
	if args.Get(0) != nil{
		return args.Get(0).(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) UpdateProduct(ctx context.Context, p *models.Product) error{
	return m.Called(ctx, p).Error(0)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, id string, stock int) error{
	return m.Called(ctx, id, stock).Error(0)
}

func (m *MockProductRepository) UpdateProductStatus(ctx context.Context, id string, isActive bool)error{
	return m.Called(ctx,id,isActive).Error(0)
}

func (m *MockProductRepository) GetProducts(ctx context.Context, limit, offset int, search, category string) ([]models.Product, error){
	args := m.Called(ctx, limit, offset, search, category)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) CountProducts(ctx context.Context, search, category string) (int64, error){
	args := m.Called(ctx,search, category)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductRepository) ReserveStock(ctx context.Context, items []models.ReserveItem)error{
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockProductRepository) ReleaseStock(ctx context.Context, items []models.ReserveItem) error{
	args := m.Called(ctx, items)
	return args.Error(0)
}