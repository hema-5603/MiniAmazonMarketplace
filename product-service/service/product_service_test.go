package service

import (
	"context"
	"errors"
	"testing"
	
	"product-service/mocks"
	"product-service/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductService(t *testing.T){
	ctx := context.Background()

	t.Run("CreateProduct - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.CreateProductRequest{Name: "Basketball", Price: 1000}
		mockRepo.On("CreateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()

		prod, err := svc.CreateProduct(ctx, "seller1", req)

		assert.NoError(t, err)
		assert.Equal(t, "Basketball", prod.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateProduct - Failure(Resource Ownership)", func(t *testing.T) { // 403
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		//Product belongs to seller1
		existingProd := &models.Product{ID: "prod1", SellerID: "seller1", Name: "Basketball"}
		mockRepo.On("GetProductByID", ctx, "prod1").Return(existingProd, nil).Once()

		//seller2 tries to update it
		req := models.UpdateProductRequest{Name:"Football", Price: 2000}
		prod, err := svc.UpdateProduct(ctx, "prod1", "seller2", req)

		assert.Error(t, err)
		assert.Nil(t, prod)
		assert.Equal(t, "Unauthorized: You do not own this product", err.Error())

		// Assert the repo's update method was never called to save the data
		mockRepo.AssertNotCalled(t, "UpdateProduct")
	})

	t.Run("UpdateStock - Failure(Negative stock)", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.UpdateProductStockRequest{Stock: -4}
		err := svc.UpdateStock(ctx, "prod1", "seller1", req)

		assert.Error(t, err)
		assert.Equal(t, "Invalid operation: Stock cannot be negative", err.Error())
	})

	t.Run("GetProductDetail - Failure(Deactivated product)", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		//Product exists but it's deactivated
		inactiveProd := &models.Product{ID: "prod1", Name: "Basketball", IsActive: false}
		mockRepo.On("GetProductByID", ctx, "prod1").Return(inactiveProd, nil).Once()

		prod, err := svc.GetProductDetail(ctx, "prod1")

		assert.Error(t, err)
		assert.Nil(t, prod)
		assert.Equal(t, "Product not found", err.Error())
	})
	t.Run("UpdateProductStatus - Unauthorized attempt (Resource ownership)", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		//Product belongs to seller1
		dbProd := &models.Product{ID: "prod1", SellerID: "seller1", Name: "Basketball", IsActive: true}
		mockRepo.On("GetProductByID", ctx, "prod1").Return(dbProd, nil).Once()

		//seller2 tries to update it
		req := models.UpdateStatusRequest{IsActive: false}
		err := svc.UpdateProductStatus(ctx, "prod1", "seller2", req)

		// Assert it fails with the correct ownership error
		assert.Error(t, err)
		assert.Equal(t, "Unauthorized: You do not own this product", err.Error())

		// Assert the repo's update status method was never called to save the data
		mockRepo.AssertNotCalled(t, "UpdateProductStatus")
	})
	t.Run("ValidateStock - Success & Failure", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.StockCheckRequest{
			Items: []models.StockCheckItem{
				{ProductID: "valid_prod", RequestedQuantity: 2},
				{ProductID: "low_stock_prod", RequestedQuantity: 100},
			},
		}
		mockRepo.On("GetProductByID", ctx, "valid_prod").Return(&models.Product{ID: "valid_prod", Stock: 5, IsActive: true}, nil).Once()
		mockRepo.On("GetProductByID", ctx, "low_stock_prod").Return(&models.Product{ID: "low_stock_prod", Stock: 50, IsActive: true}, nil).Once()

		results, allAvailable, err := svc.ValidateStock(ctx, req)

		assert.NoError(t, err)
		assert.False(t, allAvailable)
		assert.Equal(t, 2, len(results))
		assert.True(t, results[0].HasStock)
		assert.False(t, results[1].HasStock)
		assert.Equal(t, "Insufficient stock", results[1].Message)
	})

	t.Run("ValidateStock - Failure (Deactivated product)", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.StockCheckRequest{
			Items: []models.StockCheckItem{
				{ProductID: "deactivated_prod", RequestedQuantity: 2},
			},
		}

		// Mock the database as returning a product where IsActive is FALSE

		inactivateProd := &models.Product{ID: "deactivated_prod", Stock: 50, IsActive: false}
		mockRepo.On("GetProductByID", ctx, "deactivated_prod").Return(inactivateProd, nil).Once()

		results, allAvailable, err := svc.ValidateStock(ctx, req)

		assert.NoError(t, err) 
		assert.False(t, allAvailable) // Stock is rejected
		assert.Equal(t, 1, len(results))
		assert.False(t, results[0].HasStock)
		assert.Equal(t, "Product is no longer available", results[0].Message)
	})

	// Reserve stock tests
	t.Run("Reserve Stock - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.ReserveStockRequest{
			Items: []models.ReserveItem{
				{ProductID: "prod1", Quantity: 2},
			},
		}

		// Expect the repo to be called with the items and return no error
		mockRepo.On("ReserveStock", ctx, req.Items).Return(nil).Once()

		err := svc.ReserveStock(ctx, req)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ReserveStock - Failure (Insufficient stock)", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.ReserveStockRequest{
			Items: []models.ReserveItem{
				{ProductID: "prod1", Quantity: 50000},
			},
		}

		// Simulate the database rejecting the reservation
		mockRepo.On("ReserveStock", ctx, req.Items).Return(errors.New("Insufficient stock")).Once()

		err := svc.ReserveStock(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, "Insufficient stock", err.Error())
		mockRepo.AssertExpectations(t)
	})

	// Release stock tests
	t.Run("ReleaseStock - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.ReserveStockRequest{
			Items: []models.ReserveItem{
				{ProductID: "prod1", Quantity: 2},
			},
		}

		// Expect the repo to be called to put the stock back
		mockRepo.On("ReleaseStock", ctx, req.Items).Return(nil).Once()

		err := svc.ReleaseStock(ctx, req)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ReleaseStock - Failure (Early return)",func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)


		// A Request comes in, but there are no items to release
		req := models.ReserveStockRequest{
			Items: []models.ReserveItem{},
		}
		err := svc.ReleaseStock(ctx, req)

		// It should return nil, but it should never call the database
		assert.NoError(t, err)
		mockRepo.AssertNotCalled(t, "ReleaseStock")
	})

	t.Run("ReleaseStock - Failure (Database crash)",func(t *testing.T) {
		mockRepo := new(mocks.MockProductRepository)
		svc := NewProductService(mockRepo)

		req := models.ReserveStockRequest{
			Items: []models.ReserveItem{
				{ProductID: "prod1", Quantity: 2},
			},
		}

		// Simulate the database failure while trying to put stock back
		mockRepo.On("ReleaseStock", ctx, req.Items).Return(errors.New("Database connection lost")).Once()

		err := svc.ReleaseStock(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, "Database connection lost", err.Error())
		mockRepo.AssertExpectations(t)
	})
}


