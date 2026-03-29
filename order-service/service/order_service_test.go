package service

import (
	"context"
	"errors"
	"testing"

	"order-service/client"
	"order-service/mocks"
	"order-service/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderService(t *testing.T){
	ctx := context.Background()

	t.Run("CreateOrder - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		mockClient := new(mocks.MockProductClient)
		svc := NewOrderService(mockRepo, mockClient)

		req := models.CheckoutRequest{
			Items: []models.CheckoutItem{
				{
					ProductID: "prod1",
					Quantity: 2,
					Price: 200,
				},
			},
		}

		// 1. Mock validate cart
		mockClient.On("ValidateCart", ctx, req.Items).Return(&client.ProductValidationResponse{
			Success: true,
			AllAvailable: true,
			Data: []struct{
				ProductID string `json:"product_id"`
				HasStock bool `json:"has_stock"`
				Price float64 `json:"price"`
				SellerID string `json:"seller_id"`
				Message string `json:"message"`
			}{
				{
					ProductID : "prod1", 
					HasStock : true, 
					Price : 200.0, 
					SellerID: "seller1",
				},
			},
		},nil).Once()


		// 2. Mock stock reservation
		mockClient.On("ReserveStock", ctx, req.Items).Return(nil).Once()

		// 3. Mock DB Insertion
		mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()

		order, err := svc.CreateOrder(ctx, "user1", req)

		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, 400.0, order.TotalAmount)
		assert.Equal(t, models.StatusPending, order.Status)
		mockRepo.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("CreateOrder - Failure (Price mismatch)", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		mockClient := new(mocks.MockProductClient)
		svc := NewOrderService(mockRepo, mockClient)

		req := models.CheckoutRequest{
			Items: []models.CheckoutItem{
				{
					ProductID: "prod1",
					Quantity: 2,
					Price: 100,
				},
			},
		}

		// 1. Mock validate cart
		mockClient.On("ValidateCart", ctx, req.Items).Return(&client.ProductValidationResponse{
			Success: true,
			AllAvailable: true,
			Data: []struct{
				ProductID string `json:"product_id"`
				HasStock bool `json:"has_stock"`
				Price float64 `json:"price"`
				SellerID string `json:"seller_id"`
				Message string `json:"message"`
			}{
				{
					ProductID : "prod1", 
					HasStock : true, 
					Price : 200, 
					SellerID: "seller1",
				},
			},
		},nil).Once()

		order, err := svc.CreateOrder(ctx, "user1", req)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t,"The Price of one or more items has changed.", err.Error())
		mockClient.AssertNotCalled(t,"ReserveStock")
	})

	t.Run("GetOrderDetail - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		svc := NewOrderService(mockRepo, nil)

		existingOrder := &models.Order{
			ID: "order1",
			UserID: "user1",
		}

		mockRepo.On("GetOrderByID", ctx, "order1").Return(existingOrder, nil).Once()

		order, err := svc.GetOrderDetail(ctx, "order1", "user1")

		assert.NoError(t, err)
		assert.Equal(t, "order1", order.ID)
		mockRepo.AssertExpectations(t)
	})
	t.Run("GetOrderDetail - Failure (Order not found)", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		svc := NewOrderService(mockRepo, nil)


		mockRepo.On("GetOrderByID", ctx, "order101").Return(nil, errors.New("Order not found")).Once()

		order, err := svc.GetOrderDetail(ctx, "order101", "user1")

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "Order not found", err.Error())
	})
	t.Run("GetOrderDetail - Failure (Unauthorized ownership)", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		svc := NewOrderService(mockRepo, nil)

		mockRepo.On("GetOrderByID", ctx, "order101").Return(nil, errors.New("Order not found")).Once()

		order, err := svc.GetOrderDetail(ctx, "order101", "user1")

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "Order not found", err.Error())
	})

	t.Run("GetOrderHistory - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		svc := NewOrderService(mockRepo, nil)

		mockOrders := []models.Order{{
			ID: "order1",
			UserID: "user1",
		}}

		mockRepo.On("CountOrdersByUserID", ctx, "user1", "").Return(int64(1), nil).Once()
		mockRepo.On("GetOrdersByUserID", ctx, "user1", "", 10, 0).Return(mockOrders, nil).Once()

		res, err := svc.GetOrderHistory(ctx, "user1", 1, 10, "")

		assert.NoError(t, err)
		assert.Equal(t, 1, len(res.Data))
		assert.Equal(t, 1, res.Meta.TotalItems)
		assert.Equal(t, 1, res.Meta.CurrentPage)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetOrderHistory - Failure (Database error)", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		svc := NewOrderService(mockRepo, nil)

		mockRepo.On("CountOrdersByUserID", ctx, "user1", "").Return(int64(0), errors.New("Database error")).Once()

		res, err := svc.GetOrderHistory(ctx, "user1", 1, 10, "")

		assert.Error(t, err)
		assert.Nil(t,res)
		assert.Equal(t, "Failed to retrieve order history", err.Error())
	})

	t.Run("ExpireUnpaidOrders - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		mockClient := new(mocks.MockProductClient)
		svc := NewOrderService(mockRepo, mockClient)

		expiredIDs := []string{"order1"}
		expiredOrder := &models.Order{
			ID: "order1",
			Items: []models.OrderItem{{
				ProductID: "prod1",
				Quantity: 1,
			}},
		}

		mockRepo.On("GetPendingOrdersOlderThan", ctx, mock.AnythingOfType("time.Time")).Return(expiredIDs, nil).Once()
		mockRepo.On("GetOrderByID", ctx, "order1").Return(expiredOrder, nil).Once()
		mockClient.On("ReleaseStock", ctx, expiredOrder.Items).Return(nil).Once()
		mockRepo.On("UpdateOrderStatus", ctx, "order1", models.StatusExpired).Return(nil).Once()

		err := svc.ExpireUnpaidOrders(ctx)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("ExpireUnpaidOrders - Success", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		svc := NewOrderService(mockRepo, nil)

		mockRepo.On("GetPendingOrdersOlderThan", ctx, mock.AnythingOfType("time.Time")).Return([]string{}, nil).Once()

		err := svc.ExpireUnpaidOrders(ctx)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})


	t.Run("CancelOrder - Failure(Unauthorized)", func(t *testing.T) {
		mockRepo := new(mocks.MockOrderRepository)
		mockClient := new(mocks.MockProductClient)
		svc := NewOrderService(mockRepo, mockClient)

		existingOrder := &models.Order{
			ID: "order1",
			UserID: "user1", // Order belongs to user1
			Status: models.StatusPending,
		}

		mockRepo.On("GetOrderByID", ctx, "order1").Return(existingOrder, nil).Once()

		// user2 tries to cancel it
		err := svc.CancelOrder(ctx, "order1", "user2")

		assert.Error(t, err)
		assert.Equal(t, "Unauthorized: You do not own this product", err.Error())
		mockClient.AssertNotCalled(t, "ReleaseStock")
	})
}