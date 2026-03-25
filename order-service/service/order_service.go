package service

import (
	"context"
	"errors"
	"time"

	"order-service/models"
	"order-service/repository"

	"github.com/google/uuid"
)

type OrderService interface{
	CreateOrder(ctx context.Context, userID string, req models.CheckoutRequest) (*models.Order, error)
}

type orderService struct{
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService{
	return &orderService{repo: repo}
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, req models.CheckoutRequest) (*models.Order, error){
	if len(req.Items) == 0 {
		return nil, errors.New("Couldn't create an order with an empty cart")
	}

	// 1. Initialize the order
	orderID := uuid.New().String()
	var totalAmount float64
	var orderItems []models.OrderItem
	now := time.Now()

	// 2. Process each item and calculate the total amount
	for _, item := range req.Items{
		if item.Quantity <= 0{
			return nil, errors.New("Item's quantity must be greater than zero")
		}

		lineItemTotal := item.Price * float64(item.Quantity)
		totalAmount += lineItemTotal

		orderItems = append(orderItems, models.OrderItem{
			ID: uuid.New().String(),
			OrderID: orderID,
			ProductID: item.ProductID,
			SellerID: item.SellerID,
			Quantity: item.Quantity,
			Price: item.Price,
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

	// 4. Call the repository which triggers the SQL transaction
	err := s.repo.CreateOrder(ctx, order)
	if err != nil{
		return nil , err
	}

	return order, nil
}