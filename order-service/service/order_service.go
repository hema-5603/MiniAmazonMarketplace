package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"order-service/client"
	"order-service/models"
	"order-service/repository"

	"github.com/google/uuid"
)

type OrderService interface{
	CreateOrder(ctx context.Context, userID string, req models.CheckoutRequest) (*models.Order, error)
}

type orderService struct{
	repo repository.OrderRepository
	productClient client.ProductClient
}

func NewOrderService(repo repository.OrderRepository, productClient client.ProductClient) OrderService{
	return &orderService{
		repo: repo,
		productClient: productClient,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, req models.CheckoutRequest) (*models.Order, error){
	if len(req.Items) == 0 {
		return nil, errors.New("Couldn't create an order with an empty cart")
	}

	// 1. Calling the product service to validate the cart
	validationResp, err := s.productClient.ValidateCart(ctx, req.Items)
	if err != nil{
		return nil, err
	}

	// 2. Hard block if any items is out of stock or deactivated
	if !validationResp.AllAvailable{
		return nil, errors.New("One or more items in your cart are out of stock or unavailable")
	}

	// 3. Create a map of the TRUE prices and seller IDs from the product service
	truthMap :=  make(map[string]struct{
		Price float64
		SellerID string
	})
	for _, item := range validationResp.Data{
		truthMap[item.ProductID] = struct{
			Price float64 
			SellerID string
		}{item.Price, item.SellerID}
	}

	// 2. Initialize the order
	orderID := uuid.New().String()
	var totalAmount float64
	var orderItems []models.OrderItem
	now := time.Now()

	// 2. Process each item and calculate the total amount
	for _, reqItem := range req.Items{
		trueData, exists := truthMap[reqItem.ProductID]
		if !exists{
			return nil, errors.New("Mismatch in product validation")
		}

		lineItemTotal := trueData.Price * float64(reqItem.Quantity)
		totalAmount += lineItemTotal

		orderItems = append(orderItems, models.OrderItem{
			ID: uuid.New().String(),
			OrderID: orderID,
			ProductID: reqItem.ProductID,
			SellerID: trueData.SellerID,
			Quantity: reqItem.Quantity,
			Price: trueData.Price,
		})
	}

	// 3. Build the final order object
	order := &models.Order{
		ID: orderID,
		UserID: userID,
		TotalAmount: totalAmount,
		Status: models.StatusPending, // Always starts as pending
		Items: orderItems,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Reserve the inventory
	err = s.productClient.ReserveStock(ctx, req.Items)
	if err != nil{
		// The product service blocked it
		return nil, errors.New("Checkout failed during inventory reservation: " + err.Error())
	}
	// 4. Call the repository which triggers the SQL transaction
	err = s.repo.CreateOrder(ctx, order)
	if err != nil{
		slog.Error("Failed to save order after reserving stock", slog.String("order_id",orderID))
		return nil , err
	}

	return order, nil
}